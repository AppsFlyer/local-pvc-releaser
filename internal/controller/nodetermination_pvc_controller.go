/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"

	"github.com/AppsFlyer/local-pvc-releaser/internal/exporters"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	RemovingNode            = "RemovingNode"
	NodeControllerComponent = "node-controller"
)

// NodeTerminationPVCReconciler reconciles a PersistentVolumeClaim object
type NodeTerminationPVCReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	Logger            *logr.Logger
	DryRun            bool
	Recorder          record.EventRecorder
	Collector         *exporters.Collector
	PvcSelector       bool
	PvcAnoCustomKey   string
	PvcAnoCustomValue string
}

// +kubebuilder:rbac:groups="",resources=events,verbs=list;get;create;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch

func (r *NodeTerminationPVCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodeTerminationEvent := &v1.Event{}
	if err := r.Get(ctx, req.NamespacedName, nodeTerminationEvent); err != nil {
		r.Logger.Error(err, "did not find the related NodeTermination event")

		return ctrl.Result{}, err
	}

	r.Logger.Info("node termination event found", "Message", nodeTerminationEvent.Message, "EventID", nodeTerminationEvent.UID, "EventTime", nodeTerminationEvent.LastTimestamp)

	pvList := &v1.PersistentVolumeList{}
	if err := r.List(ctx, pvList); err != nil {
		return ctrl.Result{}, err
	}

	allPvcList := &v1.PersistentVolumeClaimList{}
	if err := r.List(ctx, allPvcList); err != nil {
		return ctrl.Result{}, err
	}

	nodePvList := r.FilterPVListByNodeName(pvList, nodeTerminationEvent.InvolvedObject.Name)

	if len(nodePvList) == 0 {
		r.Logger.Info(fmt.Sprintf("could not find any bounded pv objects for node - %s. will not take any action", nodeTerminationEvent.InvolvedObject.Name))
		return ctrl.Result{}, nil
	}

	relatedPvcList := make([]*v1.PersistentVolumeClaim, 0)
	for _, pv := range nodePvList {
		if pvc := r.FilterPVCListByPV(allPvcList, pv); pvc != nil {
			r.Logger.Info(fmt.Sprintf("pvc - %s is bounded to pv - %s and marked for deletion", pvc.Name, pv.Name))
			relatedPvcList = append(relatedPvcList, pvc)
			continue
		}
		r.Logger.Info(fmt.Sprintf("could not find the pvc object for pv - %s. pvc handling will be skipped", pv.Name))
	}

	if err := r.CleanPVCS(ctx, relatedPvcList); err != nil {
		r.Logger.Error(err, "failed to delete pvc objects from kubernetes")
	}

	return ctrl.Result{}, nil
}

func (r *NodeTerminationPVCReconciler) CleanPVCS(ctx context.Context, pvcs []*v1.PersistentVolumeClaim) error {
	for _, pvc := range pvcs {

		if r.PvcSelector && pvc.Annotations[r.PvcAnoCustomKey] != r.PvcAnoCustomValue {
			r.Logger.Info(fmt.Sprintf("pvc - %s does not match the filtered key:value annotation of - %s:%s and will be skipped", pvc.Name, r.PvcAnoCustomKey, r.PvcAnoCustomValue))
			continue
		}

		err := r.Client.Delete(ctx, pvc)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to delete object - %s,", pvc.GetName()))
		}

		r.Recorder.Eventf(pvc, "Normal", "PVC-Released", "The PersistentVolumeClaim %s has been released", pvc.Name)
		r.Collector.DeletedPVC.With(prometheus.Labels{"dryrun": strconv.FormatBool(r.DryRun)}).Inc()

		r.Logger.Info(fmt.Sprintf("pvc object - %s was deleted successfully", pvc.GetName()))
	}

	return nil
}

func (r *NodeTerminationPVCReconciler) FilterPVListByNodeName(pvList *v1.PersistentVolumeList, nodeName string) []*v1.PersistentVolume {
	var relatedPVs []*v1.PersistentVolume

	for i := 0; i < len(pvList.Items); i++ {
		pv := &pvList.Items[i]
		// Ignoring PVs without affinity rules or PVs that already got released
		if pv.Spec.NodeAffinity != nil && pv.Spec.NodeAffinity.Required == nil || pv.Status.Phase != v1.VolumeBound {
			continue
		}

		for _, nst := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
			for _, matchEx := range nst.MatchExpressions {
				if containsString(matchEx.Values, nodeName) {
					r.Logger.Info(fmt.Sprintf("pv - %s is bounded to node - %s. will be marked for pvc cleanup", pv.Name, nodeName))
					relatedPVs = append(relatedPVs, pv)

					break
				}
			}
		}
	}

	return relatedPVs
}

func (r *NodeTerminationPVCReconciler) FilterPVCListByPV(pvcList *v1.PersistentVolumeClaimList, pv *v1.PersistentVolume) *v1.PersistentVolumeClaim {
	for i := 0; i < len(pvcList.Items); i++ {
		claim := &pvcList.Items[i]

		if claim.Spec.VolumeName == pv.Name {
			return claim
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeTerminationPVCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Event{}).WithEventFilter(onNodeTerminationEventCreatedPredicate()).
		Complete(r)
}

func onNodeTerminationEventCreatedPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			obj := e.Object.(*v1.Event)
			return obj.Reason == RemovingNode && obj.Source.Component == NodeControllerComponent
		},
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

func containsString(slice []string, target string) bool {
	for _, str := range slice {
		if str == target {
			return true
		}
	}
	return false
}

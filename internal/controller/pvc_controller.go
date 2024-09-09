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
	PVCnodeAnnotationKey    = "volume.kubernetes.io/selected-node"
)

// PVCReconciler reconciles a PersistentVolumeClaim object
type PVCReconciler struct {
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

func (r *PVCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodeTerminationEvent := &v1.Event{}
	if err := r.Get(ctx, req.NamespacedName, nodeTerminationEvent); err != nil {
		r.Logger.Error(err, "did not find the related NodeTermination event")

		return ctrl.Result{}, err
	}

	r.Logger.Info("node termination event found", "Message", nodeTerminationEvent.Message, "EventID", nodeTerminationEvent.UID, "EventTime", nodeTerminationEvent.LastTimestamp)

	terminatedNodeName := nodeTerminationEvent.InvolvedObject.Name

	pvcList := &v1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList); err != nil {
		return ctrl.Result{}, err
	}

	// Filtering the related PVC objects bounded to the terminated node
	nodePvcList := r.FilterPVCListByNodeName(pvcList, terminatedNodeName)

	if len(nodePvcList) == 0 {
		r.Logger.Info(fmt.Sprintf("could not find any bounded pvc objects for node - %s. will not take any action", terminatedNodeName))
		return ctrl.Result{}, nil
	}

	pvcListPendingDeletion := make([]*v1.PersistentVolumeClaim, 0)
	for _, nodePvc := range nodePvcList {

		err, isLocal := r.CheckLocalPvStoragePluginByPVC(ctx, nodePvc)
		if err != nil {
			return ctrl.Result{}, err
		}

		if isLocal {
			r.Logger.Info(fmt.Sprintf("pvc - %s is bounded to a pv with local storage on the terminated node and will be marked for deletion", nodePvc.Name))
			pvcListPendingDeletion = append(pvcListPendingDeletion, nodePvc)
		}
	}

	if err := r.CleanPVCS(ctx, pvcListPendingDeletion); err != nil {
		r.Logger.Error(err, "failed to delete pvc objects from kubernetes")
	}

	return ctrl.Result{}, nil
}

func (r *PVCReconciler) CleanPVCS(ctx context.Context, pvcs []*v1.PersistentVolumeClaim) error {
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

func (r *PVCReconciler) FilterPVCListByNodeName(pvcList *v1.PersistentVolumeClaimList, nodeName string) []*v1.PersistentVolumeClaim {
	var relatedPVCs []*v1.PersistentVolumeClaim

	for i := 0; i < len(pvcList.Items); i++ {
		pvc := &pvcList.Items[i]

		if pvcNode, exists := pvc.Annotations[PVCnodeAnnotationKey]; exists && pvcNode == nodeName {
			r.Logger.Info(fmt.Sprintf("pvc - %s is bounded to node - %s. will be marked for pv 'local' plugin scan.", pvc.Name, nodeName))
			relatedPVCs = append(relatedPVCs, pvc)
		}
	}

	return relatedPVCs
}

func (r *PVCReconciler) CheckLocalPvStoragePluginByPVC(ctx context.Context, pvc *v1.PersistentVolumeClaim) (error, bool) {
	pv := &v1.PersistentVolume{}
	pvKey := client.ObjectKey{Name: pvc.Spec.VolumeName}

	if err := r.Get(ctx, pvKey, pv); err != nil {
		r.Logger.Error(err, fmt.Sprintf("could not find the attached pv object - %s", pvc.Spec.VolumeName))
		return err, false
	}

	if pv.Spec.Local != nil {
		return nil, true
	}

	return nil, false
}

// SetupWithManager sets up the controller with the Manager.
func (r *PVCReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

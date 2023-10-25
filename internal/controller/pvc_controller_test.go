package controller

import (
	"fmt"
	"github.com/AppsFlyer/local-pvc-releaser/internal/exporters"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func getReconciler() (*PVCReconciler, error) {
	fakeClient, err := getFakeClient()
	if err != nil {
		fmt.Println("unable to create fake client. exiting..")
		return nil, err
	}
	logr := klogr.New()

	collector := exporters.NewCollector()
	reconciler := &PVCReconciler{
		Client:            fakeClient,
		Scheme:            fakeClient.Scheme(),
		Logger:            &logr,
		DryRun:            false,
		Collector:         collector,
		PvcSelector:       true,
		PvcAnoCustomKey:   "appsflyer.com/local-pvc-releaser",
		PvcAnoCustomValue: "enabled",
	}

	return reconciler, nil
}

func getFakeClient(initObjs ...client.Object) (client.WithWatch, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build(), nil
}

func TestFilterPVCListByPV(t *testing.T) {
	reconciler, err := getReconciler()
	if err != nil {
		t.Error("Failed to initialize client.")
	}

	// Create a sample PVC list
	pvcList := &v1.PersistentVolumeClaimList{
		Items: []v1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pvc1"},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "pv1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pvc2"},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "pv2",
				},
			},
		},
	}

	// Create a sample PV
	pvExist := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "pv1"},
	}

	pvNil := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "pv3"},
	}

	// Simulate successful filtering
	result := reconciler.FilterPVCListByPV(pvcList, pvExist)

	// Check if the result is as expected
	if result == nil {
		t.Error("Expected a non-nil result, got nil")
	}

	if result.Name != "pvc1" {
		t.Errorf("Expected result to be pvc1, got %s", result.Name)
	}

	// Simulate non existed pv
	result = reconciler.FilterPVCListByPV(pvcList, pvNil)

	// Check if the result is as expected
	if result != nil {
		t.Error("Expected nil result, got non-nil object as the filter return matched value")
	}
}

func TestFilterPVListByNodeName(t *testing.T) {
	reconciler, err := getReconciler()
	if err != nil {
		t.Error("Failed to initialize client.")
	}

	// Create a sample PV list
	pvList := &v1.PersistentVolumeList{
		Items: []v1.PersistentVolume{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pv1"},
				Spec: v1.PersistentVolumeSpec{
					NodeAffinity: &v1.VolumeNodeAffinity{
						Required: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/hostname",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"node1"},
										},
									},
								},
							},
						},
					},
				},
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeBound,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pv2"},
				Spec: v1.PersistentVolumeSpec{
					NodeAffinity: &v1.VolumeNodeAffinity{
						Required: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/hostname",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"node2"},
										},
									},
								},
							},
						},
					},
				},
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeBound,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pv3"},
				Spec: v1.PersistentVolumeSpec{
					NodeAffinity: &v1.VolumeNodeAffinity{
						Required: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "kubernetes.io/hostname",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"node3"},
										},
									},
								},
							},
						},
					},
				},
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeBound,
				},
			},
			// Add more PVs as needed for additional test cases
		},
	}

	// Call the function for matched results
	nodeName := "node2"
	result := reconciler.FilterPVListByNodeName(pvList, nodeName)

	// Check if the result is as expected
	if len(result) != 1 {
		t.Errorf("Expected 1 related PV, got %d", len(result))
	}

	if result[0].Name != "pv2" {
		t.Errorf("Expected result to be pv2, got %s", result[0].Name)
	}

	// Call the function with empty PV list
	nodeNameNil := ""
	result = reconciler.FilterPVListByNodeName(&v1.PersistentVolumeList{}, nodeNameNil)

	// Check if the result is as expected
	if len(result) != 0 {
		t.Errorf("Expected 0 related PV, got %d", len(result))
	}
}

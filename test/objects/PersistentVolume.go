package objects

import (
	"context"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PV interface {
	Create(pvName, nodeName, storageClassName string) *corev1.PersistentVolume
	DeleteAll(ctx context.Context, client client.Client)
}

type persistentVolume struct{}

func NewPV() PV {
	return &persistentVolume{}
}

func (_ *persistentVolume) Create(pvName, nodeName, storageClassName string) *corev1.PersistentVolume {
	fsType := "ntfs"
	pv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              pvName,
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			StorageClassName:              storageClassName,
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path:   "/test",
					FSType: &fsType,
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{nodeName},
								},
							},
						},
					},
				},
			},
		},
	}

	return pv
}
func (_ *persistentVolume) DeleteAll(ctx context.Context, client client.Client) {
	pvList := &corev1.PersistentVolumeList{}
	gomega.Expect(client.List(ctx, pvList)).To(gomega.Succeed())

	for _, pv := range pvList.Items {
		gomega.Expect(client.Delete(ctx, &pv)).To(gomega.Succeed())
	}
}

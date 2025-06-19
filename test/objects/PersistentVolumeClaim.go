package objects

import (
	"context"

	"github.com/go-openapi/swag"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PVC interface {
	Create(name, pvName string, storageClassName string, annotations map[string]string) *corev1.PersistentVolumeClaim
	DeleteAll(ctx context.Context, client client.Client)
	RemoveProtectionFinalizer(ctx context.Context, client client.Client, pvc *corev1.PersistentVolumeClaim, finalizerName string) error
}

type persistentVolumeClaim struct {
}

func NewPVC() PVC {
	return &persistentVolumeClaim{}
}

func (persistentVolumeClaim) Create(name, pvName string, storageClassName string, annotations map[string]string) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
			Annotations:       annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			StorageClassName: &storageClassName,
			VolumeName:       pvName,
		},
	}

	return pvc
}

func (persistentVolumeClaim) RemoveProtectionFinalizer(ctx context.Context, client client.Client, pvc *corev1.PersistentVolumeClaim, finalizerName string) error {
	if swag.ContainsStrings(pvc.Finalizers, finalizerName) {
		pvc.Finalizers = removeString(pvc.Finalizers, finalizerName)
		gomega.Expect(client.Update(ctx, pvc)).To(gomega.Succeed())
	}

	return nil
}

func (claim persistentVolumeClaim) DeleteAll(ctx context.Context, client client.Client) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	gomega.Expect(client.List(ctx, pvcList)).To(gomega.Succeed())

	for _, pvc := range pvcList.Items {
		gomega.Expect(claim.RemoveProtectionFinalizer(ctx, client, &pvc, "kubernetes.io/pvc-protection")).To(gomega.Succeed())
		gomega.Expect(client.Delete(ctx, &pvc)).To(gomega.Succeed())
	}
}

func removeString(slice []string, target string) []string {
	result := []string{}
	for _, str := range slice {
		if str != target {
			result = append(result, str)
		}
	}
	return result
}

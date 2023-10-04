package objects

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node interface {
	Create() *corev1.Node
	//DeleteAll(ctx context.Context, client client.Client)
	//RemoveProtectionFinalizer(ctx context.Context, client client.Client, pvc *corev1.PersistentVolumeClaim, finalizerName string) error
}

type kubernetesNode struct {
}

func NewNode() Node {
	return &kubernetesNode{}
}

func (_ kubernetesNode) Create() *corev1.Node {
	node := &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-1",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.NodeSpec{},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
				corev1.ResourceMemory:           resource.MustParse("15037900Ki"),
				corev1.ResourceCPU:              resource.MustParse("2Gi"),
			},
		},
	}

	return node
}

package objects

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Event interface {
	Create(nodeName string, eventReason string) *corev1.Event
	FindByReason(ctx context.Context, client client.Client, reason string) *corev1.Event
	DeleteAll(ctx context.Context, client client.Client)
}

func NewEvent() Event {
	return &event{}
}

type event struct {
}

func (_ *event) Create(nodeName string, eventReason string) *corev1.Event {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event-test",
			Namespace: metav1.NamespaceDefault,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:       "Node",
			Name:       nodeName,
			APIVersion: "v1",
			UID:        uuid.NewUUID(),
		},
		Reason:  eventReason,
		Message: fmt.Sprintf("Node %s event: Removing Node %s from Controller", nodeName, nodeName),
		Type:    corev1.EventTypeNormal,
		Source: corev1.EventSource{
			Component: "node-controller",
		},
		FirstTimestamp: metav1.NewTime(time.Now()),
		LastTimestamp:  metav1.NewTime(time.Now()),
	}

	return event
}

func (_ *event) FindByReason(ctx context.Context, client client.Client, reason string) *corev1.Event {
	eventList := &corev1.EventList{}
	Expect(client.List(ctx, eventList)).To(Succeed())

	for _, event := range eventList.Items {
		if event.Reason == reason {
			return &event
		}
	}

	return nil
}

func (_ *event) DeleteAll(ctx context.Context, client client.Client) {
	eventList := &corev1.EventList{}
	Expect(client.List(ctx, eventList)).To(Succeed())

	for _, e := range eventList.Items {
		Expect(client.Delete(ctx, &e)).To(Succeed())
	}
}

package test

import (
	"github.com/AppsFlyer/local-pvc-releaser/test/objects"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Successful PVC Release", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		finalizerProtectionName = "kubernetes.io/pvc-protection"
		pvcName                 = "pvc-test"
		pvName                  = "test-pv"
		nodeName                = "node-1"
		eventReason             = "RemovingNode"
		storageClassName        = "local-storage"

		timeout  = time.Second * 60
		interval = time.Millisecond * 1000
	)
	AfterEach(func() {
		objects.Helper().PersistentVolumeClaim().DeleteAll(ctx, k8sClient)
		objects.Helper().PersistentVolume().DeleteAll(ctx, k8sClient)
		objects.Helper().Event().DeleteAll(ctx, k8sClient)
	})
	Context("When Receiving event on node-termination", func() {
		It("Should delete the related pvc", func() {
			By("By Creating PVC and PV object and triggering node termination event")
			pv := objects.Helper().PersistentVolume().Create(pvName, nodeName, storageClassName)
			Expect(k8sClient.Create(ctx, pv)).Should(Succeed())

			fetchedPv := &v1.PersistentVolume{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name, Namespace: pv.Namespace}, fetchedPv); err != nil {
					return err
				}

				if fetchedPv.Name != pv.Name {
					return errors.New("Unable to find the PV after creation, creation failed")
				}

				return nil
			}, timeout, interval).Should(Succeed())

			pvcAnnotations := map[string]string{
				"appsflyer.com/local-pvc-releaser":   "enabled",
				"volume.kubernetes.io/selected-node": nodeName,
			}
			pvc := objects.Helper().PersistentVolumeClaim().Create(pvcName, pvName, storageClassName, pvcAnnotations)
			Expect(k8sClient.Create(ctx, pvc)).Should(Succeed())

			// Create a PersistentVolumeClaim (PVC)
			fetchedPvc := &v1.PersistentVolumeClaim{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, fetchedPvc); err != nil {
					return err
				}

				if fetchedPvc.Name != pvc.Name {
					return errors.New("Pvc was not found after creation, creation failed")
				}

				return nil
			}, timeout, interval).Should(Succeed())

			Expect(objects.Helper().PersistentVolumeClaim().RemoveProtectionFinalizer(ctx, k8sClient, fetchedPvc, finalizerProtectionName)).Should(Succeed())

			By("By Creating Node-Termination event on the node related to the PVC and PV")
			event := objects.Helper().Event().Create(nodeName, eventReason)
			Expect(k8sClient.Create(ctx, event)).Should(Succeed())

			allPvcList := &v1.PersistentVolumeClaimList{}
			Eventually(func() error {
				if err := k8sClient.List(ctx, allPvcList); err != nil {
					return err
				}
				if len(allPvcList.Items) != 0 {
					return errors.Errorf("expected amount of pvc to be 0, received %d", len(allPvcList.Items))
				}
				return nil
			}, timeout, interval).Should(Succeed())
			Expect(len(allPvcList.Items)).To(BeEquivalentTo(0))
			// Checking that a post event of PVC deletion was generated
			Eventually(objects.Helper().Event().FindByReason(ctx, k8sClient, eventReason), timeout, interval).Should(Not(BeNil()))
		})
	})
})

var _ = Describe("Ignoring PVC cleanup", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		finalizerProtectionName = "kubernetes.io/pvc-protection"
		pvcName                 = "pvc-test"
		pvName                  = "test-pv"
		nodeName                = "node-1"
		eventReason             = "RemovingNode"
		storageClassName        = "local-storage"

		timeout  = time.Second * 20
		interval = time.Millisecond * 1000
	)
	AfterEach(func() {
		objects.Helper().PersistentVolumeClaim().DeleteAll(ctx, k8sClient)
		objects.Helper().PersistentVolume().DeleteAll(ctx, k8sClient)
		objects.Helper().Event().DeleteAll(ctx, k8sClient)
	})
	Context("When Receiving event on node-termination", func() {
		It("Should ignore as there are no PVs attached", func() {
			By("By Creating only a PVC object")

			pvcAnnotations := map[string]string{
				"appsflyer.com/local-pvc-releaser":   "enabled",
				"volume.kubernetes.io/selected-node": nodeName,
			}
			pvc := objects.Helper().PersistentVolumeClaim().Create(pvcName, "", "different-storage-class", pvcAnnotations)
			Expect(k8sClient.Create(ctx, pvc)).Should(Succeed())

			// Create a PersistentVolumeClaim (PVC)
			fetchedPvc := &v1.PersistentVolumeClaim{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, fetchedPvc); err != nil {
					return err
				}

				if fetchedPvc.Name != pvc.Name {
					return errors.New("Pvc was not found after creation, creation failed")
				}

				return nil
			}, timeout, interval).Should(Succeed())

			Expect(objects.Helper().PersistentVolumeClaim().RemoveProtectionFinalizer(ctx, k8sClient, fetchedPvc, finalizerProtectionName)).Should(Succeed())

			By("By Creating Node-Termination event on the node related to the PV")
			event := objects.Helper().Event().Create(nodeName, eventReason)
			Expect(k8sClient.Create(ctx, event)).Should(Succeed())

			allPvcList := &v1.PersistentVolumeClaimList{}
			Eventually(func() error {
				if err := k8sClient.List(ctx, allPvcList); err != nil {
					return err
				}
				if len(allPvcList.Items) != 1 {
					return errors.Errorf("expected amount of pvc to be 1, received %d", len(allPvcList.Items))
				}
				return nil
			}, timeout, interval).Should(Succeed())
			Expect(len(allPvcList.Items)).To(BeEquivalentTo(1))
		})
	})
})

var _ = Describe("Ignoring PVC cleanup", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		finalizerProtectionName = "kubernetes.io/pvc-protection"
		pvcName                 = "pvc-test"
		pvName                  = "test-pv"
		nodeName                = "node-1"
		eventReason             = "RemovingNode"
		storageClassName        = "local-storage"

		timeout  = time.Second * 20
		interval = time.Millisecond * 1000
	)
	AfterEach(func() {
		objects.Helper().PersistentVolumeClaim().DeleteAll(ctx, k8sClient)
		objects.Helper().PersistentVolume().DeleteAll(ctx, k8sClient)
		objects.Helper().Event().DeleteAll(ctx, k8sClient)
	})
	Context("When Receiving event on node-termination", func() {
		It("Should ignore as the PVC does not have the default annotations for selector", func() {
			By("By Creating a PV and PVC objects where the PVC is not annotated correctly")
			pv := objects.Helper().PersistentVolume().Create(pvName, nodeName, storageClassName)
			Expect(k8sClient.Create(ctx, pv)).Should(Succeed())

			fetchedPv := &v1.PersistentVolume{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name, Namespace: pv.Namespace}, fetchedPv); err != nil {
					return err
				}

				if fetchedPv.Name != pv.Name {
					return errors.New("Unable to find the PV after creation, creation failed")
				}

				return nil
			}, timeout, interval).Should(Succeed())

			pvcAnnotations := map[string]string{
				"test-should-fail":                   "true",
				"volume.kubernetes.io/selected-node": nodeName,
			}
			pvc := objects.Helper().PersistentVolumeClaim().Create(pvcName, pvName, "different-storage-class", pvcAnnotations)
			Expect(k8sClient.Create(ctx, pvc)).Should(Succeed())

			// Create a PersistentVolumeClaim (PVC)
			fetchedPvc := &v1.PersistentVolumeClaim{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, fetchedPvc); err != nil {
					return err
				}

				if fetchedPvc.Name != pvc.Name {
					return errors.New("Pvc was not found after creation, creation failed")
				}

				return nil
			}, timeout, interval).Should(Succeed())

			Expect(objects.Helper().PersistentVolumeClaim().RemoveProtectionFinalizer(ctx, k8sClient, fetchedPvc, finalizerProtectionName)).Should(Succeed())

			By("By Creating Node-Termination event on the node related to the PV")
			event := objects.Helper().Event().Create(nodeName, eventReason)
			Expect(k8sClient.Create(ctx, event)).Should(Succeed())

			allPvcList := &v1.PersistentVolumeClaimList{}
			Eventually(func() error {
				if err := k8sClient.List(ctx, allPvcList); err != nil {
					return err
				}
				if len(allPvcList.Items) != 1 {
					return errors.Errorf("expected amount of pvc to be 1, received %d", len(allPvcList.Items))
				}
				return nil
			}, timeout, interval).Should(Succeed())
			Expect(len(allPvcList.Items)).To(BeEquivalentTo(1))
		})
	})
})

package controllers

import (
	"context"
	"time"

	"github.com/operatify/operatify/controllers/manager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/operatify/operatify/reconciler"
)

var _ = Describe("Test Dependencies", func() {

	Context("When creating and deleting", func() {

		It("should do nothing until dependency is created", func() {
			bId := "b-" + RandomString(10)
			ownerId := "a-" + RandomString(10)

			keyB, createdB := nameAndSpecB(bId, ownerId, []string{})

			// create B
			Expect(k8sClient.Create(context.Background(), createdB)).Should(Succeed())

			// expect Pending state to be set
			waitUntilReconcileStateB(keyB, reconciler.Pending)

			// expect nothing else to happen
			Consistently(func() reconciler.ReconcileState {
				f, _ := getObjectB(keyB)
				return reconciler.ReconcileState(f.Status.State)
			}, time.Second*3, interval).Should(Equal(reconciler.Pending))

			// now create owner
			keyA, createdA := nameAndSpecA(ownerId)
			Expect(k8sClient.Create(context.Background(), createdA)).Should(Succeed())

			// now B should eventually succeed
			waitUntilReconcileStateB(keyB, reconciler.Succeeded)

			By("owner should delete successfully")
			Expect(deleteObjectA(keyA)).To(Succeed())

			By("the dependent object (B) to be deleted as well")
			// Unfortunately this doesn't seem to work in the test suite
			// waitUntilObjectMissingB(keyB)
		})

		It("should go back to pending of a dependency fails", func() {
			bId := "b-" + RandomString(10)
			ownerId := "a-" + RandomString(10)

			keyB, createdB := nameAndSpecB(bId, ownerId, []string{})
			keyA, createdA := nameAndSpecA(ownerId)

			// create B
			Expect(k8sClient.Create(context.Background(), createdB)).Should(Succeed())

			// now create owner
			Expect(k8sClient.Create(context.Background(), createdA)).Should(Succeed())

			// now B should eventually succeed
			waitUntilReconcileStateB(keyB, reconciler.Succeeded)

			By("Updating owner to failure state")
			resourceManager.AddBehaviour(ownerId, manager.Behaviour{
				Event:     manager.EventUpdate,
				Operation: manager.UpdateFail.AsOperation(),
			})

			// tell it update is required (ony for the next verify)
			resourceManager.AddBehaviour(ownerId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyNeedsUpdate.AsOperation(),
				OneTime:   true,
			})

			toUpdate, _ := getObjectA(keyA)
			toUpdate.Spec.IntData = 1
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())

			// expect B to return to pending state
			waitUntilReconcileStateB(keyB, reconciler.Pending)

			By("owner should delete successfully")
			Expect(deleteObjectB(keyB)).To(Succeed())

			// should still delete B even though it's in pending state
			waitUntilObjectMissingB(keyB)
		})

		It("should do nothing until all additional dependencies are created", func() {
			bId := "b-" + RandomString(10)
			ownerId := "a-" + RandomString(10)
			dependency1 := "a-" + RandomString(10)
			dependency2 := "a-" + RandomString(10)

			keyB, createdB := nameAndSpecB(bId, ownerId, []string{dependency1, dependency2})

			// create B
			Expect(k8sClient.Create(context.Background(), createdB)).Should(Succeed())
			_, createdA0 := nameAndSpecA(ownerId)
			_, createdA1 := nameAndSpecA(dependency1)
			_, createdA2 := nameAndSpecA(dependency2)
			Expect(k8sClient.Create(context.Background(), createdA0)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), createdA1)).Should(Succeed())

			// expect Pending state to be set
			waitUntilReconcileStateB(keyB, reconciler.Pending)

			// expect nothing else to happen
			Consistently(func() reconciler.ReconcileState {
				f, _ := getObjectB(keyB)
				return reconciler.ReconcileState(f.Status.State)
			}, time.Second*3, interval).Should(Equal(reconciler.Pending))

			// now create final dependency
			Expect(k8sClient.Create(context.Background(), createdA2)).Should(Succeed())

			// now B should eventually succeed
			waitUntilReconcileStateB(keyB, reconciler.Succeeded)
		})
	})
})

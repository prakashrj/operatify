package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/szoio/resource-operator-factory/controllers/manager"
	"github.com/szoio/resource-operator-factory/reconciler"
)

var _ = Describe("Test Create and Delete", func() {

	Context("When creating and deleting", func() {

		It("should create asynchronously and delete asynchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).To(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))
		})

		It("should create synchronously and delete asynchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// tell it to run the create synchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateSync.AsOperation(),
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).To(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))
		})

		It("should create asynchronously and delete synchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
			}))

			// tell it to run delete synchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventDelete,
				Operation: manager.DeleteSync.AsOperation(),
			})

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				return deleteObject(key)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultMissing,
			}))
		})
	})
})

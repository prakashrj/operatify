/*

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

package controllers

import (
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/operatify/operatify/controllers/b"

	"github.com/operatify/operatify/controllers/shared"

	"github.com/operatify/operatify/controllers/a"
	"github.com/operatify/operatify/controllers/manager"
	"github.com/operatify/operatify/reconciler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api "github.com/operatify/operatify/api/v1alpha1"
	testv1alpha1 "github.com/operatify/operatify/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

const timeout = time.Second * 5
const interval = time.Millisecond * 100

var accessPermissionAnnotation = shared.AnnotationBaseName + reconciler.AccessPermissionAnnotation

var resourceManager = manager.CreateManager()

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = api.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = testv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// Create test controllers
	err = (&a.ControllerFactory{
		ResourceManagerCreator: a.CreateResourceManager,
		Scheme:                 scheme.Scheme,
		Manager:                resourceManager,
	}).SetupWithManager(k8sManager, reconciler.ReconcileParameters{
		RequeueAfter: 100,
	}, nil)
	Expect(err).ToNot(HaveOccurred())

	err = (&b.ControllerFactory{
		ResourceManagerCreator: b.CreateResourceManager,
		Scheme:                 scheme.Scheme,
		Manager:                resourceManager,
	}).SetupWithManager(k8sManager, reconciler.ReconcileParameters{
		RequeueAfter:        100,
		RequeueAfterSuccess: 1000,
		RequeueAfterFailure: 1000,
	}, nil)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func randomStringWithCharset(length int, charset string) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

const charset = "abcdefghijklmnopqrstuvwxyz"

func RandomString(length int) string {
	return randomStringWithCharset(length, charset)
}

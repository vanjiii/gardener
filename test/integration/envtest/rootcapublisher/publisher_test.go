// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rootcapublisher_test

import (
	"time"

	. "github.com/gardener/gardener/pkg/utils/test/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Root CA Controller tests", func() {
	var (
		namespace *corev1.Namespace
		configMap *corev1.ConfigMap
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"},
		}

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-root-ca.crt",
				Namespace: namespace.Name,
			},
		}

		Expect(testClient.Create(ctx, namespace)).To(Or(Succeed(), BeAlreadyExistsError()))

		Eventually(func() error {
			return testClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
		}, time.Millisecond*500, time.Millisecond*10).Should(Succeed())
	})

	It("should successfully create a config map on creating a namespace", func() {})

	It("should keep the secret in the desired state after Delete/Update of the secret", func() {
		By("Deleting the secret")
		Expect(testClient.Delete(ctx, configMap)).To(Succeed())
		Eventually(func() error {
			return testClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
		}, time.Millisecond*300, time.Millisecond*10).Should(BeNotFoundError())

		Eventually(func() error {
			return testClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
		}, time.Millisecond*300, time.Millisecond*10).Should(Succeed())

		By("Updating the secret")
		configMap.Data = nil
		Expect(testClient.Update(ctx, configMap)).To(Succeed())

		Eventually(func() bool {
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap); err != nil {
				return false
			}

			return configMap.Data != nil
		}, time.Millisecond*300, time.Millisecond*10).Should(BeTrue())

		By("Ignoring annotating configmap")
		configMap.Data = nil
		configMap.Annotations = map[string]string{"kubernetes.io/description": "test description"}
		Expect(testClient.Update(ctx, configMap)).To(Succeed())

		Consistently(func() bool {
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap); err != nil {
				return false
			}
			return configMap.Data == nil
		}, time.Millisecond*300, time.Millisecond*10).Should(BeTrue())

		By("Ignoring configmap with different name")
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "secret",
				Namespace:   namespace.Name,
				Annotations: map[string]string{"foo": "bar"},
			},
		}
		Expect(testClient.Create(ctx, cm)).To(Succeed())

		baseCM := cm.DeepCopyObject().(client.Object)
		cm.Annotations["foo"] = "newbar"
		Expect(testClient.Patch(ctx, cm, client.MergeFrom(baseCM))).To(Succeed())

		Expect(cm.Annotations).To(HaveLen(1))
		Expect(cm.Annotations).To(HaveKeyWithValue("foo", "newbar"))
	})
})
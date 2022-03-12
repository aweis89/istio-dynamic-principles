/*
Copyright 2022 Aaron Weisberg.

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
	"context"
	"fmt"
	"time"

	v1 "github.com/aweis89/istio-dynamic-principles/api/v1"
	// . "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	timeout  = time.Second * 5        // nolint:gochecknoglobals
	interval = time.Millisecond * 250 // nolint:gochecknoglobals
)

var _ = Describe("DynamicAuthorizationPolicy controller", func() {
	labelSel := map[string]string{"labelKey": "labelVal"}

	dynamicPolicy := v1.DynamicPolicy{
		PodSelectors: labelSel,
		Name:         "policy",
		TrustDomain:  "cluster.local",
	}

	namespace := "default"

	tests := map[string]struct {
		dap  *v1.DynamicAuthorizationPolicy
		pods []*corev1.Pod
	}{
		"With multiple pods with same label selector": {
			&v1.DynamicAuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dap",
					Namespace: namespace,
				},
				Spec: v1.DynamicAuthorizationPolicySpec{
					DynamicPolicies: []v1.DynamicPolicy{
						dynamicPolicy,
					},
				},
			},
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-a",
						Namespace: namespace,
						Labels:    labelSel,
					},
					Spec: corev1.PodSpec{
						Containers:         []corev1.Container{{Image: "image", Name: "container"}},
						ServiceAccountName: "service-account",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-b",
						Namespace: namespace,
						Labels:    labelSel,
					},
					Spec: corev1.PodSpec{
						Containers:         []corev1.Container{{Image: "image", Name: "container"}},
						ServiceAccountName: "service-account-b",
					},
				},
			},
		},
	}

	for desc, test := range tests {
		matchingSelectorsTest(desc, test.dap, test.pods)
	}
})

func matchingSelectorsTest(desc string, dap *v1.DynamicAuthorizationPolicy, pods []*corev1.Pod) {
	Describe(desc, func() {
		It("Pod exists with mathching selectors, policies are added to DAP", func() {
			ctx := context.Background()

			Expect(k8sClient.Create(ctx, dap)).Should(Succeed())
			if !useFakeClient {
				// ensure dap gets triggered from later Pod change
				time.Sleep(time.Millisecond * 1000)
			}
			for _, pod := range pods {
				Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
			}

			dapNN := client.ObjectKeyFromObject(dap)

			if useFakeClient {
				dapr := DynamicAuthorizationPolicyReconciler{
					Scheme: scheme.Scheme,
					Client: k8sClient,
				}
				dapr.Reconcile(ctx, ctrl.Request{NamespacedName: dapNN})
			}

			Eventually(func(g Gomega) { // nolint:varnamelen
				createdDap := v1.DynamicAuthorizationPolicy{}
				err := k8sClient.Get(ctx, dapNN, &createdDap)
				g.Expect(err).NotTo(HaveOccurred())
				for _, policy := range dap.Spec.DynamicPolicies {
					name := policy.Name
					selectors := policy.PodSelectors
					for _, pod := range pods {
						if hasSelectors(pod, selectors) {
							podPolicy := fmt.Sprintf("%s/ns/%s/sa/%s",
								policy.TrustDomain, pod.GetNamespace(), pod.Spec.ServiceAccountName)
							policy := createdDap.Status.ServiceAccountPolicyMapping[name]
							g.Expect(policy).To(HaveKey(podPolicy))
						}
					}
				}
			}, timeout, interval).Should(Succeed())
		})
	})
}

func hasSelectors(pod *corev1.Pod, selectors map[string]string) bool {
	labels := pod.GetLabels()
	if labels == nil {
		return false
	}
	for k, v := range selectors {
		if val, ok := labels[k]; !ok || val != v {
			return false
		}
	}
	return true
}

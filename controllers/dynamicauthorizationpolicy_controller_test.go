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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250

	fakeClient = Label("fake")
	realClient = Label("real")
)

//var _ = Describe("DynamicAuthorizationPolicy controller", func() {
//	genTests(time.Duration(0), useFakeClient)
//})

// func genTests(waitBetweenCreate time.Duration, fakeClient bool) {
var _ = Describe("DynamicAuthorizationPolicy controller", func() {
	It("When Pod is exists mathching selectors, policies are added to DAP", func() {
		ctx := context.Background()

		name, namespace := "name", "default"
		labelSel := map[string]string{"labelKey": "labelVal"}

		dynamicPolicy := v1.DynamicPolicy{
			PodSelectors: labelSel,
			Name:         "policy",
			TrustDomain:  "cluster.local",
		}

		dap := &v1.DynamicAuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1.DynamicAuthorizationPolicySpec{
				DynamicPolicies: []v1.DynamicPolicy{
					dynamicPolicy,
				},
			},
		}

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labelSel,
			},
			Spec: corev1.PodSpec{
				Containers:         []corev1.Container{{Image: "image", Name: "container"}},
				ServiceAccountName: "service-account",
			},
		}

		Expect(k8sClient.Create(ctx, dap)).Should(Succeed())
		// tests if dap gets triggered from pod change
		if !useFakeClient {
			time.Sleep(time.Millisecond * 1000)
		}
		Expect(k8sClient.Create(ctx, pod)).Should(Succeed())

		dapNN := types.NamespacedName{Name: name, Namespace: namespace}

		if useFakeClient {
			dapr := DynamicAuthorizationPolicyReconciler{
				Scheme: scheme.Scheme,
				Client: k8sClient,
			}
			dapr.Reconcile(ctx, ctrl.Request{NamespacedName: dapNN})
		}

		Eventually(func(g Gomega) {
			createdDap := v1.DynamicAuthorizationPolicy{}
			err := k8sClient.Get(ctx, dapNN, &createdDap)
			g.Expect(err).NotTo(HaveOccurred())
			policy := createdDap.Status.ServiceAccountPolicyMapping[dynamicPolicy.Name]
			expectedPolicy := fmt.Sprintf("%s/ns/%s/sa/%s",
				dynamicPolicy.TrustDomain, pod.GetNamespace(), pod.Spec.ServiceAccountName)
			g.Expect(policy).To(HaveKey(expectedPolicy))
		}, timeout, interval).Should(Succeed())
	})
})

// })

// var _ = Describe("DynamicAuthorizationPolicy controller", func() {
// 	DescribeTable("When Pod is exists mathching selectors, policies are added to DAP",
// 		func(waitBetweenCreate time.Duration, triggerReconcile bool) {
// 			ctx := context.Background()
//
// 			name, namespace := "name", "default"
// 			labelSel := map[string]string{"labelKey": "labelVal"}
// 			dynamicPolicy := v1.DynamicPolicy{PodSelectors: labelSel, Name: "policy", TrustDomain: "cluster.local"}
//
// 			dap := &v1.DynamicAuthorizationPolicy{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      name,
// 					Namespace: namespace,
// 				},
// 				Spec: v1.DynamicAuthorizationPolicySpec{
// 					DynamicPolicies: []v1.DynamicPolicy{
// 						dynamicPolicy,
// 					},
// 				},
// 			}
//
// 			pod := &corev1.Pod{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      name,
// 					Namespace: namespace,
// 					Labels:    labelSel,
// 				},
// 				Spec: corev1.PodSpec{
// 					Containers:         []corev1.Container{{Image: "image", Name: "container"}},
// 					ServiceAccountName: "service-account",
// 				},
// 			}
//
// 			Expect(k8sClient.Create(ctx, dap)).Should(Succeed())
// 			time.Sleep(waitBetweenCreate)
// 			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
//
// 			dapNN := types.NamespacedName{Name: name, Namespace: namespace}
//
// 			if triggerReconcile {
// 				dapr := DynamicAuthorizationPolicyReconciler{
// 					Scheme: scheme.Scheme,
// 					Client: k8sClient,
// 				}
// 				dapr.Reconcile(ctx, ctrl.Request{NamespacedName: dapNN})
// 			}
//
// 			Eventually(func(g Gomega) {
// 				createdDap := v1.DynamicAuthorizationPolicy{}
// 				err := k8sClient.Get(ctx, dapNN, &createdDap)
// 				g.Expect(err).NotTo(HaveOccurred())
// 				policy := createdDap.Status.ServiceAccountPolicyMapping[dynamicPolicy.Name]
// 				expectedPolicy := fmt.Sprintf("%s/ns/%s/sa/%s",
// 					dynamicPolicy.TrustDomain, pod.GetNamespace(), pod.Spec.ServiceAccountName)
// 				g.Expect(policy).To(HaveKey(expectedPolicy))
// 			}, timeout, interval).Should(Succeed())
// 		},
// 		Entry("using a fake client", time.Duration(0), true, fakeClient),
// 		Entry("using a real client", time.Duration(0), false, realClient),
// 	)
// })

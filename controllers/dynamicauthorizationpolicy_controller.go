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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	istiov1alpha1 "github.com/aweis89/istio-abac-policy/api/v1alpha1"
	"github.com/pkg/errors"
	securityv1beta1 "istio.io/api/security/v1beta1"
)

// DynamicAuthorizationPolicyReconciler reconciles a DynamicAuthorizationPolicy object
type DynamicAuthorizationPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=istio.aweis.io,resources=dynamicauthorizationpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=istio.aweis.io,resources=dynamicauthorizationpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=istio.aweis.io,resources=dynamicauthorizationpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DynamicAuthorizationPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DynamicAuthorizationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	apr := &istiov1alpha1.DynamicAuthorizationPolicy{}
	err := r.Get(ctx, req.NamespacedName, apr)
	if err != nil {
		context := "unable to retrieve AuthorizationPolicyAbac"
		log.Error(err, context)
		return ctrl.Result{Requeue: true}, errors.Wrap(err, context)
	}
	dynamicPrinciples := apr.Spec.DynamicPrinciples
	for _, dp := range dynamicPrinciples {
		namespaceList := v1.NamespaceList{}
		for _, sel := range dp.NamespaceSelectors {
			for _, labels := range sel.MatchLabels {
				r.List(ctx, &namespaceList, client.MatchingLabels(labels))
			}
		}

		pods := v1.PodList{}
		for _, sel := range dp.PodSelectors {
			for _, labels := range sel.MatchLabels {
				r.List(ctx, &pods, client.MatchingLabels(labels), client.InNamespace(""))
			}
		}

		trustDomain := dp.TrustDomain
		if trustDomain == "" {
			trustDomain = "cluster.local"
		}

		principles := []string{}
		for _, pod := range pods.Items {
			sa := pod.Spec.ServiceAccountName
			namepsace := pod.ObjectMeta.Namespace
			principle := fmt.Sprintf("%s/ns/%s/sa/%s", trustDomain, namepsace, sa)
			principles = append(principles, principle)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DynamicAuthorizationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&istiov1alpha1.DynamicAuthorizationPolicy{}).
		Owns(&securityv1beta1.AuthorizationPolicy{}).
		Complete(r)
}

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

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	peerauthv1 "github.com/aweis89/istio-dynamic-principles/api/v1"
	v1 "github.com/aweis89/istio-dynamic-principles/api/v1"
	"github.com/pkg/errors"
)

// DynamicAuthorizationPolicyReconciler reconciles a DynamicAuthorizationPolicy object
type DynamicAuthorizationPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=peerauth.aweis.io,resources=dynamicauthorizationpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=peerauth.aweis.io,resources=dynamicauthorizationpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=peerauth.aweis.io,resources=dynamicauthorizationpolicies/finalizers,verbs=update

func (r *DynamicAuthorizationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	dap := peerauthv1.DynamicAuthorizationPolicy{}
	err := r.Get(ctx, req.NamespacedName, &dap)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("resource no longer available", "DynamicAuthorizationPolicy", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrapf(err,
			"unable to get DynamicAuthorizationPolicy %s", req.NamespacedName)
	}

	sapm := peerauthv1.ServiceAccountPolicyMapping{}
	for i := range dap.Spec.DynamicPolicies {
		policy := dap.Spec.DynamicPolicies[i]
		pods := corev1.PodList{}

		// copy pod selectors to DynamicAuthorizationPolicy labels for reverse triggering
		if dap.ObjectMeta.Labels == nil {
			dap.ObjectMeta.Labels = map[string]string{}
		}
		for k, v := range policy.PodSelectors {
			dap.ObjectMeta.Labels[k] = v
			log.Info("Adding DAP labels from PodSelectors", "key", k, "val", v)
		}

		err := r.List(ctx, &pods, client.MatchingLabels(policy.PodSelectors))
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err,
				"unable to list ServiceAccountPolicyMapping using labels %+v", policy.PodSelectors)
		}

		for _, pod := range pods.Items {
			log.Info("adding pod to DAP policies", "Pod", pod.GetName())
			sapm.Map(policy, pod)
		}
	}

	dap.Status.ServiceAccountPolicyMapping = sapm
	log.Info(fmt.Sprintf("%+v", dap))

	if err := r.Update(ctx, &dap); err != nil {
		return ctrl.Result{}, errors.Wrapf(err,
			"unable to update DynamicAuthorizationPolicy %s", req.NamespacedName)
	}

	// dap = peerauthv1.DynamicAuthorizationPolicy{}
	// err = r.Get(ctx, req.NamespacedName, &dap)
	// if err != nil {
	// 	if kerrors.IsNotFound(err) {
	// 		log.Info("resource no longer available", "DynamicAuthorizationPolicy", req.NamespacedName)
	// 		return ctrl.Result{}, nil
	// 	}
	// 	return ctrl.Result{}, errors.Wrapf(err,
	// 		"unable to get DynamicAuthorizationPolicy %s", req.NamespacedName)
	// }
	// dap.Status.ServiceAccountPolicyMapping = sapm
	// if err := r.Status().Update(ctx, &dap); err != nil {
	// 	return ctrl.Result{}, errors.Wrapf(err,
	// 		"unable to update DynamicAuthorizationPolicy %s", req.NamespacedName)
	// }

	return ctrl.Result{}, nil
}

func (r *DynamicAuthorizationPolicyReconciler) MapFunc(obj client.Object) []reconcile.Request {
	pod := obj.(*corev1.Pod)
	podLabels := pod.GetLabels()
	if podLabels == nil {
		return []reconcile.Request{}
	}
	log.Log.Info(fmt.Sprintf("%+v", podLabels))
	rr := []reconcile.Request{}
	for k, v := range podLabels {
		log.Log.Info("querying DAPS with labels", "labelKey", k, "labelVal", v)
		dapList := v1.DynamicAuthorizationPolicyList{}
		err := r.List(context.Background(), &dapList, client.MatchingLabels{k: v})
		if err != nil {
			log.Log.Error(err, "unable to list DynamicAuthorizationPolicy")
		}
		for _, dap := range dapList.Items {
			log.Log.Info("GOT DAPPPPPPPPPP", dap.Name, dap.Namespace)
			rr = append(rr,
				reconcile.Request{NamespacedName: types.NamespacedName{
					Name:      dap.Name,
					Namespace: dap.Namespace,
				}})
		}
	}
	log.Log.Info(fmt.Sprintf("Triggering %+v", rr))
	return rr
}

// SetupWithManager sets up the controller with the Manager.
func (r *DynamicAuthorizationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&peerauthv1.DynamicAuthorizationPolicy{}).
		Complete(r)
}

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

	istiov1alpha1 "github.com/aweis89/istio-abac-policy/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	pod := corev1.Pod{}
	namespace := pod.ObjectMeta.Namespace
	podLabels := pod.ObjectMeta.Labels
	selector := labels.NewSelector()
	for k, v := range podLabels {
		requirement, err := labels.NewRequirement(namespace+"/"+k, selection.Equals, []string{v})
		if err != nil {
			logr.Error(err, "unable to build requirements")
			return ctrl.Result{}, err
		}
		selector.Add(*requirement)
	}

	dapList := istiov1alpha1.DynamicAuthorizationPolicyList{}
	_ = r.List(ctx, &dapList, &client.ListOptions{LabelSelector: selector})

	for _, dap := range dapList.Items {
		dap.Status.WatchingPods = append(dap.Status.WatchingPods, req.NamespacedName)
		err := r.Update(ctx, &dap)
		if err != nil {
			// logr.Error(err, "unable to update", "DynamicAuthorizationPolicy", dap.GetName())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

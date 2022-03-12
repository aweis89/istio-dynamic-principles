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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object.
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const podSelectorIndex = ".spec.podSelector"

func indexKey(key, val string) string {
	return fmt.Sprintf("%s=%s", key, val)
}
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	pod := &corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrapf(err, "unable to retrieve pod")
	}
	podLabels := pod.GetLabels()
	if podLabels == nil {
		return ctrl.Result{}, nil
	}
	for k, v := range podLabels {
		log.Info("querying DAPS with labels", "labelKey", k, "labelVal", v)
		dapList := v1.DynamicAuthorizationPolicyList{}
		err := r.List(ctx, &dapList, client.MatchingFields{podSelectorIndex: indexKey(k, v)})
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "unable to list associated DAPs")
		}
		for i := range dapList.Items {
			dap := dapList.Items[i]
			log.Info("triggering DAP", "name", dap.Name, "namespace", dap.Namespace)
			if dap.GetAnnotations() == nil {
				dap.SetAnnotations(map[string]string{})
			}
			dap.Annotations[fmt.Sprintf("trigger-reconcile-%s", pod.GetName())] = time.Now().String()
			err := r.Update(ctx, &dap)
			if err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "unable to update pod")
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
	return errors.Wrap(err, "unable to start POD controller")
}

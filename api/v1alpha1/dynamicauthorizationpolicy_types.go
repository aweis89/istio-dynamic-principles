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

package v1alpha1

import (
	securityv1beta1 "istio.io/api/security/v1beta1"
	v1beta1 "istio.io/api/type/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DynamicAuthorizationPolicySpec defines the desired state of DynamicAuthorizationPolicy
type DynamicAuthorizationPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DynamicPrinciples           []DynamicPrinciples         `json:"dynamicPrinciples"`
	AuthorizationPolicyTemplate AuthorizationPolicyTemplate `json:"template"`
}

type AuthorizationPolicyTemplate struct {
	Metadata metav1.ObjectMeta       `json:"metadata"`
	Spec     AuthorizationPolicySpec `json:"spec"`
}

// Seme fields are missing json tags in securityv1beta1.AuthorizationPolicy
type AuthorizationPolicySpec struct {
	Selector *v1beta1.WorkloadSelector `json:"selector,omitempty"`
	Rules    *[]securityv1beta1.Rule   `json:"rules,omitempty"`
	// Optional. The action to take if the request is matched with the rules. Default is ALLOW if not specified.
	Action securityv1beta1.AuthorizationPolicy_Action `json:"action,omitempty"`
}

type DynamicPrinciples struct {
	Name               string              `json:"name"`
	NamespaceSelectors []NamespaceSelector `json:"namespaceSelectors"`
	PodSelectors       []PodSelector       `json:"podSelectors"`
	// TrustDomain is the istio trust domain, defaults to cluster.local
	TrustDomain string `json:"trustDomain"`
}

type PodSelector struct {
	MatchLabels []MatchLabel `json:"matchLabels"`
}

type NamespaceSelector struct {
	MatchLabels []MatchLabel `json:"matchLabels"`
}

type MatchLabel map[string]string

// DynamicAuthorizationPolicyStatus defines the observed state of DynamicAuthorizationPolicy
type DynamicAuthorizationPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	WatchingPods []types.NamespacedName

	WatchingNamespaces []string
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DynamicAuthorizationPolicy is the Schema for the dynamicauthorizationpolicies API
type DynamicAuthorizationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamicAuthorizationPolicySpec   `json:"spec,omitempty"`
	Status DynamicAuthorizationPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DynamicAuthorizationPolicyList contains a list of DynamicAuthorizationPolicy
type DynamicAuthorizationPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DynamicAuthorizationPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DynamicAuthorizationPolicy{}, &DynamicAuthorizationPolicyList{})
}

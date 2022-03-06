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

package v1

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DynamicAuthorizationPolicySpec defines the desired state of DynamicAuthorizationPolicy
type DynamicAuthorizationPolicySpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	DynamicPolicies []DynamicPolicy `json:"dynamicPolicies"`
}

type DynamicPolicy struct {
	Name         string     `json:"name"`
	PodSelectors labels.Set `json:"podSelectors"`
	// +kubebuilder:default:="cluster.local"
	TrustDomain string `json:"trustDomain"`
}

func (dp DynamicPolicy) ListPods(ctx context.Context, c client.Client, pl *corev1.PodList) error {
	return c.List(ctx, pl, client.MatchingLabels(dp.PodSelectors))
}

// DynamicAuthorizationPolicyStatus defines the observed state of DynamicAuthorizationPolicy
type DynamicAuthorizationPolicyStatus struct {
	ServiceAccountPolicyMapping ServiceAccountPolicyMapping `json:"serviceAccountPolicyMapping"`
	// ServiceAccountPolicyMapping ServiceAccountPolicyMappingType `json:"serviceAccountPolicyMapping"`
}

type ServiceAccountPolicyMapping map[string]HashSet

// type ServiceAccountPolicyMapping struct {
// 	Map     map[string][]string
// 	hashset map[string]HashSet
// }

// func (sapm *ServiceAccountPolicyMapping) Get(key string) (HashSet, bool) {
// 	if *sapm == nil {
// 		*sapm = ServiceAccountPolicyMapping{}
// 	}
// 	if (*sapm)[key] == nil {
// 		(*sapm)[key] = HashSet{}
// 	}
// 	hashSet, ok := (*sapm)[key]
// 	return hashSet, ok
// }

func (sapm *ServiceAccountPolicyMapping) Map(policy DynamicPolicy, pod corev1.Pod) {
	sa := pod.Spec.ServiceAccountName
	principle := fmt.Sprintf("%s/ns/%s/sa/%s", policy.TrustDomain, pod.GetNamespace(), sa)
	// log.Info("adding principle", "principle", principle)
	sapm.add(policy.Name, principle)
}

func (sapm *ServiceAccountPolicyMapping) add(key, val string) {
	if *sapm == nil {
		*sapm = ServiceAccountPolicyMapping{}
	}
	if (*sapm)[key] == nil {
		(*sapm)[key] = HashSet{}
	}
	hashset := (*sapm)[key]
	hashset.Add(val)
	(*sapm)[key] = hashset
}

// func (sapm *ServiceAccountPolicyMapping) MarshalJSON() ([]byte, error) {
// 	if *sapm == nil {
// 		*sapm = ServiceAccountPolicyMapping{}
// 	}
// 	conertTo := map[string][]string{}
// 	for k, v := range *sapm {
// 		conertTo[k] = v.Slice()
// 	}
// 	bytes, err := json.Marshal(&conertTo)
// 	err = errors.Wrap(err, "unable to MarshalJSON")
// 	return bytes, err
// }
//
// func (sapm *ServiceAccountPolicyMapping) UnmarshalJSON(b []byte) error {
// 	if *sapm == nil {
// 		*sapm = ServiceAccountPolicyMapping{}
// 	}
// 	fromJSON := map[string][]string{}
// 	if err := json.Unmarshal(b, &fromJSON); err != nil {
// 		return err
// 	}
//
// 	for k, v := range fromJSON {
// 		(*sapm)[k] = FromSlice(v)
// 	}
// 	return nil
// }

// +kubebuilder:validation:type=array
type HashSet map[string]bool

// func (hs *HashSet) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(hs.Slice())
// }
//
// func (hs *HashSet) UnmarshalJSON(b []byte) error {
// 	fromSlice := []string{}
// 	if err := json.Unmarshal(b, &fromSlice); err != nil {
// 		return err
// 	}
// 	*hs = FromSlice(fromSlice)
// 	return nil
// }

func FromSlice(slice []string) HashSet {
	hs := HashSet{}
	for _, v := range slice {
		hs[v] = true
	}
	return hs
}

func (hs *HashSet) Get(val string) bool {
	_, ok := (*hs)[val]
	return ok
}

func (hs *HashSet) Add(val string) {
	if *hs == nil {
		*hs = make(HashSet)
	}
	(*hs)[val] = true
}

func (hs *HashSet) Slice() []string {
	if *hs == nil {
		*hs = make(HashSet)
	}
	keys := []string{}
	for k := range *hs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

func (dap *DynamicAuthorizationPolicy) GetPolicies() []DynamicPolicy {
	return dap.Spec.DynamicPolicies
}

// func (dap *DynamicAuthorizationPolicy) AddPolicyMapping(policyName, serviceAccountNamespace, serviceAccount string) {
// 	if dap.Status.ServiceAccountPolicyMapping == nil {
// 		dap.Status.ServiceAccountPolicyMapping = make(map[string][]string)
// 	}
// 	// TODO ensure uniqness
// 	mapping := dap.Status.ServiceAccountPolicyMapping
// 	saNamespaceName := serviceAccountNamespace + "/" + serviceAccount
// 	for _, sa := range mapping[policyName] {
// 		if sa == saNamespaceName {
// 			return
// 		}
// 	}
// 	mapping[policyName] = append(mapping[policyName], saNamespaceName)
// }

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

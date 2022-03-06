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

package v1_test

// func TestServiceAccountPolicyMapping_MarshalJSON(t *testing.T) {
// 	t.Parallel()
// 	tests := map[string]struct {
// 		sapm    *ServiceAccountPolicyMapping
// 		want    []byte
// 		wantErr bool
// 	}{
// 		"basic encode": {
// 			sapm: &ServiceAccountPolicyMapping{
// 				// "foo": HashSet{"cluster.local/ns/namepsace/sa/service-account": true},
// 				"foo": FromSlice([]string{"cluster.local/ns/namepsace/sa/service-account"}),
// 			},
// 			wantErr: false,
// 			want:    []byte(`{"foo":["cluster.local/ns/namepsace/sa/service-account"]}`),
// 		},
// 	}
// 	for name, tt := range tests {
// 		name, tt := name, tt
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()
// 			got, err := tt.sapm.MarshalJSON()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ServiceAccountPolicyMapping.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ServiceAccountPolicyMapping.MarshalJSON() = %s, want %s", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestServiceAccountPolicyMapping_UnmarshalJSON(t *testing.T) {
// 	t.Parallel()
// 	type args struct {
// 		b []byte
// 	}
// 	tests := map[string]struct {
// 		sapm    *ServiceAccountPolicyMapping
// 		args    args
// 		wantErr bool
// 		want    *ServiceAccountPolicyMapping
// 	}{
// 		"basic decode": {
// 			sapm: &ServiceAccountPolicyMapping{},
// 			// "foo": HashSet{"cluster.local/ns/namepsace/sa/service-account": true},
// 			// "foo": FromSlice([]string{"cluster.local/ns/namepsace/sa/service-account"}),
// 			wantErr: false,
// 			args:    args{[]byte(`{"foo":["cluster.local/ns/namepsace/sa/service-account"]}`)},
// 			// want:    []byte(`{"foo":["cluster.local/ns/namepsace/sa/service-account"]}`),
// 			want: &ServiceAccountPolicyMapping{
// 				"foo": FromSlice([]string{"cluster.local/ns/namepsace/sa/service-account"}),
// 			},
// 		},
// 	}
// 	for name, tt := range tests {
// 		name, tt := name, tt
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()
// 			err := tt.sapm.UnmarshalJSON(tt.args.b)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ServiceAccountPolicyMapping.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			for k, want := range *tt.want {
// 				if got, ok := tt.sapm.Get(k); !ok {
// 					t.Errorf("ServiceAccountPolicyMapping.UnmarshalJSON() want = %v, want %v", got, want)
// 				}
// 			}
// 		})
// 	}
// }

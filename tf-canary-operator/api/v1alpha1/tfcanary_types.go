/*
Copyright 2021.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TfCanarySpec defines the desired state of TfCanary
type TfCanarySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Models are the tensorflow containers running with their respective models
	Models []Model `json:"models"`
}

type Model struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Weight   int32  `json:"weight"`
}

// TfCanaryStatus defines the observed state of TfCanary
type TfCanaryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TfCanary is the Schema for the tfcanaries API
//+kubebuilder:subresource:status
type TfCanary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TfCanarySpec   `json:"spec,omitempty"`
	Status TfCanaryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TfCanaryList contains a list of TfCanary
type TfCanaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TfCanary `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TfCanary{}, &TfCanaryList{})
}

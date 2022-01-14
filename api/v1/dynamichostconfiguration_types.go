/*
Copyright 2022 Zhang Kunpeng.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AddressRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Mask  string `json:"mask,omitempty"`
}

type Bind struct {
	Address   string `json:"address,omitempty"`
	Port      int    `json:"port,omitempty"`
	Interface string `json:"interface,omitempty"`
}

// DynamicHostConfigurationSpec defines the desired state of DynamicHostConfiguration
type DynamicHostConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=4;6
	ProtocolVersion int            `json:"protocolVersion"`
	Listen          []Bind         `json:"listen"`
	Range           []AddressRange `json:"range"`
	Router          string         `json:"router,omitempty"`
	Lease           int            `json:"lease,omitempty"`
}

// DynamicHostConfigurationStatus defines the observed state of DynamicHostConfiguration
type DynamicHostConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State string `json:"state,omitempty"`
}

// +kubebuilder:printcolumn:JSONPath=".status.state",name=Status,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:resource:categories=pxe,shortName=dhcp,scope=Namespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DynamicHostConfiguration is the Schema for the dynamichostconfigurations API
type DynamicHostConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamicHostConfigurationSpec   `json:"spec,omitempty"`
	Status DynamicHostConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DynamicHostConfigurationList contains a list of DynamicHostConfiguration
type DynamicHostConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DynamicHostConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DynamicHostConfiguration{}, &DynamicHostConfigurationList{})
}

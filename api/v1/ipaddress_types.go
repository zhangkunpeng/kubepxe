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

// IPAddressSpec defines the desired state of IPAddress
type IPAddressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Host    string `json:"host"`
	Address string `json:"address"`
}

// IPAddressStatus defines the observed state of IPAddress
type IPAddressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Host       string         `json:"host"`
	Network    string         `json:"network"`
	Address    string         `json:"address"`
	Mac        string         `json:"mac"`
	Allocation AllocationType `json:"allocation"`
	Lease      metav1.Time    `json:"leaseTimestamp,omitempty"`
	State      IPAddressState `json:"state"`
}

// +kubebuilder:validation:Enum=Dynamic;Static
type AllocationType string

// +kubebuilder:validation:Enum=Pending;Running
type IPAddressState string

// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:printcolumn:JSONPath=".status.address",name=Address,type=string
// +kubebuilder:printcolumn:JSONPath=".status.network",name=network,type=string
// +kubebuilder:printcolumn:JSONPath=".status.leaseTimestamp",name=Lease,type=string
// +kubebuilder:printcolumn:JSONPath=".status.allocation",name=Allocation,type=string
// +kubebuilder:resource:categories=pxe,shortName=ip,scope=Namespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IPAddress is the Schema for the ipaddresses API
type IPAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPAddressSpec   `json:"spec,omitempty"`
	Status IPAddressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPAddressList contains a list of IPAddress
type IPAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAddress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPAddress{}, &IPAddressList{})
}

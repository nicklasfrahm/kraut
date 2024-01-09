/*
Copyright 2024.

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
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetadataMatcher defines a selector that matches against the metadata
// of a Kubernetes resource. Note that this supports regular expressions.
type MetadataMatcher struct {
	// Name is the metadata name. Supports regular expressions.
	//+kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// Namespace is the metadata namespace. Supports regular expressions.
	// If not specified, the namespace of the resource being matched is used.
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

func (m *MetadataMatcher) Matches(meta *metav1.ObjectMeta) (error, bool) {
	nameRegex, err := regexp.Compile(m.Name)
	if err != nil {
		return fmt.Errorf("invalid matcher: name: %s", err), false
	}

	// Default to the namespace of the resource being matched.
	namespaceExpression := m.Namespace
	if namespaceExpression == "" {
		namespaceExpression = meta.Namespace
	}

	namespaceRegex, err := regexp.Compile(namespaceExpression)
	if err != nil {
		return fmt.Errorf("invalid matcher: namespace: %s", err), false
	}

	return nil, nameRegex.MatchString(meta.Name) && namespaceRegex.MatchString(meta.Namespace)
}

// FirewallSpecHostSelector defines the host selector for the host that will enforce the firewall rules.
type FirewallSpecHostSelector struct {
	// MatchMetadata defines the metadata that must match for the host to be selected.
	// Note that this supports regular expressions.
	MatchMetadata MetadataMatcher `json:"matchMetadata,omitempty"`
}

// FirewallSpec defines the desired state of Firewall
type FirewallSpec struct {
	// HostSelector defines the host selector for the host that will enforce the firewall rules.
	//+kubebuilder:validation:Required
	HostSelector FirewallSpecHostSelector `json:"hostSelector,omitempty"`
}

// FirewallStatus defines the observed state of Firewall
type FirewallStatus struct {
	// HostCount is the number of hosts that are currently enforcing the firewall rules.
	HostCount int `json:"hostCount,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories={fw,firewall},shortName=fw,path=firewalls,singular=firewall
//+kubebuilder:printcolumn:name="Host-Selector",type=string,JSONPath=`.spec.hostSelector.matchMetadata.name`
//+kubebuilder:printcolumn:name="Host-Count",type=integer,JSONPath=`.status.hostCount`

// Firewall is the Schema for the firewalls API
type Firewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallSpec   `json:"spec,omitempty"`
	Status FirewallStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FirewallList contains a list of Firewall
type FirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Firewall `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Firewall{}, &FirewallList{})
}

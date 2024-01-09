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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Protocol is the protocol used to connect to the host.
type Protocol string

const (
	// ProtocolSSH is the TCP protocol.
	ProtocolSSH Protocol = "SSH"
)

const (
	// OSUbuntu is the Ubuntu operating system.
	OSUbuntu = "Ubuntu"
	// OSNXOS is the Cisco NX-OS operating system.
	OSNXOS = "NX-OS"
)

// OSVersion describes the version of an operating system.
type OSVersion string

// Major returns the major version of the operating system.
// Returns -1 if the version could not be parsed.
func (v OSVersion) Major() int {
	segments := strings.Split(strings.TrimPrefix(string(v), "v"), ".")

	if len(segments) > 0 {
		major, err := strconv.Atoi(segments[0])
		if err == nil {
			return major
		}
	}

	return -1
}

// Minor returns the minor version of the operating system.
// Returns -1 if the version could not be parsed.
func (v OSVersion) Minor() int {
	segments := strings.Split(string(v), ".")

	if len(segments) > 1 {
		minor, err := strconv.Atoi(segments[1])
		if err == nil {
			return minor
		}
	}

	return -1
}

// OSInfo describes the operating system of a system.
type OSInfo struct {
	// Name is the name of the operating system.
	Name string `json:"name,omitempty"`
	// Version is the version of the operating system.
	Version OSVersion `json:"version,omitempty"`
	// KernelVersion is the kernel version of the operating system.
	KernelVersion string `json:"kernelVersion,omitempty"`
}

// HostSpecSSHOptions defines the SSH connection options.
type HostSpecSSHOptions struct {
	// Fingerprint is the SSH host key fingerprint in the format `{algorithm}:{hash}`.
	Fingerprint string `json:"fingerprint,omitempty"`
	// User is the SSH user to connect as.
	User string `json:"user,omitempty"`
	// ProxyHost is the SSH proxy host to connect to.
	ProxyHost string `json:"proxyHost,omitempty"`
	// ProxyPort is the SSH proxy port to connect to.
	ProxyPort int `json:"proxyPort,omitempty"`
	// ProxyFingerprint is the SSH proxy host key fingerprint in the format `{algorithm}:{hash}`.
	ProxyFingerprint string `json:"proxyFingerprint,omitempty"`
	// ProxyUser is the SSH proxy user to connect as.
	ProxyUser string `json:"proxyUser,omitempty"`
}

// HostSpec defines the desired state of Host
type HostSpec struct {
	// Host is the host to connect to.
	//+kubebuilder:validation:Required
	Host string `json:"host,omitempty"`
	// Port is the port to connect to.
	Port int `json:"port,omitempty"`
	// Protocol is the protocol used to connect to the host.
	// Currently only supports `SSH`.
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Enum=SSH
	Protocol Protocol `json:"protocol"`
	// SSH contains additional SSH connection options.
	SSH HostSpecSSHOptions `json:"ssh,omitempty"`
	// SecretRef is the reference to a secret containing sensitive connection credentials.
	//+kubebuilder:validation:Required
	SecretRef corev1.SecretReference `json:"secretRef"`
}

// HostStatus defines the observed state of Host
type HostStatus struct {
	// OS contains information about the discovered operating system.
	OS OSInfo `json:"os,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories={mgmt,management},shortName=host,path=hosts,singular=host
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.host`
//+kubebuilder:printcolumn:name="Protocol",type=string,JSONPath=`.spec.protocol`
//+kubebuilder:printcolumn:name="OS-Name",type=string,JSONPath=`.status.os.name`
//+kubebuilder:printcolumn:name="OS-Version",type=string,JSONPath=`.status.os.version`
//+kubebuilder:printcolumn:name="Kernel-Version",type=string,JSONPath=`.status.os.kernelVersion`

// Host is the Schema for the hosts API
type Host struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostSpec   `json:"spec,omitempty"`
	Status HostStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HostList contains a list of Host
type HostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Host `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Host{}, &HostList{})
}

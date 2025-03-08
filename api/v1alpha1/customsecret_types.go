/*
Copyright 2025 heheh.

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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomSecretSpec defines the desired state of CustomSecret.
type CustomSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SecretType   corev1.SecretType `json:"secretType"`
	RotationTime time.Duration     `json:"rotationTime"`
}

type SecretUpdateStatus struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	UpdatedAt metav1.Time `json:"updatedAt"`
}

// CustomSecretStatus defines the observed state of CustomSecret.
type CustomSecretStatus struct {
	UpdatedSecrets []SecretUpdateStatus `json:"secreteUpdatedstatus"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CustomSecret is the Schema for the customsecrets API.
type CustomSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomSecretSpec   `json:"spec,omitempty"`
	Status CustomSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomSecretList contains a list of CustomSecret.
type CustomSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomSecret{}, &CustomSecretList{})
}

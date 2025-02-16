/*
Copyright 2025.

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

type CRDInstallationStatus string
type ControllerInstallationStatus string

const (
	CRDInstallationStatusInvalidHelmChartValues CRDInstallationStatus        = "InvalidHelmChartValues"
	CRDInstallationStatusHelmChartFailed        CRDInstallationStatus        = "HelmChartFailedToRender"
	CRDInstallationStatusNoCRDsFound            CRDInstallationStatus        = "NoCRDsFoundInHelmChart"
	CRDInstallationStatusInstalled              CRDInstallationStatus        = "Installed"
	ControllerInstallationStatusPending         ControllerInstallationStatus = "Pending"
	ControllerInstallationStatusInstallFailed   ControllerInstallationStatus = "InstallFailed"
	ControllerInstallationStatusInstalled       ControllerInstallationStatus = "Installed"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ControllerWatchSpec defines the desired state of ControllerWatch.
type ControllerWatchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The helm install options where the CRD and controller to install and watch are defined
	HelmControllerSpec HelmInstallSpec `json:"helmSpec,omitempty"`
}

type HelmInstallSpec struct {
	// The name [location] of the chart to install
	Chart string `json:"chart"`
	// The namespace to install the chart into
	Namespace string `json:"namespace"`
	// The release name of the chart to install
	ReleaseName string `json:"releaseName"`
	// The version of the chart to install
	// +optional
	Version string `json:"version,omitempty"`
	// Optional helm values to pass to the chart. Should be a valid yaml or json string
	// +optional
	Values string `json:"values,omitempty"`
	// CreateNamespace if true will create the namespace if it does not exist
	// +optional
	CreateNamespace *bool `json:"createNamespace,omitempty"`
}

// ControllerWatchStatus defines the observed state of ControllerWatch.
type ControllerWatchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The status of the CRD installation
	// +optional
	CRDsInstallationStatus CRDInstallationStatus `json:"crdInstallationStatus,omitempty"`

	// The list of CRDs that were installed for this controller
	// +optional
	InstalledCRDs []GroupVersionKind `json:"installedCRDs,omitempty"`

	// The status of the controller installation
	// +optional
	ControllerInstallationStatus ControllerInstallationStatus `json:"controllerInstallationStatus,omitempty"`

	// LastUpdated is the last time which this status was updated
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ControllerWatch is the Schema for the controllerwatches API.
type ControllerWatch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControllerWatchSpec   `json:"spec,omitempty"`
	Status ControllerWatchStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ControllerWatchList contains a list of ControllerWatch.
type ControllerWatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControllerWatch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControllerWatch{}, &ControllerWatchList{})
}

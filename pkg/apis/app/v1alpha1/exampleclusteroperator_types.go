package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExampleClusterOperatorSpec defines the desired state of ExampleClusterOperator
// +k8s:openapi-gen=true
type ExampleClusterOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	OperatorAvailable   string `json:"operatorAvailable,omitempty"`
	OperatorProgressing string `json:"operatorProgressing,omitempty"`
	OperatorDegraded    string `json:"operatorDegraded,omitempty"`
	OperatorUpgradeable string `json:"operatorUpgradeable,omitempty"`
}

// ExampleClusterOperatorStatus defines the observed state of ExampleClusterOperator
// +k8s:openapi-gen=true
type ExampleClusterOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExampleClusterOperator is the Schema for the exampleclusteroperators API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ExampleClusterOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExampleClusterOperatorSpec   `json:"spec,omitempty"`
	Status ExampleClusterOperatorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExampleClusterOperatorList contains a list of ExampleClusterOperator
type ExampleClusterOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExampleClusterOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExampleClusterOperator{}, &ExampleClusterOperatorList{})
}

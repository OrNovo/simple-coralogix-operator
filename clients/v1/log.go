package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LogSpec struct {
	LogMessage string `json:"logMessage"`
}

type Log struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec LogSpec `json:"spec"`
}

type LogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Log `json:"items"`
}

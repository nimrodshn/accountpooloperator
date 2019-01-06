/*
Copyright (c) 2018 Red Hat, Inc.

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
	"fmt"
	"github.com/segmentio/ksuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudProvider is a type of cloud provider i.e. "aws", "azure", "gcp" etc..
type CloudProvider string

const (
	//AWSCloudProvider constant for the aws cloud.
	AWSCloudProvider CloudProvider = "AWS"
	//GoogleCloudProvider constant for the gcp cloud.
	GoogleCloudProvider CloudProvider = "GCP"
	//AzureCloudProvider constant for the azure cloud.
	AzureCloudProvider CloudProvider = "Azure"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AccountPool represents a Pool of avavilable accounts for OSD cluster deployment.
type AccountPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec AccountPoolSpec `json:"spec"`
}

// AccountPoolSpec is a struct with fields required for the construction AccountPool struct.
type AccountPoolSpec struct {
	// Provider is the cloud provider name.
	Provider CloudProvider `json:"cloud_provider"`

	// Credentials is a reference to a secret with the credentials for a user with 'Organization:CreateAccount' Permissions.
	Credentials corev1.LocalObjectReference `json:"credentials"`

	// AccountPoolSize is the size of the account pool.
	PoolSize int `json:"pool_size"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AccountPoolList is a list of account managers.
type AccountPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AccountPool `json:"items"`
}

// NewAWSAccountPool creates a new account pool and immediately creats AWS account CRD's
// to fill that pool - The accounts are not assigned to anyone.
func NewAWSAccountPool(
	namespace string,
	poolSize int,
	credentials corev1.LocalObjectReference) (*AccountPool, error) {
	uid, err := ksuid.NewRandom()
	if err != nil {
		return nil, err
	}
	accountPoolName := fmt.Sprintf("libra-%s", uid.String())

	return &AccountPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      accountPoolName,
			Namespace: namespace,
		},
		Spec: AccountPoolSpec{
			Credentials: credentials,
			PoolSize:    poolSize,
			Provider:    AWSCloudProvider,
		},
	}, nil
}

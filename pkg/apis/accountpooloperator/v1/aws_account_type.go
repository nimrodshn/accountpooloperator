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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// AccountStatus represents the status of the account.
type AccountStatus string

const (
	//StatusReady represents an account which is ready for use.
	StatusReady AccountStatus = "Ready"
	//StatusInstalling represents an account which is installing.
	StatusInstalling AccountStatus = "Installing"
	//StatusError represents an account which has an error.
	StatusError AccountStatus = "Error"
	//StatusPending represents an account which is to be installed.
	StatusPending AccountStatus = "Pending"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSAccount represents an aws account attached to a user.
type AWSAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec AWSAccountSpec `json:"spec"`
}

// AWSAccountSpec represents the spec for the aws account.
type AWSAccountSpec struct {
	AccountName string        `json:"account_name"`
	Email       string        `json:"email"`
	Status      AccountStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSAccountList represents a list of aws accounts.
type AWSAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AWSAccount `json:"items"`
}

// NewAvailableAccount creates a new account in a given namespace.
// The account is not assigned to anyone.
func NewAvailableAccount(namespace, poolName string) (*AWSAccount, error) {
	uid, err := ksuid.NewRandom()
	if err != nil {
		return nil, err
	}

	accountLabels := map[string]string{
		"available": "true",
		"pool_name": poolName,
	}

	accountName := fmt.Sprintf("libra-%s", strings.ToLower(uid.String()))

	return &AWSAccount{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: accountName,
			Name:         accountName,
			Namespace:    namespace,
			Labels:       accountLabels,
		},
		Spec: AWSAccountSpec{
			AccountName: accountName,
			Email:       "test@test.com",
			Status:      StatusPending,
		},
	}, nil
}

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

package utils

import (
	"fmt"

	accountpool "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	clientset "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Select only the available accounts.

// ClaimAccount claims an AWS account and returns the claimed account or
// returns an error if one occurred.
func ClaimAccount(clientset clientset.Interface, user string) (*accountpool.AWSAccount, error) {
	labelSelector := fmt.Sprintf("available=true")
	noAccountErr := fmt.Errorf("there are no available accounts at this point please try in a few minutes")
	accountList, err := clientset.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	if len(accountList.Items) == 0 {
		return nil, noAccountErr
	}
	for _, account := range accountList.Items {
		if account.Spec.Status == accountpool.StatusReady {
			// Claim the first available account.
			account.Labels["available"] = "false"
			account.Labels["user"] = user
			_, err := clientset.AccountpooloperatorV1().AWSAccounts(account.Namespace).Update(&account)
			if err != nil {
				// Print an error and attempt to claim another account.
				fmt.Printf("An error occurred while trying to claim the account %v: %v",
					account.Name, err)
				continue
			}
			return &account, nil
		}
	}
	// If reached here there are no available and healthy accounts
	return nil, noAccountErr
}

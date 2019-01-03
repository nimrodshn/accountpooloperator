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
	"testing"
	//nolint
	fakeclient "gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/client/clientset/versioned/fake"
	//nolint
	accountpool "gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/apis/accountpooloperator/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// create an account pool of size five.
	testPool = &accountpool.AccountPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test_pool",
		},
		Spec: accountpool.AccountPoolSpec{
			Provider: accountpool.AWSCloudProvider,
			PoolSize: 5,
		},
	}
)

type fixture struct {
	t *testing.T

	client *fakeclient.Clientset
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.client = fakeclient.NewSimpleClientset()
	f.t = t
	return f
}

func TestClaimAccount(t *testing.T) {
	f := newFixture(t)

	acc, err := accountpool.NewAvailableAccount(metav1.NamespaceDefault, testPool.Name)
	if err != nil {
		f.t.Errorf("error creating account: %v", err)
	}

	_, err = f.client.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceDefault).Create(acc)
	if err != nil {
		f.t.Errorf("error adding account: %v", err)
	}

	acc.Spec.Status = accountpool.StatusReady
	_, err = f.client.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceDefault).Update(acc)
	if err != nil {
		f.t.Errorf("error updating account: %v", err)
	}

	_, err = ClaimAccount(f.client, testPool.Name, "test-user")
	if err != nil {
		f.t.Errorf("error claiming account: %v", err)
	}

	claimedAccount, err := f.client.AccountpooloperatorV1().
		AWSAccounts(metav1.NamespaceDefault).
		Get(acc.Name, metav1.GetOptions{})
	if err != nil {
		f.t.Errorf("error getting account: %v", err)
	}
	if claimedAccount.Labels["available"] != "false" || claimedAccount.Labels["user"] != "test-user" {
		f.t.Errorf("Expected 'available' label to be 'false' instead it was 'true'.")
	}
}

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
package controllers

import (
	"context"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"testing"
	"time"

	accountpool "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	fakeclient "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned/fake"
	informerfactory "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// create an account pool of size five.
	testPool = &accountpool.AccountPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_pool",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: accountpool.AccountPoolSpec{
			Provider:    accountpool.AWSCloudProvider,
			Credentials: corev1.LocalObjectReference{},
			PoolSize:    5,
		},
	}
)

type fakeAccountProvisioner struct{}

func (f *fakeAccountProvisioner) ProvisionAccount(account *accountpool.AWSAccount,
	creds corev1.LocalObjectReference,
	stopCh <-chan struct{}) {
}

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

func (f *fixture) newController(stopCh <-chan struct{}) *AWSAccountController {
	fakeAccountProvisioner := &fakeAccountProvisioner{}
	accountControllerFactory := &AWSAccountControllerFactory{
		accountprovisioner:  fakeAccountProvisioner,
		awsaccountclientset: f.client,
		factory:             informerfactory.NewSharedInformerFactory(f.client, resyncPeriod),
		stopCh:              stopCh,
	}

	controller := accountControllerFactory.CreateControllerAndRun()
	return controller
}

func getKey(acc *accountpool.AWSAccount, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(acc)
	if err != nil {
		t.Errorf("Unexpected error getting key for foo %v: %v", acc.Name, err)
		return ""
	}
	return key
}

func TestSyncAWSAccount(t *testing.T) {
	f := newFixture(t)
	// Use a timeout to keep the test from hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := accountpool.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	// runController with account pool of size 5.
	// adding acc means we should add four more accounts.
	c := f.newController(ctx.Done())

	_, err = f.client.AccountpooloperatorV1().AccountPools(metav1.NamespaceDefault).Create(testPool)
	if err != nil {
		f.t.Errorf("error adding accountpool: %v", err)
	}

	acc, err := accountpool.NewAvailableAccount(metav1.NamespaceDefault, testPool.Name)
	if err != nil {
		f.t.Errorf("error creating account: %v", err)
	}

	// Actually create the account.
	_, err = f.client.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceDefault).Create(acc)
	if err != nil {
		f.t.Errorf("error adding account: %v", err)
	}

	for !c.synced() {
		f.t.Logf("waiting for sync...")
		time.Sleep(1 * time.Millisecond)
	}
	err = c.reconcileAccounts(getKey(acc, f.t))
	if err != nil {
		f.t.Errorf("error syncing account: %v", err)
	}

	accList, err := f.client.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceDefault).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Errorf("error adding account: %v", err)
	}

	if len(accList.Items) < testPool.Spec.PoolSize {
		f.t.Errorf("Expected controller to add accounts to fill"+
			"account pool with size %d instead account pool is of size: %d",
			testPool.Spec.PoolSize, len(accList.Items))
	}

	if ctx.Err() != nil {
		f.t.Errorf("timed out: %v", ctx.Err())
	}
}

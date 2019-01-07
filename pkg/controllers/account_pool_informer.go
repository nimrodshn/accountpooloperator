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
	"fmt"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	// nolint
	informers "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions/accountpooloperator/v1"
	// nolint
	accountpool "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	// nolint
	clientset "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned"
	// nolint
	informerfactory "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions"
)

// the number of worker threads.
const threadCount = 3

// The informers cache updates from etcd period.
const resyncPeriod = time.Minute * 10

// AccountPoolInformerFactory is a wrapper for AccountPoolInformer factory.
type AccountPoolInformerFactory struct {
	factory             informerfactory.SharedInformerFactory
	awsaccountclientset clientset.Interface
	stopCh              <-chan struct{}
	controllerMap       map[string]*AWSAccountController
}

// NewAccountPoolInformerFactory is a constructor for creating an AccountPoolInformerFactory.
func NewAccountPoolInformerFactory(config *rest.Config,
	stopCh <-chan struct{}) (*AccountPoolInformerFactory, error) {
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	factory := informerfactory.NewSharedInformerFactory(clientset, resyncPeriod)
	return &AccountPoolInformerFactory{
		factory:             factory,
		awsaccountclientset: clientset,
		stopCh:              stopCh,
		controllerMap:       make(map[string]*AWSAccountController),
	}, nil
}

// CreateInformerAndRun initializes an account pool informer and runs it.
func (f *AccountPoolInformerFactory) CreateInformerAndRun() *informers.AccountPoolInformer {

	// Creating account pool informer.
	accountpoolinformer := f.factory.Accountpooloperator().V1().AccountPools()

	glog.Info("Setting up event handlers...")
	accountpoolinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: f.addAccountPoolHandler,
	})

	glog.Info("Starting AccountPool informer..")
	go accountpoolinformer.Informer().Run(f.stopCh)
	return &accountpoolinformer
}

func (f *AccountPoolInformerFactory) addAccountPoolHandler(new interface{}) {
	newPool := new.(*accountpool.AccountPool)
	// Create the account pool if it does not exist.
	if !accountsExist(f.awsaccountclientset, newPool.Name) {
		createNewAvailablePool(f.awsaccountclientset, newPool)
	}
}

func createNewAvailablePool(
	awsaccountclientset clientset.Interface,
	newPool *accountpool.AccountPool) {
	for i := 0; i < newPool.Spec.PoolSize; i++ {
		account, err := accountpool.NewAvailableAccount(metav1.NamespaceDefault, newPool.Name)
		if err != nil {
			glog.Errorf("Could not create inital account pool.")
		}

		_, err = awsaccountclientset.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceDefault).Create(account)
		if err != nil {
			glog.Errorf("Failed to create new account: %v", err)
		}
	}
}

func accountsExist(awsaccountclientset clientset.Interface,
	poolName string) bool {
	lableSelector := fmt.Sprintf("pool_name=%v", poolName)
	accountList, err := awsaccountclientset.AccountpooloperatorV1().AWSAccounts(metav1.NamespaceDefault).List(
		metav1.ListOptions{
			LabelSelector: lableSelector,
		})
	if err != nil {
		glog.Errorf("An error occured trying to fetch account list")
		return false
	}
	if len(accountList.Items) > 0 {
		return true
	}
	return false
}

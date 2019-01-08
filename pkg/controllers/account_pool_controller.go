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
	"log"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	accountpool "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	clientset "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned"
	informerfactory "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions"
	informers "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions/accountpooloperator/v1"
)

// The informers cache updates from etcd period.
const resyncPeriod = time.Minute * 10

// AccountPoolControllerFactory is a wrapper for AccountPoolInformer factory.
type AccountPoolControllerFactory struct {
	factory             informerfactory.SharedInformerFactory
	awsaccountclientset clientset.Interface
	stopCh              <-chan struct{}
	controllerMap       map[string]*AWSAccountController
}

// NewAccountPoolControllerFactory is a constructor for creating an AccountPoolInformerFactory.
func NewAccountPoolControllerFactory(config *rest.Config,
	stopCh <-chan struct{}) (*AccountPoolControllerFactory, error) {
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	factory := informerfactory.NewSharedInformerFactory(clientset, resyncPeriod)
	return &AccountPoolControllerFactory{
		factory:             factory,
		awsaccountclientset: clientset,
		stopCh:              stopCh,
		controllerMap:       make(map[string]*AWSAccountController),
	}, nil
}

// CreateControllerAndRun initializes an account pool informer and runs it.
func (f *AccountPoolControllerFactory) CreateControllerAndRun() *informers.AccountPoolInformer {

	// Creating account pool informer.
	accountpoolinformer := f.factory.Accountpooloperator().V1().AccountPools()

	log.Println("Setting up event handlers...")
	accountpoolinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: f.addAccountPoolHandler,
	})

	log.Println("Starting Account Pool Controller..")
	go accountpoolinformer.Informer().Run(f.stopCh)
	return &accountpoolinformer
}

func (f *AccountPoolControllerFactory) addAccountPoolHandler(new interface{}) {
	newPool := new.(*accountpool.AccountPool)
	// Create the account pool if it does not exist.
	if !f.accountsExistInPool(newPool) {
		log.Println("Creating new AccountPool...")
		f.createNewAvailablePool(newPool)
	}
}

func (f *AccountPoolControllerFactory) createNewAvailablePool(newPool *accountpool.AccountPool) {
	for i := 0; i < newPool.Spec.PoolSize; i++ {
		account, err := accountpool.NewAvailableAccount(newPool.Namespace, newPool.Name)
		if err != nil {
			glog.Errorf("Could not create inital account pool.")
		}

		_, err = f.awsaccountclientset.AccountpooloperatorV1().AWSAccounts(newPool.Namespace).Create(account)
		if err != nil {
			glog.Errorf("Failed to create new account: %v", err)
		}
	}
}

func (f *AccountPoolControllerFactory) accountsExistInPool(pool *accountpool.AccountPool) bool {
	lableSelector := fmt.Sprintf("pool_name=%v", pool.Name)
	accountList, err := f.awsaccountclientset.
		AccountpooloperatorV1().
		AWSAccounts(pool.Namespace).
		List(
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

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

	accountpool "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	clientset "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned"
	informerfactory "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions"
	informers "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions/accountpooloperator/v1"
	listers "github.com/nimrodshn/accountpooloperator/pkg/client/listers/accountpooloperator/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nimrodshn/accountpooloperator/pkg/accountprovisioner"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// AWSAccountControllerFactory is a factory for AWSAccountController
type AWSAccountControllerFactory struct {
	awsaccountclientset clientset.Interface
	factory             informerfactory.SharedInformerFactory
	accountprovisioner  accountprovisioner.AccountProvisioner
	stopCh              <-chan struct{}
}

// NewAWSAccountControllerFactory is a constructor for AWSAccountControllerFactory
func NewAWSAccountControllerFactory(
	config *rest.Config,
	accountprovisioner accountprovisioner.AccountProvisioner,
	stopCh <-chan struct{}) (*AWSAccountControllerFactory, error) {
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	factory := informerfactory.NewSharedInformerFactory(clientset, resyncPeriod)
	return &AWSAccountControllerFactory{
		awsaccountclientset: clientset,
		accountprovisioner:  accountprovisioner,
		factory:             factory,
		stopCh:              stopCh,
	}, nil
}

// CreateControllerAndRun initializes an account controller and runs it.
func (f *AWSAccountControllerFactory) CreateControllerAndRun() *AWSAccountController {
	informer := f.factory.Accountpooloperator().V1().AWSAccounts()
	accountController := NewAWSAccountController(
		f.awsaccountclientset,
		informer,
		f.accountprovisioner,
		f.stopCh)

	log.Println("Starting AWS Account Controller..")
	// Running informer loop.
	go informer.Informer().Run(f.stopCh)

	return accountController
}

// AWSAccountController is a controller for the managed accounts.
type AWSAccountController struct {
	accountprovisioner  accountprovisioner.AccountProvisioner
	awsaccountclientset clientset.Interface
	awsaccountlister    listers.AWSAccountLister
	synced              cache.InformerSynced
	workqueue           workqueue.RateLimitingInterface
	stopCh              <-chan struct{}
}

// NewAWSAccountController is a constructor for the Controller
func NewAWSAccountController(
	awsaccountclientset clientset.Interface,
	awsaccountinformer informers.AWSAccountInformer,
	accountprovisioner accountprovisioner.AccountProvisioner,
	stopCh <-chan struct{},
) *AWSAccountController {

	controller := &AWSAccountController{
		awsaccountclientset: awsaccountclientset,
		awsaccountlister:    awsaccountinformer.Lister(),
		synced:              awsaccountinformer.Informer().HasSynced,
		workqueue:           workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		accountprovisioner:  accountprovisioner,
		stopCh:              stopCh,
	}

	log.Println("Setting up event handlers...")

	awsaccountinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			controller.addAccountHandler(new)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.reconcileAccounts(new)
		},
		DeleteFunc: func(old interface{}) {
			controller.reconcileAccounts(old)
		},
	})
	return controller
}

func (c *AWSAccountController) addAccountHandler(new interface{}) {
	newAccount := new.(*accountpool.AWSAccount)
	log.Printf("processing new account: %v\n", newAccount.Spec.AccountName)
	pool, err := c.retrievePoolFromAccount(newAccount)
	if err != nil {
		log.Printf("Could not retreieve account pool: %s\n", newAccount.Labels["pool_name"])
	}
	// run provision account job.
	go c.accountprovisioner.ProvisionAccount(
		newAccount,
		pool.Spec.Credentials,
		c.stopCh)
}

// reconcileAccounts receives an event (adding, deleting, updating of an accout) and tries to
// reconcile the current list of available accounts with respect to the desired pool size
// as described by the AccountPool.
func (c *AWSAccountController) reconcileAccounts(new interface{}) {
	account := new.(*accountpool.AWSAccount)
	namespace := account.Namespace
	log.Printf("processing updated account: %v\n", account.Spec.AccountName)

	// List the available accounts in the pool of the updated account -
	// We need to check if updating the account resulted with
	// insufficient available account in the pool.
	poolName := account.Labels["pool_name"]
	availabelSelector := fmt.Sprintf("pool_name = %v, available = true", poolName)
	availableAccounts, err := c.awsaccountclientset.
		AccountpooloperatorV1().
		AWSAccounts(namespace).
		List(metav1.ListOptions{
			LabelSelector: availabelSelector,
		})
	if err != nil {
		log.Printf("could not reconcile: %s, an error occurred trying to list available accounts in pool %s: %v", account.Name, poolName, err)
		return
	}

	// Retrieve the actoual pool object from the account - this is needed in order to check the pool is kept full.
	pool, err := c.retrievePoolFromAccount(account)
	if err != nil {
		log.Printf("could not reconcile: %s, could not find pool  %s: %v", account.Name, poolName, err)
		return
	}

	// create the missing accounts in the pool (if exist).
	c.fillAccountPool(len(availableAccounts.Items), pool.Spec.PoolSize, namespace, pool.Name)

}

func (c *AWSAccountController) retrievePoolFromAccount(account *accountpool.AWSAccount) (*accountpool.AccountPool, error) {
	pool, err := c.awsaccountclientset.
		AccountpooloperatorV1().
		AccountPools(account.Namespace).
		Get(account.Labels["pool_name"], metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// Check if number of available accounts in etcd is smaller than the desired pool size.
// If yes - create the accounts.
func (c *AWSAccountController) fillAccountPool(availableAccountsSize, poolSize int, namespace, poolName string) {
	if availableAccountsSize < poolSize {
		log.Printf("AccountPool depleted - Number of available accounts: %d, Pool size: %d\n. Creating new accounts..",
			availableAccountsSize, poolSize)
		numOfMissingAcc := poolSize - availableAccountsSize
		for i := 0; i < numOfMissingAcc; i++ {
			acc, err := accountpool.NewAvailableAccount(namespace, poolName)
			if err != nil {
				log.Printf("error creating account: %v\n", err)
			}

			_, err = c.awsaccountclientset.AccountpooloperatorV1().AWSAccounts(namespace).Create(acc)
			if err != nil {
				log.Println("Failed to create new account..")
			}
		}
	}
}

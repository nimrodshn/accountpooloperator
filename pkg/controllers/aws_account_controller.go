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

	accountpool "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	clientset "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned"
	informerfactory "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions"
	informers "github.com/nimrodshn/accountpooloperator/pkg/client/informers/externalversions/accountpooloperator/v1"
	listers "github.com/nimrodshn/accountpooloperator/pkg/client/listers/accountpooloperator/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/nimrodshn/accountpooloperator/pkg/accountprovisioner"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const threadCount = 3

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

	// Running informer loop.
	f.factory.Start(f.stopCh)

	// Running informer workers.
	go accountController.Run(threadCount)
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

	log.Println("Setting up event handlers")

	awsaccountinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			controller.addAccountHandler(new)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueAccount(new)
		},
		DeleteFunc: func(old interface{}) {
			controller.enqueueAccount(old)
		},
	})
	return controller
}

func (c *AWSAccountController) enqueueAccount(obj interface{}) {
	var err error
	var key string
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
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

// Run runs the workers processing events from infromers cache.
func (c *AWSAccountController) Run(threadiness int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown() // makes sure there are no dangling goroutines

	// Start the informer factories to begin populating the informer caches
	log.Println("Starting AWS Account controller")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, c.stopCh)
	}

	log.Println("Started workers")
	<-c.stopCh
	log.Println("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *AWSAccountController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item (Account) off the workqueue and
// attempt to process it, by calling the syncAccountHandler.
func (c *AWSAccountController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.reconcileAccounts(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		log.Printf("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// reconcileAccounts receives an event (adding, deleting, updating of an accout) and tries to
// reconcile the current list of available accounts with respect to the desired pool size
// as described by the AccountPool.
func (c *AWSAccountController) reconcileAccounts(key string) (err error) {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Retrieve the account from etcd.
	account, err := c.awsaccountlister.AWSAccounts(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("could not get account with name: %s", name))
		return nil
	}

	// List the available accounts in the pool of the updated account -
	// We need to check if updating the account resulted with
	// insufficiant available account in the pool.
	poolName := account.Labels["pool_name"]
	availabelSelector := fmt.Sprintf("pool_name = %v, available = true", poolName)
	availableAccounts, err := c.awsaccountclientset.
		AccountpooloperatorV1().
		AWSAccounts(namespace).
		List(metav1.ListOptions{
			LabelSelector: availabelSelector,
		})
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("account '%s' in work queue no longer exists", key))
			return nil
		}
		return nil
	}

	// Retrieve the actoual pool object from the account - this is needed in order to check the pool is kept full.
	pool, err := c.retrievePoolFromAccount(account)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Could not find pool  %s: %v",
			poolName, err))
		return nil
	}

	// create the missing accounts in the pool (if exist).
	c.fillAccountPool(len(availableAccounts.Items), pool.Spec.PoolSize, namespace, pool.Name)

	return nil
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

// poolSelectorRequirements creates the requirements (labels) for current available accounts in the associated account pool.
func (c *AWSAccountController) poolSelectorRequirements(poolName string) []labels.Requirement {
	result := make([]labels.Requirement, 2)
	// Errors in 'NewRequirement' are ignored as they cannot occure - see 'NewRequirement' docs.
	// Require label 'available' equals 'true'.
	availableRequirement, _ := labels.NewRequirement("available", selection.Equals, []string{"true"})
	// Require label 'pool_name' equals poolName.
	poolRequirement, _ := labels.NewRequirement("pool_name", selection.Equals, []string{poolName})

	result = append(result, *availableRequirement, *poolRequirement)
	return result
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

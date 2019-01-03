/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/apis/accountpooloperator/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AccountPoolLister helps list AccountPools.
type AccountPoolLister interface {
	// List lists all AccountPools in the indexer.
	List(selector labels.Selector) (ret []*v1.AccountPool, err error)
	// AccountPools returns an object that can list and get AccountPools.
	AccountPools(namespace string) AccountPoolNamespaceLister
	AccountPoolListerExpansion
}

// accountPoolLister implements the AccountPoolLister interface.
type accountPoolLister struct {
	indexer cache.Indexer
}

// NewAccountPoolLister returns a new AccountPoolLister.
func NewAccountPoolLister(indexer cache.Indexer) AccountPoolLister {
	return &accountPoolLister{indexer: indexer}
}

// List lists all AccountPools in the indexer.
func (s *accountPoolLister) List(selector labels.Selector) (ret []*v1.AccountPool, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.AccountPool))
	})
	return ret, err
}

// AccountPools returns an object that can list and get AccountPools.
func (s *accountPoolLister) AccountPools(namespace string) AccountPoolNamespaceLister {
	return accountPoolNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AccountPoolNamespaceLister helps list and get AccountPools.
type AccountPoolNamespaceLister interface {
	// List lists all AccountPools in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.AccountPool, err error)
	// Get retrieves the AccountPool from the indexer for a given namespace and name.
	Get(name string) (*v1.AccountPool, error)
	AccountPoolNamespaceListerExpansion
}

// accountPoolNamespaceLister implements the AccountPoolNamespaceLister
// interface.
type accountPoolNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AccountPools in the indexer for a given namespace.
func (s accountPoolNamespaceLister) List(selector labels.Selector) (ret []*v1.AccountPool, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.AccountPool))
	})
	return ret, err
}

// Get retrieves the AccountPool from the indexer for a given namespace and name.
func (s accountPoolNamespaceLister) Get(name string) (*v1.AccountPool, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("accountpool"), name)
	}
	return obj.(*v1.AccountPool), nil
}

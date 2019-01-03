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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	scheme "github.com/nimrodshn/accountpooloperator/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AWSAccountsGetter has a method to return a AWSAccountInterface.
// A group's client should implement this interface.
type AWSAccountsGetter interface {
	AWSAccounts(namespace string) AWSAccountInterface
}

// AWSAccountInterface has methods to work with AWSAccount resources.
type AWSAccountInterface interface {
	Create(*v1.AWSAccount) (*v1.AWSAccount, error)
	Update(*v1.AWSAccount) (*v1.AWSAccount, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.AWSAccount, error)
	List(opts metav1.ListOptions) (*v1.AWSAccountList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.AWSAccount, err error)
	AWSAccountExpansion
}

// aWSAccounts implements AWSAccountInterface
type aWSAccounts struct {
	client rest.Interface
	ns     string
}

// newAWSAccounts returns a AWSAccounts
func newAWSAccounts(c *AccountpooloperatorV1Client, namespace string) *aWSAccounts {
	return &aWSAccounts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the aWSAccount, and returns the corresponding aWSAccount object, and an error if there is any.
func (c *aWSAccounts) Get(name string, options metav1.GetOptions) (result *v1.AWSAccount, err error) {
	result = &v1.AWSAccount{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awsaccounts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AWSAccounts that match those selectors.
func (c *aWSAccounts) List(opts metav1.ListOptions) (result *v1.AWSAccountList, err error) {
	result = &v1.AWSAccountList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awsaccounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested aWSAccounts.
func (c *aWSAccounts) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("awsaccounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a aWSAccount and creates it.  Returns the server's representation of the aWSAccount, and an error, if there is any.
func (c *aWSAccounts) Create(aWSAccount *v1.AWSAccount) (result *v1.AWSAccount, err error) {
	result = &v1.AWSAccount{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("awsaccounts").
		Body(aWSAccount).
		Do().
		Into(result)
	return
}

// Update takes the representation of a aWSAccount and updates it. Returns the server's representation of the aWSAccount, and an error, if there is any.
func (c *aWSAccounts) Update(aWSAccount *v1.AWSAccount) (result *v1.AWSAccount, err error) {
	result = &v1.AWSAccount{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("awsaccounts").
		Name(aWSAccount.Name).
		Body(aWSAccount).
		Do().
		Into(result)
	return
}

// Delete takes name of the aWSAccount and deletes it. Returns an error if one occurs.
func (c *aWSAccounts) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awsaccounts").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *aWSAccounts) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awsaccounts").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched aWSAccount.
func (c *aWSAccounts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.AWSAccount, err error) {
	result = &v1.AWSAccount{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("awsaccounts").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

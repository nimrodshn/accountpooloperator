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

package fake

import (
	accountpooloperatorv1 "gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/apis/accountpooloperator/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAWSAccounts implements AWSAccountInterface
type FakeAWSAccounts struct {
	Fake *FakeAccountpooloperatorV1
	ns   string
}

var awsaccountsResource = schema.GroupVersionResource{Group: "accountpooloperator.openshift.io", Version: "v1", Resource: "awsaccounts"}

var awsaccountsKind = schema.GroupVersionKind{Group: "accountpooloperator.openshift.io", Version: "v1", Kind: "AWSAccount"}

// Get takes name of the aWSAccount, and returns the corresponding aWSAccount object, and an error if there is any.
func (c *FakeAWSAccounts) Get(name string, options v1.GetOptions) (result *accountpooloperatorv1.AWSAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(awsaccountsResource, c.ns, name), &accountpooloperatorv1.AWSAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*accountpooloperatorv1.AWSAccount), err
}

// List takes label and field selectors, and returns the list of AWSAccounts that match those selectors.
func (c *FakeAWSAccounts) List(opts v1.ListOptions) (result *accountpooloperatorv1.AWSAccountList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(awsaccountsResource, awsaccountsKind, c.ns, opts), &accountpooloperatorv1.AWSAccountList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &accountpooloperatorv1.AWSAccountList{ListMeta: obj.(*accountpooloperatorv1.AWSAccountList).ListMeta}
	for _, item := range obj.(*accountpooloperatorv1.AWSAccountList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested aWSAccounts.
func (c *FakeAWSAccounts) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(awsaccountsResource, c.ns, opts))

}

// Create takes the representation of a aWSAccount and creates it.  Returns the server's representation of the aWSAccount, and an error, if there is any.
func (c *FakeAWSAccounts) Create(aWSAccount *accountpooloperatorv1.AWSAccount) (result *accountpooloperatorv1.AWSAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(awsaccountsResource, c.ns, aWSAccount), &accountpooloperatorv1.AWSAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*accountpooloperatorv1.AWSAccount), err
}

// Update takes the representation of a aWSAccount and updates it. Returns the server's representation of the aWSAccount, and an error, if there is any.
func (c *FakeAWSAccounts) Update(aWSAccount *accountpooloperatorv1.AWSAccount) (result *accountpooloperatorv1.AWSAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(awsaccountsResource, c.ns, aWSAccount), &accountpooloperatorv1.AWSAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*accountpooloperatorv1.AWSAccount), err
}

// Delete takes name of the aWSAccount and deletes it. Returns an error if one occurs.
func (c *FakeAWSAccounts) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(awsaccountsResource, c.ns, name), &accountpooloperatorv1.AWSAccount{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAWSAccounts) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(awsaccountsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &accountpooloperatorv1.AWSAccountList{})
	return err
}

// Patch applies the patch and returns the patched aWSAccount.
func (c *FakeAWSAccounts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *accountpooloperatorv1.AWSAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(awsaccountsResource, c.ns, name, data, subresources...), &accountpooloperatorv1.AWSAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*accountpooloperatorv1.AWSAccount), err
}
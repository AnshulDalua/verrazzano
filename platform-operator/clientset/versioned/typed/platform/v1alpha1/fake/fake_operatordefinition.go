// Copyright (c) 2020, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/verrazzano/verrazzano/platform-operator/apis/platform/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeOperatorDefinitions implements OperatorDefinitionInterface
type FakeOperatorDefinitions struct {
	Fake *FakePlatformV1alpha1
	ns   string
}

var operatordefinitionsResource = schema.GroupVersionResource{Group: "platform", Version: "v1alpha1", Resource: "operatordefinitions"}

var operatordefinitionsKind = schema.GroupVersionKind{Group: "platform", Version: "v1alpha1", Kind: "OperatorDefinition"}

// Get takes name of the operatorDefinition, and returns the corresponding operatorDefinition object, and an error if there is any.
func (c *FakeOperatorDefinitions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.OperatorDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(operatordefinitionsResource, c.ns, name), &v1alpha1.OperatorDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OperatorDefinition), err
}

// List takes label and field selectors, and returns the list of OperatorDefinitions that match those selectors.
func (c *FakeOperatorDefinitions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.OperatorDefinitionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(operatordefinitionsResource, operatordefinitionsKind, c.ns, opts), &v1alpha1.OperatorDefinitionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.OperatorDefinitionList{ListMeta: obj.(*v1alpha1.OperatorDefinitionList).ListMeta}
	for _, item := range obj.(*v1alpha1.OperatorDefinitionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested operatorDefinitions.
func (c *FakeOperatorDefinitions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(operatordefinitionsResource, c.ns, opts))

}

// Create takes the representation of a operatorDefinition and creates it.  Returns the server's representation of the operatorDefinition, and an error, if there is any.
func (c *FakeOperatorDefinitions) Create(ctx context.Context, operatorDefinition *v1alpha1.OperatorDefinition, opts v1.CreateOptions) (result *v1alpha1.OperatorDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(operatordefinitionsResource, c.ns, operatorDefinition), &v1alpha1.OperatorDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OperatorDefinition), err
}

// Update takes the representation of a operatorDefinition and updates it. Returns the server's representation of the operatorDefinition, and an error, if there is any.
func (c *FakeOperatorDefinitions) Update(ctx context.Context, operatorDefinition *v1alpha1.OperatorDefinition, opts v1.UpdateOptions) (result *v1alpha1.OperatorDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(operatordefinitionsResource, c.ns, operatorDefinition), &v1alpha1.OperatorDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OperatorDefinition), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeOperatorDefinitions) UpdateStatus(ctx context.Context, operatorDefinition *v1alpha1.OperatorDefinition, opts v1.UpdateOptions) (*v1alpha1.OperatorDefinition, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(operatordefinitionsResource, "status", c.ns, operatorDefinition), &v1alpha1.OperatorDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OperatorDefinition), err
}

// Delete takes name of the operatorDefinition and deletes it. Returns an error if one occurs.
func (c *FakeOperatorDefinitions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(operatordefinitionsResource, c.ns, name, opts), &v1alpha1.OperatorDefinition{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeOperatorDefinitions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(operatordefinitionsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.OperatorDefinitionList{})
	return err
}

// Patch applies the patch and returns the patched operatorDefinition.
func (c *FakeOperatorDefinitions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.OperatorDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(operatordefinitionsResource, c.ns, name, pt, data, subresources...), &v1alpha1.OperatorDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OperatorDefinition), err
}

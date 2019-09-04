/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
	scheme "github.com/openebs/csi/pkg/generated/clientset/maya/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// JivaVolumeClaimsGetter has a method to return a JivaVolumeClaimInterface.
// A group's client should implement this interface.
type JivaVolumeClaimsGetter interface {
	JivaVolumeClaims(namespace string) JivaVolumeClaimInterface
}

// JivaVolumeClaimInterface has methods to work with JivaVolumeClaim resources.
type JivaVolumeClaimInterface interface {
	Create(*v1alpha1.JivaVolumeClaim) (*v1alpha1.JivaVolumeClaim, error)
	Update(*v1alpha1.JivaVolumeClaim) (*v1alpha1.JivaVolumeClaim, error)
	UpdateStatus(*v1alpha1.JivaVolumeClaim) (*v1alpha1.JivaVolumeClaim, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.JivaVolumeClaim, error)
	List(opts v1.ListOptions) (*v1alpha1.JivaVolumeClaimList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.JivaVolumeClaim, err error)
	JivaVolumeClaimExpansion
}

// jivaVolumeClaims implements JivaVolumeClaimInterface
type jivaVolumeClaims struct {
	client rest.Interface
	ns     string
}

// newJivaVolumeClaims returns a JivaVolumeClaims
func newJivaVolumeClaims(c *OpenebsV1alpha1Client, namespace string) *jivaVolumeClaims {
	return &jivaVolumeClaims{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the jivaVolumeClaim, and returns the corresponding jivaVolumeClaim object, and an error if there is any.
func (c *jivaVolumeClaims) Get(name string, options v1.GetOptions) (result *v1alpha1.JivaVolumeClaim, err error) {
	result = &v1alpha1.JivaVolumeClaim{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of JivaVolumeClaims that match those selectors.
func (c *jivaVolumeClaims) List(opts v1.ListOptions) (result *v1alpha1.JivaVolumeClaimList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.JivaVolumeClaimList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested jivaVolumeClaims.
func (c *jivaVolumeClaims) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a jivaVolumeClaim and creates it.  Returns the server's representation of the jivaVolumeClaim, and an error, if there is any.
func (c *jivaVolumeClaims) Create(jivaVolumeClaim *v1alpha1.JivaVolumeClaim) (result *v1alpha1.JivaVolumeClaim, err error) {
	result = &v1alpha1.JivaVolumeClaim{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		Body(jivaVolumeClaim).
		Do().
		Into(result)
	return
}

// Update takes the representation of a jivaVolumeClaim and updates it. Returns the server's representation of the jivaVolumeClaim, and an error, if there is any.
func (c *jivaVolumeClaims) Update(jivaVolumeClaim *v1alpha1.JivaVolumeClaim) (result *v1alpha1.JivaVolumeClaim, err error) {
	result = &v1alpha1.JivaVolumeClaim{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		Name(jivaVolumeClaim.Name).
		Body(jivaVolumeClaim).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *jivaVolumeClaims) UpdateStatus(jivaVolumeClaim *v1alpha1.JivaVolumeClaim) (result *v1alpha1.JivaVolumeClaim, err error) {
	result = &v1alpha1.JivaVolumeClaim{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		Name(jivaVolumeClaim.Name).
		SubResource("status").
		Body(jivaVolumeClaim).
		Do().
		Into(result)
	return
}

// Delete takes name of the jivaVolumeClaim and deletes it. Returns an error if one occurs.
func (c *jivaVolumeClaims) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *jivaVolumeClaims) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched jivaVolumeClaim.
func (c *jivaVolumeClaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.JivaVolumeClaim, err error) {
	result = &v1alpha1.JivaVolumeClaim{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("jivavolumeclaims").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

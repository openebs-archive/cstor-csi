// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"encoding/json"

	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"

	clientset "github.com/openebs/csi/pkg/generated/clientset/maya/internalclientset"
	client "github.com/openebs/csi/pkg/generated/maya/kubernetes/client/v1alpha1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// createFn is a typed function that abstracts
// creating csi volume instance
type createFn func(cs *clientset.Clientset, upgradeResultObj *apismaya.CStorVolumeClaim, namespace string) (*apismaya.CStorVolumeClaim, error)

// getFn is a typed function that abstracts
// fetching a csi volume instance
type getFn func(cli *clientset.Clientset, name, namespace string,
	opts metav1.GetOptions) (*apismaya.CStorVolumeClaim, error)

// listFn is a typed function that abstracts
// listing of csi volume instances
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apismaya.CStorVolumeClaimList, error)

// delFn is a typed function that abstracts
// deleting a csi volume instance
type delFn func(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error

// updateFn is a typed function that abstracts
// updating csi volume instance
type updateFn func(cs *clientset.Clientset, vol *apismaya.CStorVolumeClaim, namespace string) (*apismaya.CStorVolumeClaim, error)

// Kubeclient enables kubernetes API operations
// on csi volume instance
type Kubeclient struct {
	// clientset refers to csi volume's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	kubeConfigPath string

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	list                listFn
	del                 delFn
	create              createFn
	update              updateFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, err
			}

			return clientset.NewForConfig(config)
		}
	}

	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			config, err := client.GetConfig(client.New(client.WithKubeConfigPath(kubeConfigPath)))
			if err != nil {
				return nil, err
			}

			return clientset.NewForConfig(config)
		}
	}

	if k.create == nil {
		k.create = func(cli *clientset.Clientset, vol *apismaya.CStorVolumeClaim, namespace string) (*apismaya.CStorVolumeClaim, error) {
			return cli.OpenebsV1alpha1().CStorVolumeClaims(namespace).Create(vol)
		}
	}

	if k.get == nil {
		k.get = func(cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apismaya.CStorVolumeClaim, error) {
			return cli.OpenebsV1alpha1().CStorVolumeClaims(namespace).Get(name, opts)
		}
	}

	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apismaya.CStorVolumeClaimList, error) {
			return cli.OpenebsV1alpha1().CStorVolumeClaims(namespace).List(opts)
		}
	}

	if k.del == nil {
		k.del = func(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
			deletePropagation := metav1.DeletePropagationForeground
			opts.PropagationPolicy = &deletePropagation
			err := cli.OpenebsV1alpha1().CStorVolumeClaims(namespace).Delete(name, opts)
			return err
		}
	}

	if k.update == nil {
		k.update = func(cs *clientset.Clientset, vol *apismaya.CStorVolumeClaim, namespace string) (*apismaya.CStorVolumeClaim, error) {
			return cs.OpenebsV1alpha1().CStorVolumeClaims(namespace).Update(vol)
		}
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithNamespace sets the provided namespace
// against this Kubeclient instance
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubernetes client
// against the provided path
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeclient returns a new instance of
// kubeclient meant for csi volume operations
func NewKubeclient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}

	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}

	return k.getClientset()
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, err
	}

	k.clientset = c
	return k.clientset, nil
}

// Create creates a csi volume instance
// in kubernetes cluster
func (k *Kubeclient) Create(vol *apismaya.CStorVolumeClaim) (*apismaya.CStorVolumeClaim, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.create(cs, vol, k.namespace)
}

// Get returns csi volume object for given name
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apismaya.CStorVolumeClaim, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to get csi volume: missing csi volume name")
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.get(cli, name, k.namespace, opts)
}

// GetRaw returns csi volume instance
// in bytes
func (k *Kubeclient) GetRaw(name string, opts metav1.GetOptions) ([]byte, error) {
	csiv, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(csiv)
}

// List returns a list of csi volume
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apismaya.CStorVolumeClaimList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.list(cli, k.namespace, opts)
}

// Delete deletes the csi volume from
// kubernetes
func (k *Kubeclient) Delete(name string) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}

	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}

// Update updates this csi volume instance
// against kubernetes cluster
func (k *Kubeclient) Update(vol *apismaya.CStorVolumeClaim) (*apismaya.CStorVolumeClaim, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.update(cs, vol, k.namespace)
}

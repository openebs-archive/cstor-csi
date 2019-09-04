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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	mayav1alpha1 "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
	internalclientset "github.com/openebs/csi/pkg/generated/clientset/maya/internalclientset"
	internalinterfaces "github.com/openebs/csi/pkg/generated/informer/maya/externalversions/internalinterfaces"
	v1alpha1 "github.com/openebs/csi/pkg/generated/lister/maya/maya/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// JivaVolumeClaimInformer provides access to a shared informer and lister for
// JivaVolumeClaims.
type JivaVolumeClaimInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.JivaVolumeClaimLister
}

type jivaVolumeClaimInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewJivaVolumeClaimInformer constructs a new informer for JivaVolumeClaim type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewJivaVolumeClaimInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredJivaVolumeClaimInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredJivaVolumeClaimInformer constructs a new informer for JivaVolumeClaim type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredJivaVolumeClaimInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OpenebsV1alpha1().JivaVolumeClaims(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OpenebsV1alpha1().JivaVolumeClaims(namespace).Watch(options)
			},
		},
		&mayav1alpha1.JivaVolumeClaim{},
		resyncPeriod,
		indexers,
	)
}

func (f *jivaVolumeClaimInformer) defaultInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredJivaVolumeClaimInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *jivaVolumeClaimInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&mayav1alpha1.JivaVolumeClaim{}, f.defaultInformer)
}

func (f *jivaVolumeClaimInformer) Lister() v1alpha1.JivaVolumeClaimLister {
	return v1alpha1.NewJivaVolumeClaimLister(f.Informer().GetIndexer())
}
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

package v1alpha1

import (
	apisv1 "github.com/openebs/api/pkg/apis/cstor/v1"
)

// CStorVolumeConfig a wrapper for ume object
type CStorVolumeConfig struct {
	// actual cstorvolumeclaim object
	object *apisv1.CStorVolumeConfig
}

// List is a list of cstorvolumeclaim objects
type CStorVolumeConfigList struct {
	// list of cstor volume claims
	items []*CStorVolumeConfig
}

// ListBuilder enables building
// an instance of umeCStorVolumeConfigList
type ListBuilder struct {
	list    *CStorVolumeConfigList
	filters PredicateList
}

// NewListBuilder returns a new instance
// of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CStorVolumeConfigList{}}
}

// WithAPIList builds the list of cstorvolume claim
// instances based on the provided
// CStorVolumeConfig api instances
func (b *ListBuilder) WithAPIList(
	list *apisv1.CStorVolumeConfigList) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		c := c
		b.list.items = append(b.list.items, &CStorVolumeConfig{object: &c})
	}
	return b
}

// List returns the list of CStorVolumeConfigs (cvcs)
// instances that was built by this
// builder
func (b *ListBuilder) List() *CStorVolumeConfigList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &CStorVolumeConfigList{}
	for _, cv := range b.list.items {
		if b.filters.all(cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CStorVolumeConfigList
func (l *CStorVolumeConfigList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cstorvolume claim instance
type Predicate func(*CStorVolumeConfig) bool

// PredicateList holds a list of cstor volume claims
// based predicates
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided cstorvolumeclaim
// instance
func (l PredicateList) all(c *CStorVolumeConfig) bool {
	for _, check := range l {
		if !check(c) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cstorvolumeclaim has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// NewForAPIObject returns a new instance of cstorvolume
func NewForAPIObject(obj *apisv1.CStorVolumeConfig) *CStorVolumeConfig {
	return &CStorVolumeConfig{
		object: obj,
	}
}

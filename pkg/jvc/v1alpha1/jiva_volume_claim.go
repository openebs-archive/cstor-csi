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
	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
)

// JivaVolumeClaim a wrapper for ume object
type JivaVolumeClaim struct {
	// actual jivavolumeclaim object
	object *apismaya.JivaVolumeClaim
}

// List is a list of jivavolumeclaim objects
type JivaVolumeClaimList struct {
	// list of jiva volume claims
	items []*JivaVolumeClaim
}

// ListBuilder enables building
// an instance of umeJivaVolumeClaimList
type ListBuilder struct {
	list    *JivaVolumeClaimList
	filters PredicateList
}

// NewListBuilder returns a new instance
// of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &JivaVolumeClaimList{}}
}

// WithAPIList builds the list of jivavolume claim
// instances based on the provided
// JivaVolumeClaim api instances
func (b *ListBuilder) WithAPIList(
	list *apismaya.JivaVolumeClaimList) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		c := c
		b.list.items = append(b.list.items, &JivaVolumeClaim{object: &c})
	}
	return b
}

// List returns the list of JivaVolumeClaims (cvcs)
// instances that was built by this
// builder
func (b *ListBuilder) List() *JivaVolumeClaimList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &JivaVolumeClaimList{}
	for _, cv := range b.list.items {
		if b.filters.all(cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the JivaVolumeClaimList
func (l *JivaVolumeClaimList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided jivavolume claim instance
type Predicate func(*JivaVolumeClaim) bool

// PredicateList holds a list of jiva volume claims
// based predicates
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided jivavolumeclaim
// instance
func (l PredicateList) all(c *JivaVolumeClaim) bool {
	for _, check := range l {
		if !check(c) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the jivavolumeclaim has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// NewForAPIObject returns a new instance of jivavolume
func NewForAPIObject(obj *apismaya.JivaVolumeClaim) *JivaVolumeClaim {
	return &JivaVolumeClaim{
		object: obj,
	}
}

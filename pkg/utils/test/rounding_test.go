/*
 Copyright © 2020 The OpenEBS Authors

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

package test

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_RoundUpToGiB(t *testing.T) {
	testcases := []struct {
		name       string
		resource   resource.Quantity
		roundedVal int64
	}{
		{
			name:       "round Ki to GiB",
			resource:   resource.MustParse("1000Ki"),
			roundedVal: int64(1),
		},
		{
			name:       "round k to GiB",
			resource:   resource.MustParse("1000k"),
			roundedVal: int64(1),
		},
		{
			name:       "round Mi to GiB",
			resource:   resource.MustParse("1000Mi"),
			roundedVal: int64(1),
		},
		{
			name:       "round M to GiB",
			resource:   resource.MustParse("1000M"),
			roundedVal: int64(1),
		},
		{
			name:       "round G to GiB",
			resource:   resource.MustParse("1000G"),
			roundedVal: int64(932),
		},
		{
			name:       "round Gi to GiB",
			resource:   resource.MustParse("1000Gi"),
			roundedVal: int64(1000),
		},
		{
			name:       "round Gi to GiB",
			resource:   resource.MustParse("1500Gi"),
			roundedVal: int64(1500),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			val := RoundUpGiB(test.resource.Value())
			if val != test.roundedVal {
				t.Logf("actual rounded value: %d", val)
				t.Logf("expected rounded value: %d", test.roundedVal)
				t.Error("unexpected rounded value")
			}
		})
	}
}

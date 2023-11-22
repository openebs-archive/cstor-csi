#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file has been taken from https://github.com/kubernetes/code-generator/blob/master/generate-groups.sh
# A small modification is made in this file at line number 56.

set -o errexit
set -o nounset
set -o pipefail

# generate-groups generates everything for a project with external types only, e.g. a project based
# on CustomResourceDefinitions.

#Usage: $(basename $0) <generators> <output-package> <apis-package> <groups-versions> ...
#
#  <generators>        the generators comma separated to run (deepcopy,defaulter,client,lister,informer) or "all".
#  <output-package>    the output package name (e.g. github.com/example/project/pkg/generated).
#  <apis-package>      the external types dir (e.g. github.com/example/api or github.com/example/project/pkg/apis).
#  <groups-versions>   the groups and their versions in the format "groupA:v1,v2 groupB:v1 groupC:v2", relative
#                      to <api-package>.
#  ...                 arbitrary flags passed to all generator binaries.
#
#
#Examples:
#  $(basename $0) all             github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"
#  $(basename $0) deepcopy,client github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"

(
  # To support running this script from anywhere, we have to first cd into this directory
  # so we can install the tools.
  #cd $(dirname "${0}")
  go get -d k8s.io/code-generator@v0.20.2
  cd "$(go env GOPATH)/pkg/mod/k8s.io/code-generator@v0.20.2"
  go install ./cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen,conversion-gen,defaulter-gen}
)

function codegen::join() { local IFS="$1"; shift; echo "$*"; }

module_name="github.com/openebs/cstor-csi"

# Generate deepcopy functions for all internalapis and external APIs
deepcopy_inputs=(
  pkg/apis/cstor/v1
 # pkg/internalapis/apis/cstor
)

client_subpackage="pkg/client"
client_package="${module_name}/${client_subpackage}"
# Generate clientsets, listers and informers for user-facing API types
client_inputs=(
  pkg/apis/cstor/v1
)

# Generate defaulting functions to be used by the mutating webhook
defaulter_inputs=(
  pkg/apis/cstor/v1
)

# Generate conversion functions to be used by the conversion webhook
conversion_inputs=(
  pkg/apis/cstor/v1
)


gen-deepcopy() {
#  clean pkg/apis 'zz_generated.deepcopy.go'
  echo "Generating deepcopy methods..." >&2
  prefixed_inputs=( "${deepcopy_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/deepcopy-gen" \
    --go-header-file buildscripts/custom-boilerplate.go.txt \
    --input-dirs "$joined" \
    --output-file-base zz_generated.deepcopy \
    --bounding-dirs "${module_name}"
#  for dir in "${deepcopy_inputs[@]}"; do
#    copyfiles "$dir" "zz_generated.deepcopy.go"
#  done
}

gen-clientsets() {
#  clean "${client_subpackage}"/clientset '*.go'
  echo "Generating clientset..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/client-gen" \
    --go-header-file buildscripts/custom-boilerplate.go.txt \
    --clientset-name versioned \
    --input-base "" \
    --input "$joined" \
    --output-package "${client_package}"/clientset
#  copyfiles "${client_subpackage}/clientset" "*.go"
}

gen-listers() {
#  clean "${client_subpackage}/listers" '*.go'
  echo "Generating listers..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/lister-gen" \
    --go-header-file buildscripts/custom-boilerplate.go.txt \
    --input-dirs "$joined" \
    --output-package "${client_package}"/listers
#  copyfiles "${client_subpackage}/listers" "*.go"
}

gen-informers() {
#  clean "${client_subpackage}"/informers '*.go'
  echo "Generating informers..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/informer-gen" \
    --go-header-file buildscripts/custom-boilerplate.go.txt \
    --input-dirs "$joined" \
    --versioned-clientset-package "${client_package}"/clientset/versioned \
    --listers-package "${client_package}"/listers \
    --output-package "${client_package}"/informers
#  copyfiles "${client_subpackage}/informers" "*.go"
}

gen-defaulters() {
#  clean pkg/internalapis/apis 'zz_generated.defaults.go'
  echo "Generating defaulting functions..." >&2
  prefixed_inputs=( "${defaulter_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/defaulter-gen" \
    --go-header-file buildscripts/custom-boilerplate.go.txt \
    --input-dirs "$joined" \
    -O zz_generated.defaults
#  for dir in "${defaulter_inputs[@]}"; do
#    copyfiles "$dir" "zz_generated.defaults.go"
#  done
}

gen-conversions() {
#  clean pkg/internalapis/apis 'zz_generated.conversion.go'
  echo "Generating conversion functions..." >&2
  prefixed_inputs=( "${conversion_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/conversion-gen" \
    --go-header-file buildscripts/custom-boilerplate.go.txt \
    --input-dirs "$joined" \
    -O zz_generated.conversion
#  for dir in "${conversion_inputs[@]}"; do
#    copyfiles "$dir" "zz_generated.conversion.go"
#  done
}

gen-deepcopy
gen-clientsets
gen-listers
gen-informers
gen-defaulters
gen-conversions

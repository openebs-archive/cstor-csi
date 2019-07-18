# Copyright 2019 The OpenEBS Authors.
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

#!/usr/bin/env bash
# set -x
set -e

OPENEBS_OPERATOR=https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml
CSI_OPERATOR=https://raw.githubusercontent.com/openebs/csi/master/deploy/csi-operator.yaml

SRC_REPO="https://github.com/openebs/maya.git"
DST_PATH="$GOPATH/src/github.com/openebs"

# Prepare env for runnging BDD tests
# Minikube is already running
kubectl apply -f $OPENEBS_OPERATOR
kubectl apply -f $CSI_OPERATOR

# Start running BDD tests in maya for CSI
mkdir -p $DST_PATH
git clone $SRC_REPO $DST_PATH/maya
cd $DST_PATH/maya

function dumpCSINodeLogs() {
  LC=$1
  CSIPOD=$(kubectl get pods -l app=openebs-csi-node -o jsonpath='{.items[?(@.spec.containers[5].name=="openebs-csi-plugin")].metadata.name}' -n kube-system)
	kubectl logs --tail=${LC} $CSIPOD -n kube-system
  printf "\n\n"
}

function dumpCSIControllerLogs() {
  LC=$1
  CSIPOD=$(kubectl get pods -l app=openebs-csi-controller -o jsonpath='{.items[?(@.spec.containers[1].name=="openebs-csi-plugin")].metadata.name}' -n kube-system)
  kubectl logs --tail=${LC} $CSIPOD -n kube-system
  printf "\n\n"
}


# Run BDD tests for volume provisioning via CSI
cd $DST_PATH/maya/tests/csi/cstor/volume
result=`ginkgo -v -- -kubeconfig="$HOME/.kube/config" --cstor-replicas=1 --cstor-maxpools=1`

echo "BDD tests for volume provisioning via CSI "

echo "CSI Controller logs "
dumpCSIControllerLogs 20

echo "CSI Node logs"
dumpCSINodeLogs 20

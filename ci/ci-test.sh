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

#OPENEBS_OPERATOR=https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml
NDM_OPERATOR=https://raw.githubusercontent.com/openebs/cstor-operators/master/deploy/ndm-operator.yaml
CSTOR_RBAC=https://raw.githubusercontent.com/openebs/cstor-operators/master/deploy/rbac.yaml
CSTOR_OPERATOR=https://raw.githubusercontent.com/openebs/cstor-operators/master/deploy/cstor-operator.yaml
ALL_CRD=https://raw.githubusercontent.com/openebs/cstor-operators/master/deploy/crds/all_cstor_crds.yaml

CSI_OPERATOR="$GOPATH/src/github.com/openebs/cstor-csi/deploy/csi-operator.yaml"
SNAPSHOT_CLASS="$GOPATH/src/github.com/openebs/cstor-csi/deploy/snapshot-class.yaml"

DST_PATH="$GOPATH/src/github.com/openebs"

# Prepare env for runnging BDD tests
# Minikube is already running
kubectl apply -f $CSTOR_RBAC
kubectl apply -f $NDM_OPERATOR
kubectl apply -f $ALL_CRD
kubectl apply -f $CSTOR_OPERATOR
kubectl apply -f $CSI_OPERATOR
kubectl apply -f $SNAPSHOT_CLASS

function dumpCSINodeLogs() {
  LC=$1
  CSINodePOD=$(kubectl get pods -l app=openebs-csi-node -o jsonpath='{.items[0].metadata.name}' -n kube-system)
  kubectl describe po $CSINodePOD -n kube-system
  printf "\n\n"
  kubectl logs --tail=${LC} $CSINodePOD -n kube-system -c openebs-csi-plugin
  printf "\n\n"
}

function dumpCSIControllerLogs() {
  LC=$1
  CSIControllerPOD=$(kubectl get pods -l app=openebs-csi-controller -o jsonpath='{.items[0].metadata.name}' -n kube-system)
  kubectl describe po $CSIControllerPOD -n kube-system
  printf "\n\n"
  kubectl logs --tail=${LC} $CSIControllerPOD -n kube-system -c openebs-csi-plugin
  printf "\n\n"
}

# Run e2e tests for csi volumes
cd $DST_PATH/cstor-csi/tests/e2e
make e2e-test

if [ $? -ne 0 ]; then
echo "******************** CSI Controller logs***************************** "
dumpCSIControllerLogs 1000

echo "********************* CSI Node logs *********************************"
dumpCSINodeLogs 1000

#echo "******************CSI Maya-apiserver logs ********************"
#dumpMayaAPIServerLogs 1000

echo "get all the pods"
kubectl get pods --all-namespaces

echo "get pvc and pv details"
kubectl get pvc,pv --all-namespaces

echo "get cvc details"
kubectl get cvc -n openebs -oyaml

exit 1
fi

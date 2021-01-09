module github.com/openebs/cstor-csi

go 1.13

require (
	github.com/container-storage-interface/spec v1.2.0
	github.com/docker/go-units v0.4.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jpillora/go-ogle-analytics v0.0.0-20161213085824-14b04e0594ef
	github.com/kubernetes-csi/csi-lib-iscsi v0.0.0-20200118015005-959f12c91ca8
	github.com/kubernetes-csi/csi-lib-utils v0.7.0
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/openebs/api/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/exp v0.0.0-20191030013958-a1ab85dbe136 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200828081204-131dc92a58d5
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200828013309-97019fc2e64b // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/grpc v1.26.0
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.17.3
	k8s.io/gengo v0.0.0-20190826232639-a874a240740c // indirect
	k8s.io/kubernetes v1.17.3
	k8s.io/utils v0.0.0-20200124190032-861946025e34
)

replace (
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3

	k8s.io/apimachinery => k8s.io/apimachinery v0.17.4-beta.0

	k8s.io/apiserver => k8s.io/apiserver v0.17.3

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.3

	k8s.io/client-go => k8s.io/client-go v0.17.3

	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.3

	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.3

	k8s.io/code-generator => k8s.io/code-generator v0.17.4-beta.0

	k8s.io/component-base => k8s.io/component-base v0.17.3

	k8s.io/cri-api => k8s.io/cri-api v0.17.4-beta.0

	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.3

	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.3

	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.3

	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.3

	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.3

	k8s.io/kubectl => k8s.io/kubectl v0.17.3

	k8s.io/kubelet => k8s.io/kubelet v0.17.3

	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.3

	k8s.io/metrics => k8s.io/metrics v0.17.3

	k8s.io/node-api => k8s.io/node-api v0.17.3

	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.3

	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.3

	k8s.io/sample-controller => k8s.io/sample-controller v0.17.3
)

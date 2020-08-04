module github.com/openebs/cstor-csi

go 1.13

require (
	github.com/container-storage-interface/spec v1.2.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kubernetes-csi/csi-lib-iscsi v0.0.0-20200118015005-959f12c91ca8
	github.com/kubernetes-csi/csi-lib-utils v0.7.0
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	github.com/openebs/api v1.12.1-0.20200729172328-4b0764aeaaf6
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20191029190741-b9c20aec41a5 // indirect
	google.golang.org/grpc v1.26.0
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.17.3
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

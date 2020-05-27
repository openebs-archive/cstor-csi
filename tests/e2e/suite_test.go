package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
)

func TestMtest(t *testing.T) {
	//	if os.Getenv("E2ETEST") == "" {
	//	t.Skip("Run under e2e/")
	//}
	rand.Seed(time.Now().UnixNano())

	RegisterFailHandler(Fail)

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(time.Minute)

	RunSpecs(t, "Test on sanity")
}

func createNamespace(ns string) {
	stdout, stderr, err := kubectl("create", "namespace", ns)
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	Eventually(func() error {
		return waitCreatingDefaultSA(ns)
	}).Should(Succeed())
	fmt.Fprintln(os.Stderr, "created namespace: "+ns)
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func waitKindnet() error {
	stdout, stderr, err := kubectl("-n=kube-system", "get", "ds/kindnet", "-o", "json")
	if err != nil {
		return errors.New(string(stderr))
	}

	var ds appsv1.DaemonSet
	err = json.Unmarshal(stdout, &ds)
	if err != nil {
		return err
	}

	if ds.Status.NumberReady != 4 {
		return fmt.Errorf("numberReady is not 4: %d", ds.Status.NumberReady)
	}
	return nil
}

var _ = BeforeSuite(func() {
	podYAML := `apiVersion: v1
kind: Pod
metadata:
  name: ubuntu
  labels:
    app.kubernetes.io/name: ubuntu
spec:
  containers:
    - name: ubuntu
      image: prateek14/ubuntu:18.04
      command: ["/usr/local/bin/pause"]
`
	scYAML := `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: cstor-csi-sc
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: cstor.csi.openebs.io
allowVolumeExpansion: true
#volumeBindingMode: WaitForFirstConsumer
parameters:
  replicaCount: "1"
  cstorPoolCluster: "cspc-sparse"
  cas-type: "cstor"
  #cstorVolumePolicy: "csi-volume-policy"

`
	baseCSPCYaml := `apiVersion: cstor.openebs.io/v1
kind: CStorPoolCluster
metadata:
  name: cspc-sparse
  namespace: openebs
spec:
  pools:
    - nodeSelector:
        kubernetes.io/hostname: k8s
      dataRaidGroups:
      - blockDevices:
          - blockDeviceName: device
      poolConfig:
        dataRaidGroupType: "stripe"
`

	Eventually(func() error {
		bdName, nodeName, err := getBDAndNodeName()
		if err != nil {
			return err
		}

		cspcYaml := strings.Replace(baseCSPCYaml, "k8s", nodeName, 1)
		cspcYaml = strings.Replace(cspcYaml, "device", bdName, 1)

		_, stderr, err := kubectlWithInput([]byte(cspcYaml), "apply", "-f", "-")
		if err != nil {
			return errors.New(string(stderr))
		}
		return nil
	}).Should(Succeed())

	Eventually(func() error {
		_, stderr, err := kubectlWithInput([]byte(scYAML), "apply", "-f", "-")
		if err != nil {
			return errors.New(string(stderr))
		}
		return nil
	}).Should(Succeed())

	Eventually(func() error {
		_, stderr, err := kubectlWithInput([]byte(podYAML), "apply", "-f", "-")
		if err != nil {
			return errors.New(string(stderr))
		}
		return nil
	}).Should(Succeed())

	stdout, stderr, err := kubectlWithInput([]byte(podYAML), "delete", "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
})

var _ = Describe("Cstor CSI Tests", func() {
	// TODO publish tests
	//Context("publish", testPublishVolume)
	Context("e2e", testE2E)
})

var _ = AfterSuite(func() {

	Eventually(func() error {
		stdout, stderr, err := kubectl("delete", "cspc", "-n", "openebs", "cspc-sparse")
		if err != nil {
			return errors.New(string(stderr) + string(stdout))
		}
		return nil
	}).Should(Succeed())
})

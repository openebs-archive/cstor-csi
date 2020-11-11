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

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	openebsv1 "github.com/openebs/api/v2/pkg/apis/cstor/v1"
	corev1 "k8s.io/api/core/v1"
)

func testE2E() {
	testNamespacePrefix := "e2etest-"
	var ns string
	BeforeEach(func() {
		ns = testNamespacePrefix + randomString(10)
		createNamespace(ns)
	})

	AfterEach(func() {
		kubectl("delete", "namespaces/"+ns)
	})

	It("should be mounted in specified path FILESYSTEM", func() {
		By("deploying Pod with PVC")
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
      volumeMounts:
        - mountPath: /test1
          name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: cstor-pvc
`
		claimYAML := `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: cstor-csi-sc
`
		snapYAML := `apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: cstor-pvc-snap
spec:
  volumeSnapshotClassName: csi-cstor-snapshotclass
  source:
    persistentVolumeClaimName: cstor-pvc
`
		cloneYAML := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cstor-pvc-clone
spec:
  dataSource:
    name: cstor-pvc-snap
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: cstor-csi-sc
`
		clonePodYAML := `apiVersion: v1
kind: Pod
metadata:
  name: ubuntu-clone
  labels:
    app.kubernetes.io/name: ubuntu-clone
spec:
  containers:
    - name: ubuntu
      image: prateek14/ubuntu:18.04
      command: ["/usr/local/bin/pause"]
      volumeMounts:
        - mountPath: /test1
          name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: cstor-pvc-clone
`

		stdout, stderr, err := kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the correspond CStorVolume resource is created in cstorpool")
		volName := verifyVolumeCreated(ns, "cstor-pvc")

		By("confirming that the specified device exists in the Pod")
		verifyPodMounts(ns, "cstor-pvc", "ubuntu")

		By("writing file under /test1")
		writePath := "/test1/bootstrap.log"
		stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "cp", "/var/log/bootstrap.log", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "sync")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "cat", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		Expect(strings.TrimSpace(string(stdout))).ShouldNot(BeEmpty())

		By("deleting the Pod, then recreating it")
		stdout, stderr, err = kubectl("delete", "--now=true", "-n", ns, "pod/ubuntu")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the file exists")
		verifyFileExists(ns, "cstor-pvc", "ubuntu", writePath)

		By("creating snapshot for a volume")
		stdout, stderr, err = kubectlWithInput([]byte(snapYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the snapshot is created in cstorpool")
		snapshotContentName := verifySnapshotCreated(ns, "cstor-pvc-snap")

		By("creating clone volume from above created snapshot and app pod")
		stdout, stderr, err = kubectlWithInput([]byte(cloneYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(clonePodYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the correspond clone CStorVolume resource is created in cstorpool")
		cloneVolName := verifyVolumeCreated(ns, "cstor-pvc-clone")

		By("confirming that the specified device exists in the clone Pod")
		verifyPodMounts(ns, "cstor-pvc-clone", "ubuntu-clone")

		By("confirming the data integrity /test1")
		verifyFileExists(ns, "cstor-pvc-clone", "ubuntu-clone", writePath)

		By("deleting the clone Pod and PVC")
		stdout, stderr, err = kubectlWithInput([]byte(clonePodYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(cloneYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the clone PV is deleted")
		verifyPVDeleted(cloneVolName)

		By("confirming that the clone cvc correspond to PersistentVolume is deleted")
		verifyCVDeleted(cloneVolName)

		By("deleting the snapshot")
		stdout, stderr, err = kubectlWithInput([]byte(snapYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the snapshot-content deleted")
		verifySnapshotDeleted(snapshotContentName)

		By("deleting the Pod and PVC")
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the PV is deleted")
		verifyPVDeleted(volName)

		By("confirming that the cvc correspond to PersistentVolume is deleted")
		verifyCVDeleted(volName)
	})

	It("should create a block device for Pod", func() {
		deviceFile := "/dev/e2etest"

		By("deploying ubuntu Pod with PVC to mount a block device")
		podYAML := fmt.Sprintf(`apiVersion: v1
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
      volumeDevices:
        - devicePath: %s
          name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: cstor-block-pvc
`, deviceFile)
		claimYAML := `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-block-pvc
spec:
  volumeMode: Block
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: cstor-csi-sc
`
		clonePodYAML := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: ubuntu-clone
  labels:
    app.kubernetes.io/name: ubuntu
spec:
  containers:
    - name: ubuntu
      image: prateek14/ubuntu:18.04
      command: ["/usr/local/bin/pause"]
      volumeDevices:
        - devicePath: %s
          name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: cstor-block-pvc-clone
`, deviceFile)

		snapYAML := `apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: cstor-block-pvc-snap
spec:
  volumeSnapshotClassName: csi-cstor-snapshotclass
  source:
    persistentVolumeClaimName: cstor-block-pvc
`
		cloneYAML := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cstor-block-pvc-clone
spec:
  dataSource:
    name: cstor-block-pvc-snap
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  volumeMode: Block
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: cstor-csi-sc
`
		stdout, stderr, err := kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the correspond CStorVolume resource is created in cstorpool")
		volName := verifyVolumeCreated(ns, "cstor-block-pvc")

		By("confirming that a block device exists in ubuntu pod")
		verifyDevicePath(ns, "cstor-block-pvc", "ubuntu", deviceFile)

		By("writing data to a block device")
		// /etc/hostname contains "ubuntu"
		stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "dd", "if=/etc/hostname", "of="+deviceFile)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "sync")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "dd", "if="+deviceFile, "of=/dev/stdout", "bs=6", "count=1", "status=none")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		Expect(string(stdout)).Should(Equal("ubuntu"))

		By("deleting the Pod, then recreating it")
		stdout, stderr, err = kubectl("delete", "--now=true", "-n", ns, "pod/ubuntu")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("reading data from a block device")
		ReadBlockDevice(ns, "cstor-block-pvc", "ubuntu", deviceFile)

		By("creating snaphot for a volume")
		stdout, stderr, err = kubectlWithInput([]byte(snapYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the snapshot is created in cstorpool")
		snapshotContentName := verifySnapshotCreated(ns, "cstor-block-pvc-snap")

		By("creating clone volume from above created snapshot and app pod")
		stdout, stderr, err = kubectlWithInput([]byte(cloneYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(clonePodYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the correspond clone CStorVolume resource is created in cstorpool")
		cloneVolName := verifyVolumeCreated(ns, "cstor-block-pvc-clone")

		By("confirming that the block device exists in the ubuntu clone Pod")
		verifyDevicePath(ns, "cstor-block-pvc-clone", "ubuntu-clone", deviceFile)

		By("reading data from a block device")
		ReadBlockDevice(ns, "cstor-block-pvc-clone", "ubuntu-clone", deviceFile)

		By("deleting the source PVC to verify the webhook validation failure")
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).Should(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("deleting the clone Pod and PVC")
		stdout, stderr, err = kubectlWithInput([]byte(clonePodYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(cloneYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the clone PV is deleted")
		verifyPVDeleted(cloneVolName)

		By("confirming that the clone cvc correspond to PersistentVolume is deleted")
		verifyCVDeleted(cloneVolName)

		By("deleting the snapshot")
		stdout, stderr, err = kubectlWithInput([]byte(snapYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the snapshot-content deleted")
		verifySnapshotDeleted(snapshotContentName)

		By("deleting the Pod and PVC")
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the PV is deleted")
		verifyPVDeleted(volName)

		By("confirming that the cvc correspond to PersistentVolume is deleted")
		verifyCVDeleted(volName)
	})
	// ---------------------------------Resize Tests ---------------------------
	It("should resize filesystem", func() {
		currentK8sVersion := getCurrentK8sMinorVersion()
		if currentK8sVersion < 16 {
			Skip(fmt.Sprintf(
				"resizing is not supported on Kubernetes version: 1.%d. Min supported version is 1.16",
				currentK8sVersion,
			))
		}

		By("deploying Pod with PVC")
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
      volumeMounts:
        - mountPath: /test1
          name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: cstor-pvc
`
		baseClaimYAML := `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: %s
  storageClassName: cstor-csi-sc
`
		claimYAML := fmt.Sprintf(baseClaimYAML, "1Gi")
		stdout, stderr, err := kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the correspond CStorVolume resource is created in cstorpool")
		stdout, stderr, err = kubectl("get", "pvc", "-n", ns, "cstor-pvc", "-o=template", "--template={{.spec.volumeName}}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		volName := strings.TrimSpace(string(stdout))
		Eventually(func() (bool, error) {
			return checkCStorVolumeIsHealthy(volName, "openebs")
		}, 120, 10).Should(BeTrue())

		By("confirming that the specified device is mounted in the Pod")
		Eventually(func() error {
			return verifyMountExists(ns, "ubuntu", "/test1")
		}, time.Minute*2).Should(Succeed())

		By("resizing PVC online")
		claimYAML = fmt.Sprintf(baseClaimYAML, "2Gi")
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the specified device is resized in the Pod")
		timeout := time.Minute * 5
		Eventually(func() error {
			stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "df", "--output=size", "/test1")
			if err != nil {
				return fmt.Errorf("failed to get volume size. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			dfFields := strings.Fields((string(stdout)))
			volSize, err := strconv.Atoi(dfFields[1])
			if err != nil {
				return fmt.Errorf("failed to convert volume size string. stdout: %s, err: %v", stdout, err)
			}
			if volSize != 2031440 {
				return fmt.Errorf("failed to match volume size. actual: %d, expected: %d", volSize, 2031440)
			}
			return nil
		}, timeout).Should(Succeed())

		By("deleting Pod for offline resizing")
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("resizing PVC offline")
		claimYAML = fmt.Sprintf(baseClaimYAML, "3Gi")
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("deploying Pod")
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the specified device is resized in the Pod")
		Eventually(func() error {
			stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "df", "--output=size", "/test1")
			if err != nil {
				return fmt.Errorf("failed to get volume size. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			dfFields := strings.Fields((string(stdout)))
			volSize, err := strconv.Atoi(dfFields[1])
			if err != nil {
				return fmt.Errorf("failed to convert volume size string. stdout: %s, err: %v", stdout, err)
			}
			if volSize != 3063568 {
				return fmt.Errorf("failed to match volume size. actual: %d, expected: %d", volSize, 3063568)
			}
			return nil
		}, timeout).Should(Succeed())

		//	By("confirming that no failure event has occurred")
		//	fieldSelector := "involvedObject.kind=PersistentVolumeClaim," +
		//		"involvedObject.name=cstor-pvc," +
		//		"reason=VolumeResizeFailed"
		//	stdout, stderr, err = kubectl("get", "-n", ns, "events", "-o", "json", "--field-selector="+fieldSelector)
		//	Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		//	var events corev1.EventList
		//	err = json.Unmarshal(stdout, &events)
		//	Expect(err).NotTo(HaveOccurred(), "stdout=%s", stdout)
		//	Expect(events.Items).To(HaveLen(1))

		//		By("resizing PVC over pool capacity")
		//		claimYAML = fmt.Sprintf(baseClaimYAML, "10Gi")
		//		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		//		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		//
		//		By("confirming that a failure event occurs")
		//		Eventually(func() error {
		//			stdout, stderr, err = kubectl("get", "-n", ns, "events", "-o", "json", "--field-selector="+fieldSelector)
		//			if err != nil {
		//				return fmt.Errorf("failed to get event. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		//			}
		//
		//			var events corev1.EventList
		//			err = json.Unmarshal(stdout, &events)
		//			if err != nil {
		//				return fmt.Errorf("failed to unmarshal events. stdout: %s, err: %v", stdout, err)
		//			}
		//
		//			if len(events.Items) == 0 {
		//				return errors.New("failure event not found")
		//			}
		//			return nil
		//		}).Should(Succeed())

		By("deleting the Pod and PVC")
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the PV is deleted")
		verifyPVDeleted(volName)

		By("confirming that the cvc correspond to PersistentVolume is deleted")
		verifyCVDeleted(volName)

	})

	It("should a block device", func() {
		currentK8sVersion := getCurrentK8sMinorVersion()
		if currentK8sVersion < 16 {
			Skip(fmt.Sprintf(
				"resizing is not supported on Kubernetes version: 1.%d. Min supported version is 1.16",
				currentK8sVersion,
			))
		}

		By("deploying Pod with PVC")
		deviceFile := "/dev/e2etest"
		podYAML := fmt.Sprintf(`apiVersion: v1
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
      volumeDevices:
        - devicePath: %s
          name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: cstor-pvc
`, deviceFile)

		baseClaimYAML := `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-pvc
spec:
  volumeMode: Block
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: %s
  storageClassName: cstor-csi-sc
`

		claimYAML := fmt.Sprintf(baseClaimYAML, "1Gi")
		stdout, stderr, err := kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the correspond CStorVolume resource is created in cstorpool")
		stdout, stderr, err = kubectl("get", "pvc", "-n", ns, "cstor-pvc", "-o=template", "--template={{.spec.volumeName}}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		volName := strings.TrimSpace(string(stdout))
		Eventually(func() (bool, error) {
			return checkCStorVolumeIsHealthy(volName, "openebs")
		}, 120, 10).Should(BeTrue())

		By("confirming that a block device exists in ubuntu pod")
		Eventually(func() error {
			stdout, stderr, err := kubectl("get", "-n", ns, "pvc", "cstor-pvc", "--template={{.spec.volumeName}}")
			if err != nil {
				return fmt.Errorf("failed to get volume name of cstor-pvc. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "test", "-b", deviceFile)
			if err != nil {
				return fmt.Errorf("failed to test. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("resizing PVC")
		claimYAML = fmt.Sprintf(baseClaimYAML, "2Gi")
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the specified device is resized in the Pod")
		timeout := time.Minute * 5
		Eventually(func() error {
			stdout, stderr, err = kubectl("exec", "-n", ns, "ubuntu", "--", "blockdev", "--getsize64", deviceFile)
			if err != nil {
				return fmt.Errorf("failed to get volume size. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			volSize, err := strconv.Atoi(strings.TrimSpace(string(stdout)))
			if err != nil {
				return fmt.Errorf("failed to convert volume size string. stdout: %s, err: %v", stdout, err)
			}
			if volSize != 2147483648 {
				return fmt.Errorf("failed to match volume size. actual: %d, expected: %d", volSize, 2147483648)
			}
			return nil
		}, timeout).Should(Succeed())

		By("deleting the Pod and PVC")
		stdout, stderr, err = kubectlWithInput([]byte(podYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = kubectlWithInput([]byte(claimYAML), "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the PV is deleted")
		verifyPVDeleted(volName)

		By("confirming that the cv correspond to PersistentVolume is deleted")
		verifyCVDeleted(volName)
	})
}

func verifyMountExists(ns string, pod string, mount string) error {
	stdout, stderr, err := kubectl("exec", "-n", ns, pod, "--", "mountpoint", "-d", mount)
	if err != nil {
		return fmt.Errorf("failed to check mount point. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	}
	return nil
}

func verifyMountProperties(ns string, pod string, mount string, fsType string, size int) {
	By(fmt.Sprintf("verifying that %s is mounted as type %s", mount, fsType))

	stdout, stderr, err := kubectl("exec", "-n", ns, pod, "grep", mount, "/proc/mounts")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	mountFields := strings.Fields(string(stdout))
	Expect(mountFields[2]).To(Equal(fsType))

	By(fmt.Sprintf("verifying that the volume mounted at %s has the correct size", mount))
	stdout, stderr, err = kubectl("exec", "-n", ns, pod, "--", "df", "--output=size", mount)
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

	dfFields := strings.Fields(string(stdout))
	volSize, err := strconv.Atoi(dfFields[1])
	Expect(err).ShouldNot(HaveOccurred())
	Expect(volSize).To(Equal(size))
}

func waitCreatingDefaultSA(ns string) error {
	stdout, stderr, err := kubectl("get", "sa", "-n", ns, "default")
	if err != nil {
		return fmt.Errorf("default sa is not found. stdout=%s, stderr=%s, err=%v", stdout, stderr, err)
	}
	return nil
}

func waitCreatingPodWithPVC(podName, ns string) (string, error) {
	stdout, stderr, err := kubectl("get", "-n", ns, "pod", podName, "-o", "json")
	if err != nil {
		return "", fmt.Errorf("failed to create pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	}

	var pod corev1.Pod
	err = json.Unmarshal(stdout, &pod)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal pod. stdout: %s, err: %v", stdout, err)
	}

	if pod.Spec.NodeName == "" {
		return "", fmt.Errorf("pod is not yet scheduled")
	}

	return pod.Spec.NodeName, nil
}

func checkCStorVolumeIsHealthy(volName, ns string) (bool, error) {
	stdout, stderr, err := kubectl("get", "-n", ns, "cv", volName, "-o", "json")
	if err != nil {
		return false, fmt.Errorf("failed to get cv. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	}

	var cv openebsv1.CStorVolume
	err = json.Unmarshal(stdout, &cv)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal cv. stdout: %s, err: %v", stdout, err)
	}

	if cv.Status.Phase == "Healthy" {
		return true, nil
	}
	return false, fmt.Errorf("cstor volume not healthy, current status is %s:", cv.Status.Phase)
}

func getCurrentK8sMinorVersion() int64 {
	kubernetesVersionStr := os.Getenv("TEST_KUBERNETES_VERSION")
	kubernetesVersion := strings.Split(kubernetesVersionStr, ".")
	Expect(len(kubernetesVersion)).To(Equal(2))
	kubernetesMinorVersion, err := strconv.ParseInt(kubernetesVersion[1], 10, 64)
	Expect(err).ShouldNot(HaveOccurred())

	return kubernetesMinorVersion
}

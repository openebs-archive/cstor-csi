package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	//. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/api/pkg/apis/openebs.io/v1alpha1"
)

func execAtLocal(cmd string, input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if len(input) != 0 {
		command.Stdin = bytes.NewReader(input)
	}

	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func kubectl(args ...string) ([]byte, []byte, error) {
	return execAtLocal("kubectl", nil, args...)
}

func kubectlWithInput(input []byte, args ...string) ([]byte, []byte, error) {
	return execAtLocal("kubectl", input, args...)
}

func containString(s []string, target string) bool {
	for _, ss := range s {
		if ss == target {
			return true
		}
	}
	return false
}

func getBDAndNodeName() (string, string, error) {
	stdout, _, err := kubectl("get", "bd", "-n", "openebs", "-o", "json")
	if err != nil {
		return "", "", err
	}

	var bdList apis.BlockDeviceList
	err = json.Unmarshal(stdout, &bdList)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal blockdevices. stdout: %s, err: %v", stdout, err)
	}
	for _, bd := range bdList.Items {
		if string(bd.Status.ClaimState) != "Unclaimed" {
			continue
		}
		nodeName := bd.Labels[string("kubernetes.io/hostname")]
		return bd.Name, nodeName, nil
	}
	return "", "", fmt.Errorf("failed to get unclaimed blockdevice")
}

func verifyVolumeCreated(ns, pvc string) string {
	stdout, stderr, err := kubectl("get", "pvc", "-n", ns, pvc, "-o=template", "--template={{.spec.volumeName}}")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	volName := strings.TrimSpace(string(stdout))
	Eventually(func() (bool, error) {
		return checkCStorVolumeIsHealthy(volName, "openebs")
	}, 120, 10).Should(BeTrue())
	return volName
}

func verifySnapshotCreated(ns, snapName string) string {
	var snapshotContentName string
	Eventually(func() (bool, error) {
		stdout, stderr, err := kubectl("get", "volumesnapshots", "-n", ns, snapName, "-o=template", "--template={{.status.boundVolumeSnapshotContentName}}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		snapshotContentName = strings.TrimSpace(string(stdout))
		return IsSnapshotContentReady(snapshotContentName)
	}, 120, 10).Should(BeTrue())
	return snapshotContentName
}

func IsSnapshotContentReady(snapshotContentName string) (bool, error) {
	stdout, stderr, err := kubectl("get", "volumesnapshotcontents", snapshotContentName, "-o=template", "--template={{.status.readyToUse}}")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

	if strings.TrimSpace(string(stdout)) == "true" {
		return true, nil
	}
	return false, fmt.Errorf("snapshot content is not ready, current state is %s:", strings.TrimSpace(string(stdout)))
}

func verifyPodMounts(ns, pvc, podName string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "pvc", pvc, "-n", ns)
		if err != nil {
			return fmt.Errorf("failed to create PVC. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("get", "pods", podName, "-n", ns)
		if err != nil {
			return fmt.Errorf("failed to create Pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("exec", "-n", ns, podName, "--", "mountpoint", "-d", "/test1")
		if err != nil {
			return fmt.Errorf("failed to check mount point. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("exec", "-n", ns, podName, "grep", "/test1", "/proc/mounts")
		if err != nil {
			return err
		}
		fields := strings.Fields(string(stdout))
		if fields[2] != "ext4" {
			return errors.New("/test1 is not ext4")
		}
		return nil
	}, time.Minute*2).Should(Succeed())

}

func verifyDevicePath(ns, pvc, podName, deviceFile string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "-n", ns, "pvc", pvc, "--template={{.spec.volumeName}}")
		if err != nil {
			return fmt.Errorf("failed to get volume name of pvc %s. stdout: %s, stderr: %s, err: %v", pvc, stdout, stderr, err)
		}
		stdout, stderr, err = kubectl("exec", "-n", ns, podName, "--", "test", "-b", deviceFile)
		if err != nil {
			return fmt.Errorf("failed to test. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		return nil
	}).Should(Succeed())

}

func verifyPVDeleted(volName string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "pv", volName, "--ignore-not-found")
		if err != nil {
			return fmt.Errorf("failed to get pv/%s. stdout: %s, stderr: %s, err: %v", volName, stdout, stderr, err)
		}
		if len(strings.TrimSpace(string(stdout))) != 0 {
			return fmt.Errorf("target pv exists %s", volName)
		}
		return nil
	}).Should(Succeed())
}

func verifyCVDeleted(cvName string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "cv", cvName, "-n", "openebs", "--ignore-not-found")
		if err != nil {
			return fmt.Errorf("failed to get cstorvolume/%s. stdout: %s, stderr: %s, err: %v", cvName, stdout, stderr, err)
		}
		if len(strings.TrimSpace(string(stdout))) != 0 {
			return fmt.Errorf("cstorvolume exists %s", cvName)
		}
		return nil
	}).Should(Succeed())
}

func verifySnapshotDeleted(snapName string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "volumesnapshotcontent", snapName, "--ignore-not-found")
		if err != nil {
			return fmt.Errorf("failed to get cstorvolume/%s. stdout: %s, stderr: %s, err: %v", snapName, stdout, stderr, err)
		}
		if len(strings.TrimSpace(string(stdout))) != 0 {
			return fmt.Errorf("cstorvolume exists %s", snapName)
		}
		return nil
	}).Should(Succeed())
}

func ReadBlockDevice(ns, pvc, podName, deviceFile string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "pvc", pvc, "-n", ns)
		if err != nil {
			return fmt.Errorf("failed to create PVC. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("get", "pods", podName, "-n", ns)
		if err != nil {
			return fmt.Errorf("failed to create Pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("exec", "-n", ns, podName, "--", "dd", "if="+deviceFile, "of=/dev/stdout", "bs=6", "count=1", "status=none")
		if err != nil {
			return fmt.Errorf("failed to cat. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		if string(stdout) != "ubuntu" {
			return fmt.Errorf("expected: ubuntu, actual: %s", string(stdout))
		}
		return nil
	}).Should(Succeed())
}

func verifyFileExists(ns, pvc, podName, writePath string) {
	Eventually(func() error {
		stdout, stderr, err := kubectl("get", "pvc", pvc, "-n", ns)
		if err != nil {
			return fmt.Errorf("failed to create PVC. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("get", "pods", podName, "-n", ns)
		if err != nil {
			return fmt.Errorf("failed to create Pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}

		stdout, stderr, err = kubectl("exec", "-n", ns, podName, "--", "cat", writePath)
		if err != nil {
			return fmt.Errorf("failed to cat. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		if len(strings.TrimSpace(string(stdout))) == 0 {
			return fmt.Errorf(writePath + " is empty")
		}
		return nil
	}).Should(Succeed())

}

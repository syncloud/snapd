package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uthng/gossh"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestInstall(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-test.sh /snapd.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("device", fmt.Sprintf("snap install /testapp1_1_%s.snap --devmode", arch))
	assert.NoError(t, err, output)
}

func TestUnsquashfs_ListWithAttributes(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := Ssh("device", fmt.Sprintf("unsquashfs -ll /testapp1_1_%s.snap", arch))
	assert.NoError(t, err, output)
}

func InstallSnapd(cmd string) (string, error) {
	output, err := Ssh("device", cmd)
	if err != nil {
		return output, err
	}
	//output, err = SshWaitFor("device", "snap list",
	//	func(output string) bool {
	//		return strings.Contains(output, "No snaps")
	//	},
	//)
	//output, err = SshWaitFor("device", "snap find unknown",
	//	func(output string) bool {
	//		return !strings.Contains(output, "too early for operation")
	//	},
	//)
	return output, err
}

func SshWaitFor(host string, command string, predicate func(string) bool) (string, error) {
	retries := 10
	retry := 0
	for retry < retries {
		retry++
		output, err := Ssh(host, command)
		if err != nil {
			fmt.Printf("error: %v", err)
			time.Sleep(1 * time.Second)
			fmt.Printf("retry %d/%d", retry, retries)
			continue
		}
		if predicate(output) {
			return output, nil
		}
	}
	return "", fmt.Errorf("%d: %d (exhausted)", retry, retries)
}

func Ssh(host string, command string) (string, error) {
	config, err := gossh.NewClientConfigWithUserPass("root", "syncloud", host, 22, false)
	if err != nil {
		return "", err
	}

	client, err := gossh.NewClient(config)
	if err != nil {
		return "", err
	}
	fmt.Printf("%s: %s\n", host, command)
	output, err := client.ExecCommand(command)
	result := string(output)
	fmt.Printf("output: \n%s\n", result)
	return result, err
}

func snapArch() (string, error) {
	output, err := exec.Command("dpkg", "--print-architecture").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

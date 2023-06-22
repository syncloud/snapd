package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uthng/gossh"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	Cli      = "/usr/lib/syncloud-store/bin/cli"
	StoreDir = "/var/www/html"
)

func TestApps(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := SshWaitFor("device", "snap list", func(output string) bool { return strings.Contains(output, "No snaps") })
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap install unknown --channel=master")
	assert.Error(t, err)
	assert.Contains(t, output, "not found")

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)

	//#known issue unable to install local then refresh from master if there is no stable version in the store
	//#$SSH root@$DEVICE snap install /testapp2_1.snap --devmode
	//#$SSH root@$DEVICE timeout 1m snap refresh testapp2 --channel=master --amend

	output, err = Ssh("device", "snap install testapp2 --channel=master")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap list")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /testapp1_2_%s.snap -b stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 2 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap refresh testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /testapp1_3_%s.snap -b stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 3 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap refresh --list")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap refresh --list")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap find testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap find")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp2")
	assert.NoError(t, err, output)

	client := &http.Client{}
	_, err = client.Post("http://device:8080/v2/snaps/info/testapp1?architecture=arm64&fields=architectures", "", nil)
	assert.NoError(t, err, output)
}

func snapArch() (string, error) {
	output, err := exec.Command("dpkg", "--print-architecture").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func SshWaitFor(host string, command string, predicate func(string) bool) (string, error) {
	retries := 10
	retry := 0
	for retry < retries {
		output, err := Ssh(host, command)
		if err != nil {
			fmt.Printf("error: %v", err)
			time.Sleep(1 * time.Second)
			retry++
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

package ranchhand

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/cerebrotech/ranchhand/pkg/ssh"
	"github.com/pkg/errors"
)

const HostCheckTimeout = 1 * time.Minute

func enforceRequirements(hosts []string, sshKeyPath string) error {
	eChan := make(chan error)

	for _, hostname := range hosts {
		addr := fmt.Sprintf("%s:%d", hostname, 22)

		go func(addr, keyPath string, c chan<- error) {
			c <- checkHost(addr, keyPath)
		}(addr, sshKeyPath, eChan)
	}

	var errs []error
	for i := 0; i < len(hosts); i++ {
		select {
		case err := <-eChan:
			if err != nil {
				errs = append(errs, err)
			}
		case <-time.After(HostCheckTimeout):
			return errors.New("host check timeout exceeded")
		}
	}

	if len(errs) > 0 {
		return unifiedErrorF("one or more nodes failed requirement checks: %s", errs)
	}

	return nil
}

func checkHost(addr, keyPath string) error {
	client, err := ssh.Connect(addr, "vagrant", keyPath)
	if err != nil {
		return err
	}

	out, err := client.ExecuteCmd("cat /etc/os-release")
	if err != nil {
		return errors.Wrap(err, "os check failed")
	}

	kvPairs := strings.Split(out, "\n") // NOTE: this should be util func
	osInfo := make(map[string]string)
	for _, pair := range kvPairs {
		if len(pair) > 0 {
			z := strings.Split(pair, "=")
			osInfo[z[0]] = z[1]
		}
	}
	id := strings.Trim(osInfo["ID"], "\"")
	version := strings.Trim(osInfo["VERSION_ID"], "\"")

	switch id {
	case "ubuntu":
		return runVersionComparison("~16.04", version)
	case "rhel", "centos":
		return runVersionComparison("~7", version)
	default:
		return errors.Errorf("os %s is not supported", osInfo["PRETTY_NAME"])
	}

	// TODO: check hardware resources
}

func runVersionComparison(constraint, version string) error {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return err
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	if ok, errs := c.Validate(v); !ok {
		return unifiedErrorF("invalid os version: %s", errs)
	}

	return nil
}

func unifiedErrorF(format string, errs []error) error { // NOTE: util func
	if len(errs) == 0 {
		return nil
	}

	var msgs []string
	for idx, err := range errs {
		msgs = append(msgs, fmt.Sprintf("[%d: %s]", idx, err.Error()))
	}
	fullMsg := strings.Join(msgs, ", ")

	return errors.Errorf(format, fullMsg)
}
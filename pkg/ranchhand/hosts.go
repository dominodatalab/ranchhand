package ranchhand

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/dominodatalab/ranchhand/pkg/osi"
	"github.com/dominodatalab/ranchhand/pkg/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func processHosts(cfg *Config) error {
	errChan := make(chan error)

	for _, hostname := range cfg.Nodes {
		go func(hostname, sshUser, sshKeyPath string, sshPort uint, c chan<- error) {
			c <- processHost(hostname, sshUser, sshKeyPath, sshPort)
		}(hostname, cfg.SSHUser, cfg.SSHKeyPath, cfg.SSHPort, errChan)
	}

	var errs []error
	for i := 0; i < len(cfg.Nodes); i++ {
		select {
		case err := <-errChan:
			if err != nil {
				errs = append(errs, err)
			}
		case <-time.After(cfg.Timeout):
			return errors.New("host check timeout exceeded")
		}
	}

	if len(errs) > 0 {
		return unifiedErrorF("one or more nodes failed requirement checks: %s", errs)
	}

	return nil
}

func processHost(hostname, sshUser, keyPath string, sshPort uint) error {
	client := ssh.Connect(hostname, sshUser, keyPath, sshPort)

	// enforce requirements
	osrelease, _, _, err := client.ExecuteCmd("cat /etc/os-release")
	if err != nil {
		return errors.Wrap(err, "os info check failed")
	}
	osInfo := osi.Parse(osrelease)

	var vErr error
	switch osInfo.ID {
	case osi.UbuntuOS:
		vErr = runVersionComparison("~16.04", osInfo.VersionID)
	case osi.RHELOS, osi.CentOS:
		vErr = runVersionComparison("~7", osInfo.VersionID)
	default:
		vErr = errors.Errorf("os %s is not supported", osInfo.PrettyName)
	}
	if vErr != nil {
		return vErr
	}

	// todo: check hardware resources: should be 4 CPUs and 16GB RAM

	// install docker client and server
	// todo: this check is not good enough. if the provision step bombs out but docker is installed, it will never know
	// 	what was completed. we need something atomic or some touchfile
	dockerV, _, _, err := client.ExecuteCmd("sudo docker version --format '{{.Server.Version}}'")
	if err == nil { // docker is already installed
		v1, err := semver.NewVersion(strings.TrimSpace(dockerV))
		if err != nil {
			return err
		}
                return err

		v2, err := semver.NewVersion("17.03.3-ce")
		if err != nil {
			return err
		}

		if !v1.Equal(v2) {
			return errors.Errorf("invalid version of docker installed %q", dockerV)
		}
	} else {
		var cmds []string

		switch osInfo.ID {
		case osi.UbuntuOS:
			cmds = append(cmds,
				"sudo apt-get update",
				"sudo apt-get remove docker docker-engine docker.io containerd runc",
				"sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common",
				"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
				"sudo apt-key fingerprint 0EBFCD88",
				"sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\"",
				"sudo apt-get update",
				"sudo apt-get install -y docker-ce=17.03.3~ce-0~ubuntu-xenial containerd.io",
			)
		case osi.CentOS:
			cmds = append(cmds,
				"sudo yum remove docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine",
				"sudo sudo yum install -y yum-utils device-mapper-persistent-data lvm2",
				"sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo",
				"sudo yum install -y docker-ce-18.09.2 docker-ce-cli-18.09.2 containerd.io",
				"sudo systemctl start docker",
			)
		case osi.RHELOS:
			cmds = append(cmds,
				"sudo yum remove docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine",
                                "sudo subscription-manager repos --enable rhel-7-server-extras-rpms || echo 'Error enabling rhel extras repo, continuing...'",
				"sudo sudo yum install -y yum-utils device-mapper-persistent-data lvm2",
				"sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo",
				"sudo yum install -y docker-ce-18.09.2 docker-ce-cli-18.09.2 containerd.io",
				"sudo systemctl start docker",
			)
		}

		logrus.Infof("installing docker [%s] on host [%s]", "17.03.3-ce", hostname)
		cmds = append(cmds, "sudo usermod -aG docker $USER",
                                "timeout 3m /bin/bash -c 'until sudo docker version; do sleep 1; done'")
		stdout, stderr, isTimeout, err := client.ExecuteCmd(strings.Join(cmds, " && "))
		if err != nil {
                    return errors.Errorf("Host %s docker installation returned error: %s\nstdout: %s\nstderr: %s", hostname, err, stdout, stderr)
		} else if !isTimeout {
                    return errors.Errorf("Host %s docker installation timed out.\nstdout: %s\nstderr: %s", hostname, stdout, stderr)
                }
	}

	return nil
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

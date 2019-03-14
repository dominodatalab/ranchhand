package ranchhand

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/dominodatalab/ranchhand/pkg/osi"
	"github.com/dominodatalab/ranchhand/pkg/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	nodeRecommendation   = 3
	cpuRecommendation    = 4
	memoryRecommendation = 16.0
)

var (
	versionConstraints = map[string]string{
		"ubuntu": "~16.04.x",
		"centos": "~7.5.x",
		"rhel":   "~7.5.x",
		"docker": "~18.06.x-ce",
	}

	dockerInstallCmds = map[string][]string{
		"ubuntu": {
			"sudo apt-get update",
			"sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common",
			"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
			"sudo apt-key fingerprint 0EBFCD88",
			"sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\"",
			"sudo apt-get update",
			"sudo apt-get install -y docker-ce=18.06.3~ce~3-0~ubuntu containerd.io",
		},
		"centos": {
			"sudo yum install -y yum-utils device-mapper-persistent-data lvm2",
			"sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo",
			"sudo yum install -y docker-ce-18.06.3.ce-3.el7 containerd.io",
			"sudo systemctl enable docker",
			"sudo systemctl start docker",
		},
		"rhel": {
			"sudo subscription-manager repos --enable rhel-7-server-extras-rpms || echo 'Error enabling rhel extras repo, continuing...'",
			"sudo yum install -y yum-utils device-mapper-persistent-data lvm2",
			"sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo",
			"sudo yum install -y docker-ce-18.09.2 docker-ce-cli-18.09.2 containerd.io",
			"sudo systemctl enable docker",
			"sudo systemctl start docker",
		},
	}
)

// process remote hosts concurrently and return any errors that occurred
func processHosts(cfg *Config) error {
	nodeCount := len(cfg.Nodes)
	if nodeCount < nodeRecommendation {
		log.Warnf("node count [%d] is less than recommended value [%d]", nodeCount, nodeRecommendation)
	}

	errChan := make(chan error)
	for _, node := range cfg.Nodes {
		routine := func(addr string, port uint, user, keyPath string, c chan<- error) {
			c <- processHost(addr, port, user, keyPath)
		}

		go routine(node.PublicIP, cfg.SSHPort, cfg.SSHUser, cfg.SSHKeyPath, errChan)
	}

	var errs []error
	for i := 0; i < nodeCount; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				errs = append(errs, err)
			}
		case <-time.After(cfg.Timeout):
			return errors.New("host check timeout exceeded")
		}
	}
	if errs != nil {
		return errors.Errorf("failed to process one or more nodes: %v", errs)
	}

	return nil
}

// connect to the host, enforce node requirements, and install docker onto a vm
func processHost(addr string, port uint, username, keyPath string) error {
	var osInfo *osi.Info
	var client *ssh.Client

	err := retry.Retry(func(attempt uint) (err error) {
		client, err = ssh.Connect(addr, port, username, keyPath)

		if err != nil {
			log.Warnf("ssh connect failed on host [%s], trying again in 5 secs", addr)
		}
		return err
	}, strategy.Limit(6), strategy.Wait(5*time.Second))

	if err == nil {
		osInfo, err = loadOSInfo(client)
	}
	if err == nil {
		err = enforceSysRequirements(client, osInfo)
	}
	if err == nil {
		err = installDocker(client, osInfo)
	}

	return errors.Wrapf(err, addr)
}

// fetch and parse os identification data
func loadOSInfo(client *ssh.Client) (*osi.Info, error) {
	contents, err := client.ExecuteCmd("cat /etc/os-release")
	if err != nil {
		return nil, errors.Wrap(err, "os info check failed")
	}

	return osi.Parse(contents), nil
}

// coordinate system checks
func enforceSysRequirements(client *ssh.Client, osInfo *osi.Info) error {
	var errs []error

	if err := constrainOS(osInfo); err != nil {
		errs = append(errs, err)
	}
	if err := constrainCPU(client); err != nil {
		errs = append(errs, err)
	}
	if err := constrainMemory(client); err != nil {
		errs = append(errs, err)
	}
	if err := constrainDockerVersion(client); err != nil {
		errs = append(errs, err)
	}

	if errs != nil {
		return errors.Errorf("system checks failed: %v", errs)
	}
	return nil
}

// install docker onto a new system and mark the operation as complete thereafter
func installDocker(client *ssh.Client, osInfo *osi.Info) error {
	remoteStateDir := "/var/lib/ranchhand"
	if _, cerr := client.ExecuteCmd(fmt.Sprintf("test -d %s", remoteStateDir)); cerr != nil {
		if _, err := client.ExecuteCmd(fmt.Sprintf("sudo mkdir -p %s", remoteStateDir)); err != nil {
			return errors.Wrapf(err, "cannot create remote state directory %q", remoteStateDir)
		}
	}

	indicator := filepath.Join(remoteStateDir, "docker-installed")
	if _, err := client.ExecuteCmd(fmt.Sprintf("test -f %s", indicator)); err == nil {
		return nil
	}

	cmds := append(dockerInstallCmds[osInfo.ID], "sudo usermod -aG docker $USER")
	if _, err := client.ExecuteCmd(strings.Join(cmds, " && ")); err != nil {
		return errors.Wrap(err, "docker install failed")
	}

	err := retry.Retry(
		func(attempt uint) (err error) {
			if _, err = client.ExecuteCmd("sudo docker version"); err != nil {
				log.Warnf("attempt [%d] to verify docker is running failed on host [%s]", attempt, client.RemoteAddr())
			}
			return err
		},
		strategy.Wait(10*time.Second),
		strategy.Limit(12),
	)
	if err != nil {
		return errors.Wrap(err, "unable to verify docker install")
	}

	_, err = client.ExecuteCmd(fmt.Sprintf("sudo touch %s", indicator))
	return errors.Wrap(err, "cannot mark docker install complete")
}

// ensure operating system is compatible
func constrainOS(osInfo *osi.Info) (err error) {
	if osInfo.IsUbuntu() || osInfo.IsCentOS() || osInfo.IsRHEL() {
		err = constrainVersion(versionConstraints[osInfo.ID], osInfo.VersionID)
	} else {
		err = errors.New("support not implemented")
	}

	return errors.Wrapf(err, "invalid operating system [%s]", osInfo.PrettyName)
}

// ensure total cpu count meets requirements, issue warning otherwise
func constrainCPU(client *ssh.Client) error {
	cpuCheckCmd := "grep -c ^processor /proc/cpuinfo"

	countStr, err := client.ExecuteCmd(cpuCheckCmd)
	if err != nil {
		return errors.Wrap(err, "cpu count failed")
	}
	cpuCount, err := strconv.Atoi(countStr)
	if err != nil {
		return errors.Wrapf(err, "cpu check [%s] results unprocessable", cpuCheckCmd)
	}

	if cpuCount < cpuRecommendation {
		log.Warnf("cpu count [%d] is less than recommended value [%d] on host [%s]",
			cpuCount, cpuRecommendation, client.RemoteAddr())
	}
	return nil
}

// ensure total memory meets requirements, issue warning otherwise
func constrainMemory(client *ssh.Client) error {
	memCheckCmd := "awk '/MemTotal/ {print $2 / 1024^2}' /proc/meminfo"

	memStr, err := client.ExecuteCmd(memCheckCmd)
	if err != nil {
		return errors.Wrap(err, "memory check failed")
	}
	memSize, err := strconv.ParseFloat(memStr, 2)
	if err != nil {
		return errors.Wrapf(err, "memory check [%s] results unprocessable", memCheckCmd)
	}

	if memSize < memoryRecommendation {
		log.Warnf("memory size [%.1fGB] is less than recommended value [%.1fGB] on host [%s]",
			memSize, memoryRecommendation, client.RemoteAddr())
	}
	return nil
}

// ensure installed docker version is compatible
func constrainDockerVersion(client *ssh.Client) error {
	version, err := client.ExecuteCmd("docker version --format '{{.Server.Version}}'")
	if err == nil {
		err = constrainVersion(versionConstraints["docker"], version)
		return errors.Wrap(err, "invalid docker installed")
	}

	return nil
}

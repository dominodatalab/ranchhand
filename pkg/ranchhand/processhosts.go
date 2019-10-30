package ranchhand

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/dominodatalab/os-release"
	"github.com/dominodatalab/ranchhand/pkg/ssh"
	"github.com/pkg/errors"
)

const (
	nodeRecommendation   = 3
	cpuRecommendation    = 4
	memoryRecommendation = 16.0

	k8sCfgDir      = "/etc/kubernetes"
	remoteStateDir = "/var/lib/ranchhand"
)

var (
	k8sConfigs map[string]k8sConfig

	versionConstraints = map[string]string{
		"ubuntu": ">=16.04.x",
		"centos": "~7.x",
		"rhel":   "~7.x",
		"docker": "~18.09.x",
	}

	dockerInstallCmds = map[string][]string{
		"ubuntu": {
			"sudo apt-get update",
			"sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common",
			"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
			"sudo apt-key fingerprint 0EBFCD88",
			"sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\"",
			"sudo apt-get update",
			"sudo apt-get install -y docker-ce=5:18.09.6~3-0~* docker-ce-cli=5:18.09.6~3-0~* containerd.io",
		},
		"centos": {
			"sudo yum install -y yum-utils device-mapper-persistent-data lvm2",
			"sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo",
			"sudo yum install -y docker-ce-18.09.6-3.el7 docker-ce-cli-18.09.6-3.el7 containerd.io",
			"sudo systemctl enable docker",
			"sudo systemctl start docker",
		},
		"rhel": {
			"sudo yum install -y yum-utils device-mapper-persistent-data lvm2",
			"export REPO_ROOT=\"https://mirrors.domino.tech/artifacts/docker\"",
			"sudo yum install -y $REPO_ROOT/docker-ce-18.09.6-3.el7.x86_64.rpm $REPO_ROOT/docker-ce-cli-18.09.6-3.el7.x86_64.rpm $REPO_ROOT/containerd.io-1.2.10-3.2.el7.x86_64.rpm $REPO_ROOT/container-selinux-2.107-3.el7.noarch.rpm",
			"sudo systemctl enable docker",
			"sudo systemctl start docker",
		},
	}
)

type k8sConfig struct {
	filename, contents string
}

// process remote hosts concurrently and return any errors that occurred
func processHosts(cfg *Config) error {
	nodeCount := len(cfg.Nodes)
	if nodeCount < nodeRecommendation {
		log.Warnf("node count [%d] is less than recommended value [%d]", nodeCount, nodeRecommendation)
	}

	errChan := make(chan error)
	for _, node := range cfg.Nodes {
		routine := func(addr string, sshCfg *SSHConfig, c chan<- error) {
			c <- processHost(addr, sshCfg)
		}

		go routine(node.PublicIP, cfg.SSH, errChan)
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
func processHost(addr string, cfg *SSHConfig) error {
	var osInfo *osrelease.Data
	var client *ssh.Client

	err := dialHost(addr, cfg.Port, cfg.ConnectionTimeout)
	if err == nil {
		client, err = ssh.Connect(addr, cfg.Port, cfg.User, cfg.KeyPath)
	}
	if err == nil {
		osInfo, err = loadOSInfo(client)
	}
	if err == nil {
		err = enforceSysRequirements(client, osInfo)
	}
	if err == nil {
		err = installDocker(client, osInfo)
	}
	if err == nil {
		err = installK8sConfigs(client)
	}

	return errors.Wrapf(err, addr)
}

// attempt to verify that a host is listening at port until timeout
func dialHost(addr string, port, timeout uint) error {
	waitTime := 5 * time.Second
	attempts := timeout / uint(waitTime.Seconds())
	fullAddr := fmt.Sprintf("%s:%d", addr, port)

	return retry.Retry(func(attempt uint) error {
		conn, err := net.Dial("tcp", fullAddr)
		if err != nil {
			log.Warnf("attempt [%d] to verify host [%s] is listening failed", attempt, fullAddr)
			return err
		}

		return conn.Close()
	}, strategy.Wait(waitTime), strategy.Limit(attempts))
}

// fetch and parse os identification data
func loadOSInfo(client *ssh.Client) (*osrelease.Data, error) {
	contents, err := client.ExecuteCmd("cat /etc/os-release")
	if err != nil {
		return nil, errors.Wrap(err, "os info check failed")
	}

	return osrelease.Parse(contents), nil
}

// coordinate system checks
func enforceSysRequirements(client *ssh.Client, osInfo *osrelease.Data) error {
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
func installDocker(client *ssh.Client, osInfo *osrelease.Data) error {
	if err := ensureRemoteDirectory(client, remoteStateDir); err != nil {
		return errors.Wrap(err, "cannot create remote state dir")
	}

	indicator := filepath.Join(remoteStateDir, "docker-installed")
	if _, err := client.ExecuteCmd(fmt.Sprintf("test -f %s", indicator)); err == nil {
		return nil
	}

	cmds := append(dockerInstallCmds[osInfo.ID], "sudo usermod -aG docker $USER")
	log.Infof("installing docker on host [%s]", client.RemoteAddr())
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
func constrainOS(osInfo *osrelease.Data) (err error) {
	if osInfo.IsLikeDebian() || osInfo.IsLikeFedora() {
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

// install static k8s config files
func installK8sConfigs(client *ssh.Client) error {
	log.Info("creating kubernetes host configs")

	if err := ensureRemoteDirectory(client, k8sCfgDir); err != nil {
		return errors.Wrap(err, "cannot create k8s config dir")
	}
	for _, cfg := range k8sConfigs {
		cmdTmpl := "test -f %[1]q || echo -e %[2]q | sudo tee %[1]s && sudo chown root:root %[1]s && sudo chmod 0600 %[1]s"
		if _, err := client.ExecuteCmd(fmt.Sprintf(cmdTmpl, cfg.filename, cfg.contents)); err != nil {
			return errors.Wrap(err, "cannot create k8s config")
		}
	}
	return nil
}

func init() {
	k8sAuditCfgFile := filepath.Join(k8sCfgDir, "audit.yaml")
	k8sEventRateCfgFile := filepath.Join(k8sCfgDir, "event.yaml")
	k8sAdmissionCfgFile := filepath.Join(k8sCfgDir, "admission.yaml")

	k8sAuditCfgContents := `apiVersion: audit.k8s.io/v1beta1
kind: Policy
rules:
- level: Metadata`
	k8sEventRateCfgContents := `apiVersion: eventratelimit.admission.k8s.io/v1alpha1
kind: Configuration
limits:
- type: Server
  qps: 500
  burst: 5000`
	k8sAdmissionCfgContents := fmt.Sprintf(`apiVersion: apiserver.k8s.io/v1alpha1
kind: AdmissionConfiguration
plugins:
- name: EventRateLimit
  path: %s`, k8sEventRateCfgFile)

	k8sConfigs = map[string]k8sConfig{
		"admission": {
			filename: k8sAdmissionCfgFile,
			contents: k8sAdmissionCfgContents,
		},
		"audit": {
			filename: k8sAuditCfgFile,
			contents: k8sAuditCfgContents,
		},
		"event": {
			filename: k8sEventRateCfgFile,
			contents: k8sEventRateCfgContents,
		},
	}
}

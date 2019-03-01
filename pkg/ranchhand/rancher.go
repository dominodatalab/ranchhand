package ranchhand

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func installRancher() error {
	helmHome, err := filepath.Abs(".helm")
	if err != nil {
		return err
	}
	commonArgs := []string{"--kubeconfig", KubeConfig, "--home", helmHome}

	// add rancher repo
	args := []string{"repo", "list"}
	buffer, err := exec.Command("helm", append(args, commonArgs...)...).Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(buffer), "rancher-stable") {
		args := []string{"repo", "add", "rancher-stable", "https://releases.rancher.com/server-charts/stable"}
		if err := exec.Command("helm", append(args, commonArgs...)...).Run(); err != nil {
			return err
		}
	}

	// install cert-manager chart
	bs, err := exec.Command("helm", append([]string{"list", "cert-manager"}, commonArgs...)...).Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(bs), "cert-manager") {
		args := []string{
			"install",
			"stable/cert-manager",
			"--name=cert-manager",
			"--namespace=kube-system",
			"--description='Installed by RanchHand'",
			"--version=v0.5.2",
			"--wait",
		}
		if err := exec.Command("helm", append(args, commonArgs...)...).Run(); err != nil {
			return err
		}
	}

	// todo: kubectl -n kube-system rollout status deploy/cert-manager

	// install rancher chart
	bs2, err := exec.Command("helm", append([]string{"list", "rancher"}, commonArgs...)...).Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(bs2), "rancher") {
		args := []string{
			"install",
			"rancher-stable/rancher",
			"--name=rancher",
			"--namespace=cattle-system",
			"--description='Installed by RanchHand'",
			"--version=2019.1.2",
			"--set", "rancherImageTag=v2.1.6,tls=external",
			"--wait",
		}
		if buffer, err := exec.Command("helm", append(args, commonArgs...)...).CombinedOutput(); err != nil {
			return errors.Wrapf(err, "helm install failed: %s", string(buffer))
		}
	}

	// todo: kubectl -n cattle-system rollout status deploy/rancher

	return nil
}

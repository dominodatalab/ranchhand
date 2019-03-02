package ranchhand

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var rancherDefaultCredentials = map[string]string{
	"username": "admin",
	"password": "admin",
}

func installRancher(nodeIP string) error {
	helmHome, err := filepath.Abs(".helm")
	if err != nil {
		return err
	}
	commonArgs := []string{"--kubeconfig", KubeConfig, "--home", helmHome}

	// add rancher repo
	args := []string{"repo", "list"}
	buf1, err := exec.Command("helm", append(args, commonArgs...)...).Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(buf1), "rancher-stable") {
		args := []string{"repo", "add", "rancher-stable", "https://releases.rancher.com/server-charts/stable"}
		if err := exec.Command("helm", append(args, commonArgs...)...).Run(); err != nil {
			return err
		}
	}

	// install cert-manager chart
	buf2, err := exec.Command("helm", append([]string{"list", "cert-manager"}, commonArgs...)...).Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(buf2), "cert-manager") {
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
	buf3, err := exec.Command("helm", append([]string{"list", "rancher"}, commonArgs...)...).Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(buf3), "rancher") {
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

	loginURL, err := url.Parse(fmt.Sprintf("https://%s/v3-public/localProviders/local?action=login", nodeIP))
	if err != nil {
		return err
	}
	body, err := json.Marshal(rancherDefaultCredentials)
	if err != nil {
		return err
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Post(loginURL.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("rancher api check failed with status (%d)", resp.StatusCode)
	}

	return nil
}

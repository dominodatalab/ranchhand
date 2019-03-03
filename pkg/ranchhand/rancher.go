package ranchhand

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cerebrotech/ranchhand/pkg/helm"
	"github.com/pkg/errors"
)

var (
	rancherRepo = helm.Repository{
		Name: "rancher-stable",
		URL:  "https://releases.rancher.com/server-charts/stable",
	}

	rancherReleases = []struct {
		Chart string
		Info  helm.ReleaseInfo
	}{
		{
			"stable/cert-manager",
			helm.ReleaseInfo{
				Name:      "cert-manager",
				Namespace: "kube-system",
				Version:   "v0.5.2",
			},
		},
		{
			"rancher-stable/rancher",
			helm.ReleaseInfo{
				Name:      "rancher",
				Namespace: "cattle-system",
				Version:   "2019.1.2",
				SetValues: "rancherImageTag=v2.1.6,tls=external",
			},
		},
	}

	rancherDefaultCredentials = map[string]string{
		"username": "admin",
		"password": "admin",
	}
)

func installRancher(nodeIP string) error {
	helmCLI, err := helm.New(".helm", KubeConfig)
	if err != nil {
		return err
	}

	exists, err := helmCLI.IsRepo(rancherRepo.Name)
	if err != nil {
		return err
	}
	if !exists {
		if err := helmCLI.AddRepo(&rancherRepo); err != nil {
			return err
		}
	}

	for _, rls := range rancherReleases {
		rlsInfo := rls.Info

		installed, err := helmCLI.IsRelease(rlsInfo.Name)
		if err != nil {
			return err
		}
		if !installed {
			rlsInfo.Description = "Installed by RanchHand"
			rlsInfo.Wait = true

			if err := helmCLI.InstallRelease(rls.Chart, &rlsInfo); err != nil {
				return err
			}
		}
	}

	return pingRancherAPI(nodeIP)
}

func pingRancherAPI(host string) error {
	loginURL, err := url.Parse(fmt.Sprintf("https://%s/v3-public/localProviders/local?action=login", host))
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

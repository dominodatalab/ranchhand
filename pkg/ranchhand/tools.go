package ranchhand

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	PlatformToolVersions = map[string]string{
		"kubectl": "v1.14.3",
		"helm":    "v2.14.1",
		"rke":     "v0.2.4",
	}

	PlatformToolURLs = map[string]RequiredToolURLs{
		"darwin": {
			Kubectl: fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/darwin/amd64/kubectl", PlatformToolVersions["kubectl"]),
			Helm:    fmt.Sprintf("https://storage.googleapis.com/kubernetes-helm/helm-%s-darwin-amd64.tar.gz", PlatformToolVersions["helm"]),
			RKE:     fmt.Sprintf("https://github.com/rancher/rke/releases/download/%s/rke_darwin-amd64", PlatformToolVersions["rke"]),
		},
		"linux": {
			Kubectl: fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/amd64/kubectl", PlatformToolVersions["kubectl"]),
			Helm:    fmt.Sprintf("https://storage.googleapis.com/kubernetes-helm/helm-%s-linux-amd64.tar.gz", PlatformToolVersions["helm"]),
			RKE:     fmt.Sprintf("https://github.com/rancher/rke/releases/download/%s/rke_linux-amd64", PlatformToolVersions["rke"]),
		},
	}
)

type RequiredToolURLs struct {
	Kubectl string
	Helm    string
	RKE     string
}

func installRequiredTools() error {
	toolsDir, err := filepath.Abs("tools")
	if err != nil {
		return err
	}
	if err := ensureDirectory(toolsDir); err != nil {
		return errors.Errorf("cannot create tools dir: %s", toolsDir)
	}

	rawBinInstall := func(binPath, url string) error {
		if err := downloadFile(binPath, url); err != nil {
			return err
		}

		if err := os.Chmod(binPath, 0755); err != nil {
			return err
		}

		return nil
	}

	urls := PlatformToolURLs[runtime.GOOS]
	allTools := []struct {
		url         string
		binary      string
		installFunc func(binPath, url string) error
	}{
		{
			url:         urls.Kubectl,
			binary:      "kubectl",
			installFunc: rawBinInstall,
		},
		{
			url:         urls.RKE,
			binary:      "rke",
			installFunc: rawBinInstall,
		},
		{
			url:    urls.Helm,
			binary: "helm",
			installFunc: func(binPath, url string) error {
				dir, err := ioutil.TempDir("", "ranchhand")
				if err != nil {
					return err
				}
				defer os.RemoveAll(dir)

				archive := filepath.Join(dir, "helm.tgz")
				if err := downloadFile(archive, url); err != nil {
					return err
				}
				if err := archiver.Unarchive(archive, dir); err != nil {
					return err
				}

				archiveBin := filepath.Join(dir, fmt.Sprintf("%s-amd64", runtime.GOOS), "helm")
				if err := os.Rename(archiveBin, binPath); err != nil {
					return err
				}

				return nil
			},
		},
	}
	for _, t := range allTools {
		binPath := filepath.Join(toolsDir, t.binary)

		if _, serr := os.Stat(binPath); os.IsNotExist(serr) {
			logrus.Infof("downloading tool [%s]", t.binary)
			if err := t.installFunc(binPath, t.url); err != nil {
				return err
			}
		}
	}

	return os.Setenv("PATH", fmt.Sprintf("%s:%s", toolsDir, os.Getenv("PATH")))
}

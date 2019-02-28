package ranchhand

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type RequiredToolURLs struct {
	Kubectl string
	Helm    string
	RKE     string
}

var PlatformToolURLs = map[string]RequiredToolURLs{
	"darwin": {
		Kubectl: "https://storage.googleapis.com/kubernetes-release/release/v1.13.3/bin/darwin/amd64/kubectl",
		Helm:    "https://storage.googleapis.com/kubernetes-helm/helm-v2.12.3-darwin-amd64.tar.gz",
		RKE:     "https://github.com/rancher/rke/releases/download/v0.1.16/rke_darwin-amd64",
	},
	"linux": {
		Kubectl: "https://storage.googleapis.com/kubernetes-release/release/v1.13.3/bin/linux/amd64/kubectl",
		Helm:    "https://storage.googleapis.com/kubernetes-helm/helm-v2.12.3-linux-amd64.tar.gz",
		RKE:     "https://github.com/rancher/rke/releases/download/v0.1.16/rke_linux-amd64",
	},
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
			if err := t.installFunc(binPath, t.url); err != nil {
				return err
			}
		}
	}

	return os.Setenv("PATH", fmt.Sprintf("%s:%s", toolsDir, os.Getenv("PATH")))
}

func downloadFile(filepath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

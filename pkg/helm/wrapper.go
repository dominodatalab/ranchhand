package helm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	TillerNamespace      = "kube-system"
	TillerServiceAccount = "tiller"
	RancherNamespace     = "cattle-system"
	RancherSecret        = "tls-ca"
	RancherCAFilename    = "cacerts.pem"
)

type Helm interface {
	Init(cacert []byte) error
	IsRepo(repoName string) (bool, error)
	IsRelease(releaseName string) (bool, error)
	AddRepo(repoData *Repository) error
	InstallRelease(releaseName string, releaseInfo *ReleaseInfo) error
}

type Repository struct {
	Name string
	URL  string
}

type ReleaseInfo struct {
	Name        string
	Namespace   string
	Description string
	Version     string
	Wait        bool
	SetValues   map[string]string
}

type wrapper struct {
	helmHome   string
	kubeConfig string
}

func New(home, kubeconfig string) (*wrapper, error) {
	helmHome, err := filepath.Abs(home)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, errors.Wrap(err, "kubeconfig is required for helm operations")
	}

	return &wrapper{helmHome: helmHome, kubeConfig: kubeconfig}, nil
}

func (w *wrapper) Init(cacert []byte) error {
	// checking if tiller is already installed
	if err := w.helmCommand("version", "--server").Run(); err != nil {
		if err := w.createK8sResources(cacert); err != nil {
			return err
		}

		buffer, err := w.helmCommand("init", "--wait", "--service-account", TillerServiceAccount).CombinedOutput()
		if err != nil {
			output := string(buffer)
			return errors.Wrapf(err, "helm init failed: %s", output)
		}
	}

	return nil
}

func (w *wrapper) IsRepo(name string) (bool, error) {
	buffer, err := w.helmCommand("repo", "list").CombinedOutput()
	output := string(buffer)

	if err != nil {
		return false, errors.Wrapf(err, "helm list repos failed: %s", output)
	}
	return strings.Contains(output, name), nil
}

func (w *wrapper) IsRelease(name string) (bool, error) {
	buffer, err := w.helmCommand("list", name).CombinedOutput()
	output := string(buffer)

	if err != nil {
		return false, errors.Wrapf(err, "helm list release (%s) failed: %s", name, output)
	}
	return strings.Contains(output, name), nil
}

func (w *wrapper) AddRepo(r *Repository) error {
	buffer, err := w.helmCommand("repo", "add", r.Name, r.URL).CombinedOutput()
	if err != nil {
		output := string(buffer)
		return errors.Wrapf(err, "helm repo add (%s) failed: %s", r.URL, output)
	}

	return nil
}

func (w *wrapper) InstallRelease(chstr string, ri *ReleaseInfo) error {
	// todo: could probably use tags and reflection here instead of if..if..if..
	// https://medium.com/golangspec/tags-in-golang-3e5db0b8ef3e
	//t := reflect.TypeOf(ri)
	var flags []string
	if ri.Name != "" {
		flags = append(flags, fmt.Sprintf("--name=%s", ri.Name))
	}
	if ri.Namespace != "" {
		flags = append(flags, fmt.Sprintf("--namespace=%s", ri.Namespace))
	}
	if ri.Description != "" {
		flags = append(flags, fmt.Sprintf("--description=%q", ri.Description))
	}
	if ri.Version != "" {
		flags = append(flags, fmt.Sprintf("--version=%s", ri.Version))
	}
	if ri.Wait {
		flags = append(flags, "--wait")
	}
	for key, value := range ri.SetValues {
		flags = append(flags, "--set", fmt.Sprintf("%s=%s", key, value))
	}
	args := append([]string{"install", chstr}, flags...)

	buffer, err := w.helmCommand(args...).CombinedOutput()
	output := string(buffer)

	return errors.Wrapf(err, "helm install release (%s) failed: %s", chstr, output)
}

func (w *wrapper) helmCommand(args ...string) *exec.Cmd {
	commonArgs := []string{
		"--home", w.helmHome,
		"--kubeconfig", w.kubeConfig,
	}
	return exec.Command("helm", append(args, commonArgs...)...)
}

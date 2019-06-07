package ranchhand

import (
	"github.com/dominodatalab/ranchhand/pkg/helm"
	"github.com/dominodatalab/ranchhand/pkg/rancher"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	rancherNamespace  = "cattle-system"
	rancherSecret     = "tls-ca"
	rancherCAFilename = "cacerts.pem"
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
				Namespace: rancherNamespace,
				Version:   "2.2.4",
				SetValues: map[string]string{
					"tls":            "external",
					"privateCA":      "true",
					"addLocal":       "false",
					"auditLog.level": "1",
				},
			},
		},
	}

	rancherDefaultCredentials = rancher.LoginCredentials{
		Username: "admin",
		Password: "admin",
	}
)

func createRancherSecret(certPEM []byte, kubeConfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	getOpts := metav1.GetOptions{}

	nsAPI := clientset.CoreV1().Namespaces()
	if ns, nsErr := nsAPI.Get(rancherNamespace, getOpts); nsErr != nil && apierrors.IsNotFound(nsErr) {
		ns.Name = rancherNamespace
		ns, err = nsAPI.Create(ns)
		if err != nil {
			return errors.Wrapf(err, "failed to create %s namespace", rancherNamespace)
		}
	}

	sAPI := clientset.CoreV1().Secrets(rancherNamespace)
	if s, sErr := sAPI.Get(rancherSecret, getOpts); sErr != nil && apierrors.IsNotFound(sErr) {
		s.Name = rancherSecret
		s.StringData = map[string]string{rancherCAFilename: string(certPEM)}
		s, err = sAPI.Create(s)
		if err != nil {
			return errors.Wrapf(err, "failed to create rancher private ca secret %s", rancherSecret)
		}
	}
	return nil
}

func installRancher(h helm.Helm, host string) error {
	exists, err := h.IsRepo(rancherRepo.Name)
	if err != nil {
		return err
	}
	if !exists {
		if err := h.AddRepo(&rancherRepo); err != nil {
			return err
		}
	}

	for _, rls := range rancherReleases {
		rlsInfo := rls.Info

		installed, err := h.IsRelease(rlsInfo.Name)
		if err != nil {
			return err
		}
		if !installed {
			rlsInfo.Description = "Installed by RanchHand"
			rlsInfo.Wait = true

			if err := h.InstallRelease(rls.Chart, &rlsInfo); err != nil {
				return err
			}
		}
	}
	return rancher.Ping(host)
}

func modifyRancherAdminPassword(host, password string) error {
	token, err := rancher.Login(host, &rancherDefaultCredentials)
	if err != nil {
		if rancher.IsUnauthorized(err) {
			log.Info("default rancher admin password has already been changed, nothing to do")
			return nil
		}
		return err
	}
	if len(password) == 0 {
		log.Warn("default rancher admin password should be changed for security reasons")
		return nil
	}

	log.Info("attempting to change rancher admin password")
	input := &rancher.ChangePasswordInput{
		CurrentPassword: rancherDefaultCredentials.Password,
		NewPassword:     password,
	}
	if err := rancher.ChangePassword(host, token, input); err != nil {
		return err
	}

	log.Info("changed rancher admin password")
	return nil
}

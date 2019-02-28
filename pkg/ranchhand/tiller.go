package ranchhand

import (
	"os/exec"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubeConfig           = "kube_config_rancher-cluster.yml"
	TillerNamespace      = "kube-system"
	TillerServiceAccount = "tiller"
)

func installTiller() error {
	config, err := clientcmd.BuildConfigFromFlags("", KubeConfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	getOpts := metav1.GetOptions{}

	saAPI := clientset.CoreV1().ServiceAccounts(TillerNamespace)
	sa, saErr := saAPI.Get(TillerServiceAccount, getOpts)
	if saErr != nil && apierrors.IsNotFound(saErr) {
		sa.Name = TillerServiceAccount

		sa, err = saAPI.Create(sa)
		if err != nil {
			return err
		}
	}

	crbAPI := clientset.RbacV1().ClusterRoleBindings()
	if crb, crbErr := crbAPI.Get(TillerServiceAccount, getOpts); crbErr != nil && apierrors.IsNotFound(crbErr) {
		crAPI := clientset.RbacV1().ClusterRoles()
		cr, err := crAPI.Get("cluster-admin", getOpts)
		if err != nil && apierrors.IsNotFound(err) {
			return err
		}

		crb.Name = sa.Name
		crb.RoleRef = rbacv1.RoleRef{
			Kind: cr.Kind,
			Name: cr.Name,
		}
		crb.Subjects = []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		}

		if _, err := crbAPI.Create(crb); err != nil {
			return err
		}
	}

	checkCmd := exec.Command("helm", "version", "--kubeconfig", KubeConfig, "--server")
	if cErr := checkCmd.Run(); cErr != nil {
		initCmd := exec.Command("helm", "init", "--kubeconfig", KubeConfig, "--service-account", TillerServiceAccount, "--wait")
		if iErr := initCmd.Run(); iErr != nil {
			return iErr
		}
	}

	return nil
}

package helm

import (
	"github.com/pkg/errors"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (w *wrapper) createK8sResources() error {
	config, err := clientcmd.BuildConfigFromFlags("", w.kubeConfig)
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
			return errors.Wrapf(err, "failed to create serviceaccount %v", sa)
		}
	}

	crbAPI := clientset.RbacV1().ClusterRoleBindings()
	if crb, crbErr := crbAPI.Get(TillerServiceAccount, getOpts); crbErr != nil && apierrors.IsNotFound(crbErr) {
		crAPI := clientset.RbacV1().ClusterRoles()
		cr, err := crAPI.Get("cluster-admin", getOpts)
		if err != nil && apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "expected clusterrole \"cluster-admin\" to be present")
		}

		crb.Name = sa.Name
		crb.RoleRef = rbacv1.RoleRef{
			Kind: "ClusterRole",
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
			return errors.Wrapf(err, "failed to create clusterrolebinding %v", crb)
		}
	}

	return nil
}

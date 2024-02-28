package server

import (
	"context"
	"testing"

	"github.com/argoproj-labs/argocd-operator/tests/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"

	rbacv1 "k8s.io/api/rbac/v1"
	cntrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestServerReconciler_createAndDeleteClusterRoleBinding(t *testing.T) {
	sr := makeTestServerReconciler(
		test.MakeTestArgoCD(nil),
	)
	sr.varSetter()

	err := sr.reconcileClusterRoleBinding()
	assert.NoError(t, err)

	// cluster rolebinding should not be created as ArgoCD in not cluster scoped
	cr := &rbacv1.ClusterRoleBinding{}
	err = sr.Client.Get(context.TODO(), cntrlClient.ObjectKey{Name: "test-argocd-test-ns-server"}, cr)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))

	// make ArgoCD cluster scope
	sr.ClusterScoped = true
	err = sr.reconcileClusterRoleBinding()
	assert.NoError(t, err)

	// cluster rolebinding should be created as ArgoCD is cluster scoped
	cr = &rbacv1.ClusterRoleBinding{}
	err = sr.Client.Get(context.TODO(), cntrlClient.ObjectKey{Name: "test-argocd-test-ns-server"}, cr)
	assert.NoError(t, err)

	// disable cluster ArgoCD
	sr.ClusterScoped = false
	err = sr.reconcileClusterRoleBinding()
	assert.NoError(t, err)

	// cluster rolebinding should be deleted as ArgoCD is changed to namespace scoped
	cr = &rbacv1.ClusterRoleBinding{}
	err = sr.Client.Get(context.TODO(), cntrlClient.ObjectKey{Name: "test-argocd-test-ns-server"}, cr)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}

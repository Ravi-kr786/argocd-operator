package util

import (
	"fmt"

	argoprojv1a1 "github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	"github.com/argoproj-labs/argocd-operator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchSecret will retrieve the object with the given Name using the provided client.
// The result will be returned.
func FetchSecret(client client.Client, meta metav1.ObjectMeta, name string) (*corev1.Secret, error) {
	a := &argoprojv1a1.ArgoCD{}
	a.ObjectMeta = meta
	secret := NewSecretWithName(a, name)
	return secret, FetchObject(client, meta.Namespace, name, secret)
}

// NewSecret returns a new Secret based on the given metadata.
func NewSecret(cr *argoprojv1a1.ArgoCD) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: common.DefaultLabels(cr.Name, cr.Name, ""),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

// NewTLSSecret returns a new TLS Secret based on the given metadata with the provided suffix on the Name.
func NewTLSSecret(cr *argoprojv1a1.ArgoCD, suffix string) *corev1.Secret {
	secret := NewSecretWithSuffix(cr, suffix)
	secret.Type = corev1.SecretTypeTLS
	return secret
}

// NewSecretWithName returns a new Secret based on the given metadata with the provided Name.
func NewSecretWithName(cr *argoprojv1a1.ArgoCD, name string) *corev1.Secret {
	secret := NewSecret(cr)

	secret.ObjectMeta.Name = name
	secret.ObjectMeta.Namespace = cr.Namespace
	secret.ObjectMeta.Labels[common.AppK8sKeyName] = name

	return secret
}

// NewSecretWithSuffix returns a new Secret based on the given metadata with the provided suffix on the Name.
func NewSecretWithSuffix(cr *argoprojv1a1.ArgoCD, suffix string) *corev1.Secret {
	return NewSecretWithName(cr, fmt.Sprintf("%s-%s", cr.Name, suffix))
}

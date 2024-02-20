package argocd

import (
	"testing"

	argoproj "github.com/argoproj-labs/argocd-operator/api/v1beta1"
	"github.com/argoproj-labs/argocd-operator/controllers/argocd/argocdcommon"
	"github.com/argoproj-labs/argocd-operator/pkg/resource"
	"github.com/argoproj-labs/argocd-operator/pkg/util"
	"github.com/argoproj-labs/argocd-operator/pkg/workloads"
	"github.com/argoproj-labs/argocd-operator/tests/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_reconcileConfigMaps(t *testing.T) {
	testArgoCD := test.MakeTestArgoCD(nil)
	reconciler := makeTestArgoCDReconciler(
		testArgoCD,
		test.MakeTestSecret(
			nil,
			func(s *corev1.Secret) {
				s.Name = "test-argocd-ca"
				s.Data = map[string][]byte{
					"tls.crt": []byte(test.TestVal),
				}
			},
		),
	)

	expectedResources := []client.Object{
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "test-argocd-ca"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-gpg-keys-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-tls-certs-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-ssh-known-hosts-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-rbac-cm"
			},
		),
	}

	err := reconciler.reconcileConfigMaps()
	assert.NoError(t, err)

	for _, obj := range expectedResources {
		_, err := resource.GetObject(obj.GetName(), test.TestNamespace, obj, reconciler.Client)
		assert.NoError(t, err)
	}
}

func Test_deleteConfigMaps(t *testing.T) {
	testArgoCD := test.MakeTestArgoCD(nil)

	resources := []client.Object{
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "test-argocd-ca"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-gpg-keys-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-tls-certs-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-ssh-known-hosts-cm"
			},
		),
		test.MakeTestConfigMap(
			nil,
			func(cm *corev1.ConfigMap) {
				cm.Name = "argocd-rbac-cm"
			},
		),
	}

	reconciler := makeTestArgoCDReconciler(
		testArgoCD,
		resources...,
	)

	reconciler.cmVarSetter()

	err := reconciler.deleteConfigMaps()
	assert.NoError(t, err)

	for _, obj := range resources {
		_, err := resource.GetObject(obj.GetName(), test.TestNamespace, obj, reconciler.Client)
		assert.True(t, apierrors.IsNotFound(err))
	}

}

func Test_reconcileArgoCDCm(t *testing.T) {
	tests := []struct {
		name       string
		reconciler *ArgoCDReconciler
		expectedCm *corev1.ConfigMap
	}{
		{
			name: "default argocd-cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedCm: getTestArgoCDCm(),
		},
		{
			name: "modified argocd-cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.ResourceInclusions = "test-resource-inclusions"
						cr.Spec.ResourceExclusions = "test-resource-exclusions"
						cr.Spec.ResourceTrackingMethod = "annotation"
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(
				getTestArgoCDCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data["resource.inclusions"] = "test-resource-inclusions"
					cm.Data["resource.exclusions"] = "test-resource-exclusions"
					cm.Data["application.resourceTrackingMethod"] = "annotation"
				},
			),
		},
		{
			name: "drifted argocd-cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.ResourceInclusions = "test-resource-inclusions"
						cr.Spec.ResourceExclusions = "test-resource-exclusions"
						cr.Spec.ResourceTrackingMethod = "annotation"
					},
				),
				test.MakeTestConfigMap(
					getTestArgoCDCm(),
					func(cm *corev1.ConfigMap) {
						cm.Data["resource.inclusions"] = "random-info"
						cm.Data["resource.exclusions"] = "random-info"
						cm.Data["application.resourceTrackingMethod"] = "random"
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(
				getTestArgoCDCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data["resource.inclusions"] = "test-resource-inclusions"
					cm.Data["resource.exclusions"] = "test-resource-exclusions"
					cm.Data["application.resourceTrackingMethod"] = "annotation"
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.reconcileArgoCDCm()
			assert.NoError(t, err)

			existing, err := workloads.GetConfigMap("argocd-cm", test.TestNamespace, tt.reconciler.Client)
			assert.NoError(t, err)

			if tt.expectedCm != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedCm.Labels,
					},
					{
						Existing: existing.Annotations,
						Desired:  tt.expectedCm.Annotations,
					},
					{
						Existing: existing.Data,
						Desired:  tt.expectedCm.Data,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}

		})
	}
}

func Test_reconcileCaCm(t *testing.T) {
	tests := []struct {
		name          string
		reconciler    *ArgoCDReconciler
		expectedCm    *corev1.ConfigMap
		expectedError bool
	}{
		{
			name: "no ca secret found",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedCm:    nil,
			expectedError: true,
		},
		{
			name: "ca secret found",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
				test.MakeTestSecret(
					nil,
					func(s *corev1.Secret) {
						s.Name = "test-argocd-ca"
						s.Data = map[string][]byte{
							"tls.crt": []byte(test.TestVal),
						}
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(
				getTestCaCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data["tls.crt"] = "test-val"
				},
			),
			expectedError: false,
		},
		{
			name: "ca config map drift",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
				test.MakeTestSecret(
					nil,
					func(s *corev1.Secret) {
						s.Name = "test-argocd-ca"
						s.Data = map[string][]byte{
							"tls.crt": []byte(test.TestVal),
						}
					},
				),
				test.MakeTestConfigMap(
					nil,
					func(cm *corev1.ConfigMap) {
						cm.Name = "test-argocd-ca"
						cm.Data = map[string]string{
							"tls.crt": "random-val",
						}
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(
				getTestCaCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data["tls.crt"] = "test-val"
				},
			),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reconciler.cmVarSetter()

			err := tt.reconciler.reconcileCACm()
			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none.")
			} else {
				assert.NoError(t, err, "Expected no error but got one.")
			}

			existing, err := workloads.GetConfigMap("test-argocd-ca", test.TestNamespace, tt.reconciler.Client)

			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none.")
			} else {
				assert.NoError(t, err, "Expected no error but got one.")
			}

			if tt.expectedCm != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedCm.Labels,
					},
					{
						Existing: existing.Annotations,
						Desired:  tt.expectedCm.Annotations,
					},
					{
						Existing: existing.Data,
						Desired:  tt.expectedCm.Data,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}

		})
	}
}

func Test_reconcileGPGKeysCm(t *testing.T) {
	tests := []struct {
		name       string
		reconciler *ArgoCDReconciler
		expectedCm *corev1.ConfigMap
	}{
		{
			name: "default argocd-gpg-keys cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedCm: getTestGPGKeysCm(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.reconcileGPGKeysCm()
			assert.NoError(t, err)

			existing, err := workloads.GetConfigMap("argocd-gpg-keys-cm", test.TestNamespace, tt.reconciler.Client)
			assert.NoError(t, err)

			if tt.expectedCm != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedCm.Labels,
					},
					{
						Existing: existing.Annotations,
						Desired:  tt.expectedCm.Annotations,
					},
					{
						Existing: existing.Data,
						Desired:  tt.expectedCm.Data,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}

		})
	}
}

func Test_reconcileTLSCertsCm(t *testing.T) {
	tests := []struct {
		name       string
		reconciler *ArgoCDReconciler
		expectedCm *corev1.ConfigMap
	}{
		{
			name: "default argocd-tls-certs cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedCm: test.MakeTestConfigMap(
				getTestTLSCertsCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data = nil
				},
			),
		},
		{
			name: "modified argocd-tls-certs cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.TLS.InitialCerts = test.TestKVP
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(getTestTLSCertsCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data = test.TestKVP
				},
			),
		},
		{
			name: "drifted argocd-tls-certs cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.TLS.InitialCerts = test.TestKVP
					},
				),
				test.MakeTestConfigMap(getTestTLSCertsCm(),
					func(cm *corev1.ConfigMap) {
						cm.Data = map[string]string{
							"test-key": "random-info",
						}
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(getTestTLSCertsCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data = test.TestKVP
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.reconcileTLSCertsCm()
			assert.NoError(t, err)

			existing, err := workloads.GetConfigMap("argocd-tls-certs-cm", test.TestNamespace, tt.reconciler.Client)
			assert.NoError(t, err)

			if tt.expectedCm != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedCm.Labels,
					},
					{
						Existing: existing.Annotations,
						Desired:  tt.expectedCm.Annotations,
					},
					{
						Existing: existing.Data,
						Desired:  tt.expectedCm.Data,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}

		})
	}
}

func Test_reconcileSSHKnownHostsCm(t *testing.T) {
	tests := []struct {
		name       string
		reconciler *ArgoCDReconciler
		expectedCm *corev1.ConfigMap
	}{
		{
			name: "default ssh-known-hosts cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedCm: getTestSSHKnownHostsCm(),
		},
		{
			name: "modified ssh-known-hosts cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.InitialSSHKnownHosts.Keys = test.TestKey
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(getTestSSHKnownHostsCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data["ssh_known_hosts"] = `[ssh.github.com]:443 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
[ssh.github.com]:443 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
[ssh.github.com]:443 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
bitbucket.org ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPIQmuzMBuKdWeF4+a2sjSSpBK0iqitSQ+5BM9KhpexuGt20JpTVM7u5BDZngncgrqDMbWdxMWWOGtZ9UgbqgZE=
bitbucket.org ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIazEu89wgQZ4bqs3d63QSMzYVa0MuJ2e2gKTKqu+UUO
bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
ssh.dev.azure.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
vs-ssh.visualstudio.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
test-key`
				},
			),
		},
		{
			name: "drifted ssh-known-hosts cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.InitialSSHKnownHosts.Keys = test.TestKey
					},
				),
				test.MakeTestConfigMap(
					getTestSSHKnownHostsCm(),
					func(cm *corev1.ConfigMap) {
						cm.Data["ssh_known_hosts"] = `[ssh.github.com]:443 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
[ssh.github.com]:443 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
[ssh.github.com]:443 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
bitbucket.org ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPIQmuzMBuKdWeF4+a2sjSSpBK0iqitSQ+5BM9KhpexuGt20JpTVM7u5BDZngncgrqDMbWdxMWWOGtZ9UgbqgZE=
bitbucket.org ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIazEu89wgQZ4bqs3d63QSMzYVa0MuJ2e2gKTKqu+UUO
bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
ssh.dev.azure.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
vs-ssh.visualstudio.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
test-keyssss`
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(getTestSSHKnownHostsCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data["ssh_known_hosts"] = `[ssh.github.com]:443 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
[ssh.github.com]:443 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
[ssh.github.com]:443 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
bitbucket.org ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPIQmuzMBuKdWeF4+a2sjSSpBK0iqitSQ+5BM9KhpexuGt20JpTVM7u5BDZngncgrqDMbWdxMWWOGtZ9UgbqgZE=
bitbucket.org ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIazEu89wgQZ4bqs3d63QSMzYVa0MuJ2e2gKTKqu+UUO
bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
ssh.dev.azure.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
vs-ssh.visualstudio.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
test-key`
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.reconcileSSHKnownHostsCm()
			assert.NoError(t, err)

			existing, err := workloads.GetConfigMap("argocd-ssh-known-hosts-cm", test.TestNamespace, tt.reconciler.Client)
			assert.NoError(t, err)

			if tt.expectedCm != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedCm.Labels,
					},
					{
						Existing: existing.Annotations,
						Desired:  tt.expectedCm.Annotations,
					},
					{
						Existing: existing.Data,
						Desired:  tt.expectedCm.Data,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}

		})
	}
}

func Test_reconcileRBACCm(t *testing.T) {
	tests := []struct {
		name       string
		reconciler *ArgoCDReconciler
		expectedCm *corev1.ConfigMap
	}{
		{
			name: "default rbac cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedCm: getTestRbacCm(),
		},
		{
			name: "modified rbac cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.RBAC.Policy = util.StringPtr("p, subj, resource, action")
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(getTestRbacCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data = map[string]string{
						"policy.csv":       "p, subj, resource, action",
						"policy.default":   "",
						"scopes":           "[groups]",
						"policy.matchMode": "",
					}
				},
			),
		},
		{
			name: "drifted rbac cm",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil,
					func(cr *argoproj.ArgoCD) {
						cr.Spec.RBAC.Policy = util.StringPtr("p, subj, resource, action")
					},
				),
				test.MakeTestConfigMap(
					getTestRbacCm(),
					func(cm *corev1.ConfigMap) {
						cm.Data = map[string]string{
							"policy.csv":       "p, subj, resource 1, resource 2, action",
							"policy.default":   "",
							"scopes":           "[groups]",
							"policy.matchMode": "",
						}
					},
				),
			),
			expectedCm: test.MakeTestConfigMap(getTestRbacCm(),
				func(cm *corev1.ConfigMap) {
					cm.Data = map[string]string{
						"policy.csv":       "p, subj, resource, action",
						"policy.default":   "",
						"scopes":           "[groups]",
						"policy.matchMode": "",
					}
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.reconcileRBACCm()
			assert.NoError(t, err)

			existing, err := workloads.GetConfigMap("argocd-rbac-cm", test.TestNamespace, tt.reconciler.Client)
			assert.NoError(t, err)

			if tt.expectedCm != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedCm.Labels,
					},
					{
						Existing: existing.Annotations,
						Desired:  tt.expectedCm.Annotations,
					},
					{
						Existing: existing.Data,
						Desired:  tt.expectedCm.Data,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}

		})
	}
}

func TestDeleteConfigMap(t *testing.T) {
	tests := []struct {
		name           string
		reconciler     *ArgoCDReconciler
		configMapExist bool
		expectedError  bool
	}{
		{
			name: "ConfigMap exists",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
				test.MakeTestConfigMap(nil),
			),
			configMapExist: true,
			expectedError:  false,
		},
		{
			name: "ConfigMap does not exist",
			reconciler: makeTestArgoCDReconciler(
				test.MakeTestArgoCD(nil),
			),
			configMapExist: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.deleteConfigMap(test.TestName, test.TestNamespace)

			if tt.configMapExist {
				_, err := workloads.GetConfigMap(test.TestName, test.TestNamespace, tt.reconciler.Client)
				assert.True(t, apierrors.IsNotFound(err))
			}

			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none.")
			} else {
				assert.NoError(t, err, "Expected no error but got one.")
			}
		})
	}
}

func getTestArgoCDCm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-cm",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "argocd-cm",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
		Data: map[string]string{
			"application.instanceLabelKey":       "app.kubernetes.io/instance",
			"admin.enabled":                      "true",
			"ga.trackingid":                      "",
			"ga.anonymizeusers":                  "false",
			"configManagementPlugins":            "",
			"help.chatUrl":                       "",
			"help.chatText":                      "",
			"kustomize.buildOptions":             "",
			"oidc.config":                        "",
			"resource.exclusions":                "",
			"resource.inclusions":                "",
			"application.resourceTrackingMethod": "label",
			"repositories":                       "",
			"repository.credentials":             "",
			"statusbadge.enabled":                "false",
			"users.anonymous.enabled":            "false",
		},
	}
}

func getTestCaCm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd-ca",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "test-argocd-ca",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
		Data: make(map[string]string),
	}
}

func getTestGPGKeysCm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-gpg-keys-cm",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "argocd-gpg-keys-cm",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
	}
}

func getTestTLSCertsCm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-tls-certs-cm",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "argocd-tls-certs-cm",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
		Data: make(map[string]string),
	}
}

func getTestSSHKnownHostsCm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-ssh-known-hosts-cm",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "argocd-ssh-known-hosts-cm",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
		Data: map[string]string{
			"ssh_known_hosts": `[ssh.github.com]:443 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
[ssh.github.com]:443 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
[ssh.github.com]:443 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
bitbucket.org ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPIQmuzMBuKdWeF4+a2sjSSpBK0iqitSQ+5BM9KhpexuGt20JpTVM7u5BDZngncgrqDMbWdxMWWOGtZ9UgbqgZE=
bitbucket.org ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIazEu89wgQZ4bqs3d63QSMzYVa0MuJ2e2gKTKqu+UUO
bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
ssh.dev.azure.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
vs-ssh.visualstudio.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H
`,
		},
	}
}

func getTestRbacCm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-rbac-cm",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "argocd-rbac-cm",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
		Data: map[string]string{
			"policy.csv":       "",
			"policy.default":   "",
			"scopes":           "[groups]",
			"policy.matchMode": "",
		},
	}
}

// Copyright 2020 ArgoCD Operator Developers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package argocd

import (
	"testing"

	"github.com/go-logr/logr"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/argoproj-labs/argocd-operator/pkg/argoutil"

	argoproj "github.com/argoproj-labs/argocd-operator/api/v1beta1"
)

const (
	testNamespace             = "argocd"
	testArgoCDName            = "argocd"
	testApplicationController = "argocd-application-controller"
)

func ZapLogger(development bool) logr.Logger {
	return zap.New(zap.UseDevMode(development))
}

type namespaceOpt func(*corev1.Namespace)

func makeTestNs(opts ...namespaceOpt) *corev1.Namespace {
	a := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-ns",
			Labels: make(map[string]string),
		},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

type SchemeOpt func(*runtime.Scheme) error

func makeNewTestReconciler(client client.Client, sch *runtime.Scheme) *ArgoCDReconciler {
	return &ArgoCDReconciler{
		Client: client,
		Scheme: sch,
	}
}

func makeTestReconcilerClient(sch *runtime.Scheme, resObjs, subresObjs []client.Object, runtimeObj []runtime.Object) client.Client {
	client := fake.NewClientBuilder().WithScheme(sch)
	if len(resObjs) > 0 {
		client = client.WithObjects(resObjs...)
	}
	if len(subresObjs) > 0 {
		client = client.WithStatusSubresource(subresObjs...)
	}
	if len(runtimeObj) > 0 {
		client = client.WithRuntimeObjects(runtimeObj...)
	}
	return client.Build()
}

func makeTestReconcilerScheme(sOpts ...SchemeOpt) *runtime.Scheme {
	s := scheme.Scheme
	for _, opt := range sOpts {
		_ = opt(s)
	}

	return s
}

type argoCDOpt func(*argoproj.ArgoCD)

func makeTestArgoCD(opts ...argoCDOpt) *argoproj.ArgoCD {
	a := &argoproj.ArgoCD{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testArgoCDName,
			Namespace: testNamespace,
		},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func makeTestArgoCDForKeycloak() *argoproj.ArgoCD {
	a := &argoproj.ArgoCD{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testArgoCDName,
			Namespace: testNamespace,
		},
		Spec: argoproj.ArgoCDSpec{
			SSO: &argoproj.ArgoCDSSOSpec{
				Provider: "keycloak",
			},
			Server: argoproj.ArgoCDServerSpec{
				Route: argoproj.ArgoCDRouteSpec{
					Enabled: true,
				},
			},
		},
	}
	return a
}

func makeTestArgoCDWithResources(opts ...argoCDOpt) *argoproj.ArgoCD {
	a := &argoproj.ArgoCD{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testArgoCDName,
			Namespace: testNamespace,
		},
		Spec: argoproj.ArgoCDSpec{
			ApplicationSet: &argoproj.ArgoCDApplicationSet{
				Resources: makeTestApplicationSetResources(),
			},
			HA: argoproj.ArgoCDHASpec{
				Resources: makeTestHAResources(),
			},
			SSO: &argoproj.ArgoCDSSOSpec{
				Provider: "dex",
				Dex: &argoproj.ArgoCDDexSpec{
					Resources: makeTestDexResources(),
				},
			},
			Controller: argoproj.ArgoCDApplicationControllerSpec{
				Resources: makeTestControllerResources(),
			},
		},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func makeTestClusterRole() *v1.ClusterRole {
	return &v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testApplicationController,
			Namespace: testNamespace,
		},
		Rules: makeTestPolicyRules(),
	}
}

func makeTestDeployment() *appsv1.Deployment {
	var replicas int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testApplicationController,
			Namespace: testNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "name",
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Command: []string{"testing"},
							Image:   "test-image",
						},
					},
				},
			},
		},
	}
}

func makeTestPolicyRules() []v1.PolicyRule {
	return []v1.PolicyRule{
		{
			APIGroups: []string{
				"foo.example.com",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
	}
}

func initialCerts(t *testing.T, host string) argoCDOpt {
	t.Helper()
	return func(a *argoproj.ArgoCD) {
		key, err := argoutil.NewPrivateKey()
		assert.NoError(t, err)
		cert, err := argoutil.NewSelfSignedCACertificate(a.Name, key)
		assert.NoError(t, err)
		encoded := argoutil.EncodeCertificatePEM(cert)

		a.Spec.TLS.InitialCerts = map[string]string{
			host: string(encoded),
		}
	}
}

func makeTestControllerResources() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("1024Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("1000m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("2048Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("2000m"),
		},
	}
}

func makeTestApplicationSetResources() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("1024Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("1"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("2048Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("2"),
		},
	}
}

func makeTestHAResources() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("128Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("250m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("256Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("500m"),
		},
	}
}

func makeTestDexResources() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("128Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("250m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resourcev1.MustParse("256Mi"),
			corev1.ResourceCPU:    resourcev1.MustParse("500m"),
		},
	}
}

// func createNs(r *ArgoCDReconciler, n string, managedBy string) error {
// 	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: n}}
// 	if managedBy != "" {
// 		ns.Labels = map[string]string{common.ArgoCDArgoprojKeyManagedBy: managedBy}
// 	}

// 	if r.ResourceManagedNamespaces == nil {
// 		r.ResourceManagedNamespaces = make(map[string]string)
// 	}
// 	r.ResourceManagedNamespaces[ns.Name] = ""

// 	return r.Client.Create(context.TODO(), ns)
// }

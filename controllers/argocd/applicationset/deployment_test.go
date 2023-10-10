package applicationset

import (
	"context"
	"testing"

	"github.com/argoproj-labs/argocd-operator/controllers/argocd/argocdcommon"
	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestApplicationSetReconciler_reconcileDeployment(t *testing.T) {
	resourceName = argocdcommon.TestArgoCDName
	resourceLabels = testExpectedLabels
	ns := argocdcommon.MakeTestNamespace()
	asr := makeTestApplicationSetReconciler(t, false, ns)

	existingDeployment := asr.getDesiredDeployment()

	tests := []struct {
		name        string
		setupClient func() *ApplicationSetReconciler
		wantErr     bool
	}{
		{
			name: "create a deployment",
			setupClient: func() *ApplicationSetReconciler {
				return makeTestApplicationSetReconciler(t, false, ns)
			},
			wantErr: false,
		},
		{
			name: "update a deployment",
			setupClient: func() *ApplicationSetReconciler {
				outdatedDeployment := existingDeployment
				outdatedDeployment.ObjectMeta.Labels = argocdcommon.TestKVP
				return makeTestApplicationSetReconciler(t, false, outdatedDeployment, ns)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asr := tt.setupClient()
			err := asr.reconcileDeployment()
			if (err != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Expected error but did not get one")
				} else {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			updatedDeployment := &appsv1.Deployment{}
			err = asr.Client.Get(context.TODO(), types.NamespacedName{Name: resourceName, Namespace: argocdcommon.TestNamespace}, updatedDeployment)
			if err != nil {
				t.Fatalf("Could not get updated Deployment: %v", err)
			}
			assert.Equal(t, testExpectedLabels, updatedDeployment.ObjectMeta.Labels)
		})
	}
}

func TestApplicationSetReconciler_DeleteDeployment(t *testing.T) {
	ns := argocdcommon.MakeTestNamespace()
	resourceName = argocdcommon.TestArgoCDName
	tests := []struct {
		name        string
		setupClient func() *ApplicationSetReconciler
		wantErr     bool
	}{
		{
			name: "successful delete",
			setupClient: func() *ApplicationSetReconciler {
				return makeTestApplicationSetReconciler(t, false, ns)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asr := tt.setupClient()
			if err := asr.deleteDeployment(resourceName, ns.Name); (err != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Expected error but did not get one")
				} else {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

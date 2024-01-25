package reposerver

// func TestRepoServerReconciler_reconcileTLSService(t *testing.T) {
// 	ns := argocdcommon.MakeTestNamespace()
// 	sa := argocdcommon.MakeTestServiceAccount()
// 	resourceName = argocdcommon.TestArgoCDName

// 	tests := []struct {
// 		name        string
// 		setupClient func() *RepoServerReconciler
// 		wantErr     bool
// 	}{
// 		{
// 			name: "create a Service",
// 			setupClient: func() *RepoServerReconciler {
// 				return makeTestRepoServerReconciler(t, ns, sa)
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			rsr := tt.setupClient()
// 			err := rsr.reconcileService()
// 			if (err != nil) != tt.wantErr {
// 				if tt.wantErr {
// 					t.Errorf("Expected error but did not get one")
// 				} else {
// 					t.Errorf("Unexpected error: %v", err)
// 				}
// 			}
// 			currentService := &corev1.Service{}
// 			err = rsr.Client.Get(context.TODO(), types.NamespacedName{Name: argocdcommon.TestArgoCDName, Namespace: argocdcommon.TestNamespace}, currentService)
// 			if err != nil {
// 				t.Fatalf("Could not get current Service: %v", err)
// 			}
// 			assert.Equal(t, GetServiceSpec().Ports, currentService.Spec.Ports)
// 		})
// 	}
// }

// func TestRepoServerReconciler_DeleteService(t *testing.T) {
// 	ns := argocdcommon.MakeTestNamespace()
// 	sa := argocdcommon.MakeTestServiceAccount()
// 	resourceName = argocdcommon.TestArgoCDName
// 	tests := []struct {
// 		name        string
// 		setupClient func() *RepoServerReconciler
// 		wantErr     bool
// 	}{
// 		{
// 			name: "successful delete",
// 			setupClient: func() *RepoServerReconciler {
// 				return makeTestRepoServerReconciler(t, sa, ns)
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			rsr := tt.setupClient()
// 			if err := rsr.deleteService(resourceName, ns.Name); (err != nil) != tt.wantErr {
// 				if tt.wantErr {
// 					t.Errorf("Expected error but did not get one")
// 				} else {
// 					t.Errorf("Unexpected error: %v", err)
// 				}
// 			}
// 		})
// 	}
// }

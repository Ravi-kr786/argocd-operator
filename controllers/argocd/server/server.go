package server

import (
	"github.com/argoproj-labs/argocd-operator/pkg/cluster"
	"github.com/argoproj-labs/argocd-operator/pkg/networking"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	argoproj "github.com/argoproj-labs/argocd-operator/api/v1beta1"
)

type ServerReconciler struct {
	Client            client.Client
	Scheme            *runtime.Scheme
	Instance          *argoproj.ArgoCD
	ClusterScoped     bool
	Logger            logr.Logger
	ManagedNamespaces map[string]string
	SourceNamespaces  map[string]string
}

func (sr *ServerReconciler) Reconcile() error {

	sr.Logger = ctrl.Log.WithName(ServerControllerComponent).WithValues("instance", sr.Instance.Name, "instance-namespace", sr.Instance.Namespace)

	// namespace doesn't exist, break reconciliation
	ns, err := cluster.GetNamespace(sr.Instance.Namespace, sr.Client)
	if err != nil {
		sr.Logger.Error(err, "Reconcile: failed to retrieve namespace", "name", sr.Instance.Namespace)
		return err
	}

	// namespace is deleting, delete the resources
	if ns.DeletionTimestamp != nil {
		return sr.DeleteResources()
	}

	// perform resource reconciliation
	if err := sr.reconcileServiceAccount(); err != nil {
		return err
	}

	if err := sr.reconcileClusterRole(); err != nil {
		return err
	}

	if err := sr.reconcileClusterRoleBinding(); err != nil {
		return err
	}

	if err := sr.reconcileRoles(); err != nil {
		return err
	}

	if err := sr.reconcileRoleBindings(); err != nil {
		return err
	}

	if err := sr.reconcileDeployment(); err != nil {
		return err
	}

	if err := sr.reconcileService(); err != nil {
		return err
	}

	if err := sr.reconcileHorizontalPodAutoscaler(); err != nil {
		return err
	}

	if err := sr.reconcileIngresses(); err != nil {
		return err
	}

	if networking.IsRouteAPIAvailable() {
		if err := sr.reconcileRoute(); err != nil {
			return err
		}
	}

	return nil
}

// TO DO: fix this
func (acr *ServerReconciler) TriggerRollout(key string) error {
	return acr.TriggerDeploymentRollout("", "", key)
}

func (sr *ServerReconciler) DeleteResources() error {

	name := sr.Instance.Name
	ns := sr.Instance.Namespace

	if networking.IsRouteAPIAvailable() {
		if err := sr.deleteRoute(getRouteName(name), ns); err != nil {
			return err
		}
	}

	if err := sr.deleteIngresses(name, ns); err != nil {
		return err
	}

	if err := sr.deleteHorizontalPodAutoscaler(getHPAName(name), ns); err != nil {
		return err
	}

	if err := sr.deleteService(getServiceName(name), ns); err != nil {
		return err
	}

	if err := sr.deleteDeployment(getDeploymentName(name), ns); err != nil {
		return err
	}

	if err := sr.deleteRoleBindings(name, ns); err != nil {
		return err
	}

	if err := sr.deleteRoles(name, ns); err != nil {
		return err
	}

	if err := sr.deleteClusterRoleBinding(getClusterRoleBindingName(name, ns)); err != nil {
		return err
	}

	if err := sr.deleteClusterRole(getClusterRoleName(name, ns)); err != nil {
		return err
	}

	if err := sr.deleteServiceAccount(getServiceAccountName(name), ns); err != nil {
		return err
	}

	return nil
}

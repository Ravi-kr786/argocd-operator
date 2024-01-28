package reposerver

import (
	"context"

	"github.com/argoproj-labs/argocd-operator/common"
	"github.com/argoproj-labs/argocd-operator/pkg/workloads"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/retry"
)

// reconcileStatus will ensure that the Repo-server status is updated for the given ArgoCD instance
func (rsr *RepoServerReconciler) reconcileStatus() error {
	status := common.ArgoCDStatusUnknown

	deploy, err := workloads.GetDeployment(resourceName, rsr.Instance.Namespace, rsr.Client)
	if err != nil {
		return errors.Wrapf(err, "reconcileStatus: failed to retrieve deployment %s", resourceName)
	}

	status = common.ArgoCDStatusPending

	if deploy.Spec.Replicas != nil {
		if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
			status = common.ArgoCDStatusRunning
		}
	}

	if rsr.Instance.Status.Repo != status {
		rsr.Instance.Status.Repo = status
	}

	return rsr.UpdateInstanceStatus()
}

func (rsr *RepoServerReconciler) UpdateInstanceStatus() error {

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := rsr.Client.Status().Update(context.TODO(), rsr.Instance); err != nil {
			return errors.Wrap(err, "UpdateInstanceStatus: failed to update instance status")
		}
		return nil
	})

	if err != nil {
		// May be conflict if max retries were hit, or may be something unrelated
		// like permissions or a network error
		return err
	}
	return nil
}

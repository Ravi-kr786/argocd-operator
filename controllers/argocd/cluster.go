package argocd

import (
	"os"

	"github.com/argoproj-labs/argocd-operator/common"
	"github.com/argoproj-labs/argocd-operator/pkg/cluster"
	"github.com/argoproj-labs/argocd-operator/pkg/monitoring"
	"github.com/argoproj-labs/argocd-operator/pkg/networking"
	"github.com/argoproj-labs/argocd-operator/pkg/util"
	"github.com/argoproj-labs/argocd-operator/pkg/workloads"
)

// InspectCluster will verify the availability of extra features on the cluster, such as Prometheus and OpenShift Routes.
func InspectCluster() error {
	var inspectError error

	if err := monitoring.VerifyPrometheusAPI(); err != nil {
		inspectError = err
	}

	if err := networking.VerifyRouteAPI(); err != nil {
		inspectError = err
	}

	if err := workloads.VerifyTemplateAPI(); err != nil {
		inspectError = err
	}

	if err := cluster.VerifyVersionAPI(); err != nil {
		inspectError = err
	}

	return inspectError
}

func GetClusterConfigNamespaces() string {
	return os.Getenv(common.ArgoCDClusterConfigNamespacesEnvVar)
}

func IsClusterConfigNs(current string) bool {
	clusterConfigNamespaces := util.SplitList(GetClusterConfigNamespaces())
	if len(clusterConfigNamespaces) > 0 {
		if clusterConfigNamespaces[0] == "*" {
			return true
		}

		for _, n := range clusterConfigNamespaces {
			if n == current {
				return true
			}
		}
	}
	return false
}

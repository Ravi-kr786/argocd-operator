package common

// names
const (
	// ArgoCDServerComponent is the name of the Dex server control plane component
	ArgoCDServerComponent = "argocd-server"

	// ArgoCDServerTLSSecretName is the name of the TLS secret for the argocd-server
	ArgoCDServerTLSSecretName = "argocd-server-tls"
)

// defaults
const (

	// ArgoCDDefaultServerOperationProcessors is the number of ArgoCD Server Operation Processors to use when not specified.
	ArgoCDDefaultServerOperationProcessors = int32(10)

	// ArgoCDDefaultServerStatusProcessors is the number of ArgoCD Server Status Processors to use when not specified.
	ArgoCDDefaultServerStatusProcessors = int32(20)

	// ArgoCDDefaultServerResourceLimitCPU is the default CPU limit when not specified for the Argo CD server contianer.
	ArgoCDDefaultServerResourceLimitCPU = "1000m"

	// ArgoCDDefaultServerResourceLimitMemory is the default memory limit when not specified for the Argo CD server contianer.
	ArgoCDDefaultServerResourceLimitMemory = "128Mi"

	// ArgoCDDefaultServerResourceRequestCPU is the default CPU requested when not specified for the Argo CD server contianer.
	ArgoCDDefaultServerResourceRequestCPU = "250m"

	// ArgoCDDefaultServerResourceRequestMemory is the default memory requested when not specified for the Argo CD server contianer.
	ArgoCDDefaultServerResourceRequestMemory = "64Mi"

	// ArgoCDDefaultServerSessionKeyLength is the length of the generated default server signature key.
	ArgoCDDefaultServerSessionKeyLength = 20

	// ArgoCDDefaultServerSessionKeyNumDigits is the number of digits to use for the generated default server signature key.
	ArgoCDDefaultServerSessionKeyNumDigits = 5

	// ArgoCDDefaultServerSessionKeyNumSymbols is the number of symbols to use for the generated default server signature key.
	ArgoCDDefaultServerSessionKeyNumSymbols = 0
)

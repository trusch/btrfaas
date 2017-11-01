package deployment

import (
	"context"
	"time"
)

// Platform is the interface for all deloyable platforms
// think about:
// * (local) swarm
// * k8s
type Platform interface {
	// PrepareEnvironment prepares an environment to start deploying services
	// This should contain all one time setup like creating namespaces/networks etc.
	PrepareEnvironment(ctx context.Context, options *PrepareEnvironmentOptions) error

	// DeployService deploys a service in an environment
	DeployService(ctx context.Context, options *DeployServiceOptions) error

	// UndeployService unddeploys a service from an environment
	UndeployService(ctx context.Context, options *UndeployServiceOptions) error

	// ListServices returns a list of all deployed services
	ListServices(ctx context.Context, options *ListServicesOptions) ([]*ServiceInfo, error)

	// DeploySecret deploys a secret in an environment
	DeploySecret(ctx context.Context, options *DeploySecretOptions) error

	// UndeploySecret unddeploys a secret from an environment
	UndeploySecret(ctx context.Context, options *UndeploySecretOptions) error
}

// PrepareEnvironmentOptions contains the options for the PrepareEnvironment call
type PrepareEnvironmentOptions struct {
	EnvironmentID string
}

// DeployServiceOptions contains the options for the DeployService call
type DeployServiceOptions struct {
	EnvironmentID string
	ServiceID     string
	Image         string
	Cmd           []string
	Env           map[string]string
	Secrets       []string
}

// UndeployServiceOptions contains the options for the UndeployService call
type UndeployServiceOptions struct {
	EnvironmentID string
	ServiceID     string
}

// ListServicesOptions contains the options for the ListServices call
type ListServicesOptions struct {
	EnvironmentID string
}

// ServiceInfo contains infos about a running service
type ServiceInfo struct {
	DeployServiceOptions
	CreatedAt time.Time
	ExitedAt  time.Time
}

// DeploySecretOptions contains the options for the DeploySecret call
type DeploySecretOptions struct {
	EnvironmentID string
	SecretID      string
	Value         string
}

// UndeploySecretOptions contains the options for the UndeploySecret call
type UndeploySecretOptions struct {
	EnvironmentID string
	SecretID      string
}

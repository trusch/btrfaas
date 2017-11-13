package deployment

import (
	"context"
	"time"
)

// Platform is the interface for all deloyable platforms
// think about:
// * swarm
// * k8s
type Platform interface {
	// PrepareEnvironment prepares an environment to start deploying services
	// This should contain all one time setup like creating namespaces/networks etc.
	PrepareEnvironment(ctx context.Context, options *PrepareEnvironmentOptions) error

	// TeardownEnvironment cleans the environment completely
	TeardownEnvironment(ctx context.Context, options *TeardownEnvironmentOptions) error

	// DeployService deploys a service in an environment
	DeployService(ctx context.Context, options *DeployServiceOptions) error

	// UndeployService unddeploys a service from an environment
	UndeployService(ctx context.Context, options *UndeployServiceOptions) error

	// ListServices returns a list of all deployed services
	ListServices(ctx context.Context, options *ListServicesOptions) ([]*ServiceInfo, error)

	// ScaleService scales the service
	ScaleService(ctx context.Context, options *ScaleServiceOptions) error

	// DeploySecret deploys a secret in an environment
	DeploySecret(ctx context.Context, options *DeploySecretOptions) error

	// UndeploySecret unddeploys a secret from an environment
	UndeploySecret(ctx context.Context, options *UndeploySecretOptions) error

	// ListSecrets returns a list of all deployed secrets
	ListSecrets(ctx context.Context, options *ListSecretsOptions) ([]*SecretInfo, error)
}

// PrepareEnvironmentOptions contains the options for the PrepareEnvironment call
type PrepareEnvironmentOptions struct {
	ID string
}

// TeardownEnvironmentOptions contains the options for the TeardownEnvironment call
type TeardownEnvironmentOptions struct {
	ID string
}

// DeployServiceOptions contains the options for the DeployService call
type DeployServiceOptions struct {
	EnvironmentID string
	ID            string
	Image         string
	Labels        LabelSet
	Cmd           []string
	Ports         map[uint16]uint16 // Host Port -> Container Port
	Env           LabelSet          // Environment variables: key -> val mapping
	Secrets       LabelSet          // Secrets: secret-id -> target-path mapping
	Volumes       []*VolumeConfig
}

// VolumeConfig specifies a volume
type VolumeConfig struct {
	Type   string
	Source string
	Target string
}

// UndeployServiceOptions contains the options for the UndeployService call
type UndeployServiceOptions struct {
	EnvironmentID string
	ID            string
}

// ListServicesOptions contains the options for the ListServices call
type ListServicesOptions struct {
	EnvironmentID string
	Labels        LabelSet
}

// ServiceInfo contains infos about a running service
type ServiceInfo struct {
	ID        string
	Image     string
	Labels    LabelSet
	Cmd       []string
	Env       LabelSet // Environment variables: key -> val mapping
	Secrets   LabelSet // Secrets: secret-id -> target-path mapping
	CreatedAt time.Time
	Endpoint  string
	Scale     uint64
}

// DeploySecretOptions contains the options for the DeploySecret call
type DeploySecretOptions struct {
	EnvironmentID string
	ID            string
	Labels        LabelSet
	Value         string
}

// UndeploySecretOptions contains the options for the UndeploySecret call
type UndeploySecretOptions struct {
	EnvironmentID string
	ID            string
}

// ListSecretsOptions contains the options for the ListSecrets call
type ListSecretsOptions struct {
	EnvironmentID string
	Labels        LabelSet
}

// SecretInfo is the (inner) response type for ListSecrets calls
type SecretInfo struct {
	ID     string
	Labels LabelSet
}

// LabelSet is a set of key-value pairs (string-string)
type LabelSet map[string]string

// ScaleServiceOptions are the options for the ScaleService call
type ScaleServiceOptions struct {
	ServiceID string
	Scale     uint64
}

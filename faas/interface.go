package faas

import (
	"context"
	"io"

	"github.com/trusch/btrfaas/deployment"
)

// FaaS is the interface for a function-as-a-service platform
type FaaS interface {

	// Init initializes the FaaS
	Init(ctx context.Context, options *InitOptions) error

	// Invoke performs a function call agains the faas
	Invoke(ctx context.Context, options *InvokeOptions) error

	// Teardown cleans the FaaS completely
	Teardown(ctx context.Context, options *TeardownOptions) error

	// DeployFunction deploys a service in an environment
	DeployFunction(ctx context.Context, options *DeployFunctionOptions) error

	// UndeployFunction unddeploys a service from an environment
	UndeployFunction(ctx context.Context, options *UndeployFunctionOptions) error

	// ListFunctions returns a list of all deployed services
	ListFunctions(ctx context.Context, options *ListFunctionsOptions) ([]*FunctionInfo, error)

	// ScaleFunction scales the service
	ScaleFunction(ctx context.Context, options *ScaleFunctionOptions) error

	// DeploySecret deploys a secret in an environment
	DeploySecret(ctx context.Context, options *DeploySecretOptions) error

	// UndeploySecret unddeploys a secret from an environment
	UndeploySecret(ctx context.Context, options *UndeploySecretOptions) error

	// ListSecrets returns a list of all deployed secrets
	ListSecrets(ctx context.Context, options *ListSecretsOptions) ([]*SecretInfo, error)
}

// InitOptions contain the options for the init call
type InitOptions struct {
	deployment.PrepareEnvironmentOptions `yaml:",inline"`
}

// TeardownOptions contain the options for the teardown call
type TeardownOptions struct {
	deployment.TeardownEnvironmentOptions `yaml:",inline"`
}

// DeployFunctionOptions contains the options for the DeployFunction call
type DeployFunctionOptions struct {
	deployment.DeployServiceOptions `yaml:",inline"`
}

// UndeployFunctionOptions contains the options for the UndeployFunction call
type UndeployFunctionOptions struct {
	deployment.UndeployServiceOptions `yaml:",inline"`
}

// ListFunctionsOptions contains the options for the ListFunctions call
type ListFunctionsOptions struct {
	deployment.ListServicesOptions `yaml:",inline"`
}

// FunctionInfo contains infos about a running function service
type FunctionInfo struct {
	deployment.ServiceInfo `yaml:",inline"`
}

// DeploySecretOptions contains the options for the DeploySecret call
type DeploySecretOptions struct {
	deployment.DeploySecretOptions `yaml:",inline"`
}

// UndeploySecretOptions contains the options for the UndeploySecret call
type UndeploySecretOptions struct {
	deployment.UndeploySecretOptions `yaml:",inline"`
}

// ListSecretsOptions contains the options for the ListSecrets call
type ListSecretsOptions struct {
	deployment.ListSecretsOptions `yaml:",inline"`
}

// SecretInfo is the (inner) response type for ListSecrets calls
type SecretInfo struct {
	deployment.SecretInfo `yaml:",inline"`
}

// ScaleFunctionOptions are the options for the ScaleFunction call
type ScaleFunctionOptions struct {
	deployment.ScaleServiceOptions `yaml:",inline"`
}

// InvokeOptions are the options for the Invoke call
type InvokeOptions struct {
	GatewayAddress     string
	FunctionExpression string
	Input              io.Reader
	Output             io.Writer
}

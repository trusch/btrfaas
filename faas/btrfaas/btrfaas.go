package btrfaas

import (
	"context"
	"errors"
	"strings"

	g "google.golang.org/grpc"

	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/faas"
	"github.com/trusch/btrfaas/fgateway/grpc"
)

// BtrFaaS is the btrfaas implementation of the FaaS interface
type BtrFaaS struct {
	platform deployment.Platform
}

// New creates a new FaaS instance
func New(platform deployment.Platform) faas.FaaS {
	return &BtrFaaS{platform}
}

// Init initializes the FaaS
func (ptr *BtrFaaS) Init(ctx context.Context, options *faas.InitOptions) error {
	if err := ptr.platform.PrepareEnvironment(ctx, &options.PrepareEnvironmentOptions); err != nil {
		return err
	}
	return ptr.platform.DeployService(ctx, &deployment.DeployServiceOptions{
		EnvironmentID: options.PrepareEnvironmentOptions.ID,
		ID:            "fgateway",
		Image:         "btrfaas/fgateway",
		Ports: map[uint16]uint16{
			2424: 2424,
		},
	})
}

// Teardown cleans the FaaS completely
func (ptr *BtrFaaS) Teardown(ctx context.Context, options *faas.TeardownOptions) error {
	return ptr.platform.TeardownEnvironment(ctx, &options.TeardownEnvironmentOptions)
}

// DeployFunction deploys a service in an environment
func (ptr *BtrFaaS) DeployFunction(ctx context.Context, options *faas.DeployFunctionOptions) error {
	if options.DeployServiceOptions.Labels == nil {
		options.DeployServiceOptions.Labels = make(map[string]string)
	}
	options.DeployServiceOptions.Labels["btrfaas.function"] = "true"
	return ptr.platform.DeployService(ctx, &options.DeployServiceOptions)
}

// UndeployFunction unddeploys a service from an environment
func (ptr *BtrFaaS) UndeployFunction(ctx context.Context, options *faas.UndeployFunctionOptions) error {
	return ptr.platform.UndeployService(ctx, &options.UndeployServiceOptions)

}

// ListFunctions returns a list of all deployed services
func (ptr *BtrFaaS) ListFunctions(ctx context.Context, options *faas.ListFunctionsOptions) ([]*faas.FunctionInfo, error) {
	if options.Labels == nil {
		options.Labels = make(map[string]string)
	}
	options.Labels["btrfaas.function"] = "true"
	infos, err := ptr.platform.ListServices(ctx, &options.ListServicesOptions)
	if err != nil {
		return nil, err
	}
	res := make([]*faas.FunctionInfo, len(infos))
	for idx, info := range infos {
		res[idx] = &faas.FunctionInfo{ServiceInfo: *info}
	}
	return res, nil
}

// ScaleFunction scales the service
func (ptr *BtrFaaS) ScaleFunction(ctx context.Context, options *faas.ScaleFunctionOptions) error {
	return ptr.platform.ScaleService(ctx, &options.ScaleServiceOptions)
}

// DeploySecret deploys a secret in an environment
func (ptr *BtrFaaS) DeploySecret(ctx context.Context, options *faas.DeploySecretOptions) error {
	return ptr.platform.DeploySecret(ctx, &options.DeploySecretOptions)
}

// UndeploySecret unddeploys a secret from an environment
func (ptr *BtrFaaS) UndeploySecret(ctx context.Context, options *faas.UndeploySecretOptions) error {
	return ptr.platform.UndeploySecret(ctx, &options.UndeploySecretOptions)
}

// ListSecrets returns a list of all deployed secrets
func (ptr *BtrFaaS) ListSecrets(ctx context.Context, options *faas.ListSecretsOptions) ([]*faas.SecretInfo, error) {
	infos, err := ptr.platform.ListSecrets(ctx, &options.ListSecretsOptions)
	if err != nil {
		return nil, err
	}
	res := make([]*faas.SecretInfo, len(infos))
	for idx, info := range infos {
		res[idx] = &faas.SecretInfo{SecretInfo: *info}
	}
	return res, nil
}

// Invoke calls a function
func (ptr *BtrFaaS) Invoke(ctx context.Context, options *faas.InvokeOptions) error {
	chain, opts, err := createCallRequest(options.FunctionExpression)
	if err != nil {
		return err
	}
	cli, err := grpc.NewClient(options.GatewayAddress, g.WithInsecure())
	if err != nil {
		return err
	}
	return cli.Run(ctx, chain, opts, options.Input, options.Output)
}

func createCallRequest(expr string) (chain []string, opts []map[string]string, err error) {
	fnExpressions := strings.Split(expr, "|")
	chain = make([]string, len(fnExpressions))
	opts = make([]map[string]string, len(fnExpressions))
	for idx, fnExpression := range fnExpressions {
		parts := strings.Split(strings.Trim(fnExpression, " "), " ")
		if len(parts) < 1 {
			return nil, nil, errors.New("malformed expression")
		}
		fn := strings.Trim(parts[0], " ")
		chain[idx] = fn
		fnOpts := make(map[string]string)
		for i := 1; i < len(parts); i++ {
			pairSlice := strings.Split(parts[i], "=")
			if len(pairSlice) < 2 {
				return nil, nil, errors.New("malformed expression")
			}
			fnOpts[pairSlice[0]] = pairSlice[1]
		}
		opts[idx] = fnOpts
	}
	return chain, opts, nil
}

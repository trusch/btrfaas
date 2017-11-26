package openfaas

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/faas"
)

const trueString = "true"

// OpenFaaS is the btrfaas implementation of the FaaS interface
type OpenFaaS struct {
	platform deployment.Platform
}

// New creates a new FaaS instance
func New(platform deployment.Platform) faas.FaaS {
	return &OpenFaaS{platform}
}

// Init initializes the FaaS
func (ptr *OpenFaaS) Init(ctx context.Context, options *faas.InitOptions) error {
	if err := ptr.platform.PrepareEnvironment(ctx, &options.PrepareEnvironmentOptions); err != nil {
		return err
	}
	return ptr.platform.DeployService(ctx, &deployment.DeployServiceOptions{
		EnvironmentID: options.PrepareEnvironmentOptions.ID,
		ID:            "openfaas-gateway",
		Image:         "functions/gateway",
		Ports: []*deployment.PortConfig{
			{
				Type:          "host",
				ContainerPort: 8080,
				HostPort:      8080,
			},
		},
		Volumes: []*deployment.VolumeConfig{
			{
				Type:   "host",
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	})
}

// Teardown cleans the FaaS completely
func (ptr *OpenFaaS) Teardown(ctx context.Context, options *faas.TeardownOptions) error {
	return ptr.platform.TeardownEnvironment(ctx, &options.TeardownEnvironmentOptions)
}

// DeployFunction deploys a service in an environment
func (ptr *OpenFaaS) DeployFunction(ctx context.Context, options *faas.DeployFunctionOptions) error {
	if options.DeployServiceOptions.Labels == nil {
		options.DeployServiceOptions.Labels = make(map[string]string)
	}
	options.DeployServiceOptions.Labels["openfaas.function"] = trueString
	options.DeployServiceOptions.Labels["function"] = trueString
	if options.Ports == nil {
		options.Ports = make([]*deployment.PortConfig, 0)
	}
	options.Ports = append(options.Ports, &deployment.PortConfig{
		Type:          "cluster",
		ContainerPort: 8080,
		HostPort:      8080,
	})
	return ptr.platform.DeployService(ctx, &options.DeployServiceOptions)
}

// UndeployFunction unddeploys a service from an environment
func (ptr *OpenFaaS) UndeployFunction(ctx context.Context, options *faas.UndeployFunctionOptions) error {
	return ptr.platform.UndeployService(ctx, &options.UndeployServiceOptions)

}

// ListFunctions returns a list of all deployed services
func (ptr *OpenFaaS) ListFunctions(ctx context.Context, options *faas.ListFunctionsOptions) ([]*faas.FunctionInfo, error) {
	if options.Labels == nil {
		options.Labels = make(map[string]string)
	}
	options.Labels["openfaas.function"] = trueString
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
func (ptr *OpenFaaS) ScaleFunction(ctx context.Context, options *faas.ScaleFunctionOptions) error {
	return ptr.platform.ScaleService(ctx, &options.ScaleServiceOptions)
}

// DeploySecret deploys a secret in an environment
func (ptr *OpenFaaS) DeploySecret(ctx context.Context, options *faas.DeploySecretOptions) error {
	return ptr.platform.DeploySecret(ctx, &options.DeploySecretOptions)
}

// UndeploySecret unddeploys a secret from an environment
func (ptr *OpenFaaS) UndeploySecret(ctx context.Context, options *faas.UndeploySecretOptions) error {
	return ptr.platform.UndeploySecret(ctx, &options.UndeploySecretOptions)
}

// ListSecrets returns a list of all deployed secrets
func (ptr *OpenFaaS) ListSecrets(ctx context.Context, options *faas.ListSecretsOptions) ([]*faas.SecretInfo, error) {
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
func (ptr *OpenFaaS) Invoke(ctx context.Context, options *faas.InvokeOptions) error {
	url := fmt.Sprintf("http://%v/function/%v", options.GatewayAddress, options.FunctionExpression)
	req, _ := http.NewRequest("POST", url, options.Input)
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if _, err = io.Copy(options.Output, resp.Body); err != nil {
		return err
	}
	return nil
}

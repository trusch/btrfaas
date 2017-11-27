package btrfaas

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	g "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/faas"
	"github.com/trusch/btrfaas/fgateway/grpc"
	"github.com/trusch/btrfaas/pki"
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
	pkiManager, err := pki.NewManager(ctx, ptr.platform, options.PrepareEnvironmentOptions.ID)
	if err != nil {
		return err
	}
	if err = pkiManager.IssueServer(ctx, "fgateway"); err != nil {
		return err
	}
	if err = pkiManager.IssueClient(ctx, "client"); err != nil {
		return err
	}
	cmd := []string{}
	if deployment.Debug() {
		cmd = append(cmd, "--log-level", "debug")
	}
	return ptr.platform.DeployService(ctx, &deployment.DeployServiceOptions{
		EnvironmentID: options.PrepareEnvironmentOptions.ID,
		ID:            "fgateway",
		Image:         "btrfaas/fgateway",
		Ports: []*deployment.PortConfig{
			{
				Type:          "host",
				ContainerPort: 2424,
				HostPort:      2424,
			},
		},
		Cmd: cmd,
		Secrets: deployment.LabelSet{
			"btrfaas-ca-cert": "/run/secrets/btrfaas-ca-cert.pem",
			"fgateway-cert":   "/run/secrets/fgateway-cert.pem",
			"fgateway-key":    "/run/secrets/fgateway-key.pem",
			"client-cert":     "/run/secrets/client-cert.pem",
			"client-key":      "/run/secrets/client-key.pem",
		},
	})
}

// Teardown cleans the FaaS completely
func (ptr *BtrFaaS) Teardown(ctx context.Context, options *faas.TeardownOptions) error {
	if err := ptr.platform.TeardownEnvironment(ctx, &options.TeardownEnvironmentOptions); err != nil {
		return err
	}
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(home, ".btrfaas", options.ID))
}

// DeployFunction deploys a service in an environment
func (ptr *BtrFaaS) DeployFunction(ctx context.Context, options *faas.DeployFunctionOptions) error {
	if options.Labels == nil {
		options.DeployServiceOptions.Labels = make(map[string]string)
	}
	options.Labels["btrfaas.function"] = "true"
	pkiManager, err := pki.NewManager(ctx, ptr.platform, options.EnvironmentID)
	if err != nil {
		return err
	}
	if err := pkiManager.IssueServer(ctx, options.ID); err != nil {
		return err
	}
	if options.Secrets == nil {
		options.Secrets = make(map[string]string)
	}
	options.Secrets["btrfaas-ca-cert"] = "/run/secrets/btrfaas-ca-cert.pem"
	options.Secrets[options.ID+"-key"] = "/run/secrets/btrfaas-function-key.pem"
	options.Secrets[options.ID+"-cert"] = "/run/secrets/btrfaas-function-cert.pem"
	if options.Ports == nil {
		options.Ports = make([]*deployment.PortConfig, 0)
	}
	options.Ports = append(options.Ports, &deployment.PortConfig{
		Type:          "cluster",
		ContainerPort: 2424,
		HostPort:      2424,
	})
	options.Ports = append(options.Ports, &deployment.PortConfig{
		Type:          "cluster",
		ContainerPort: 8080,
		HostPort:      8080,
	})
	return ptr.platform.DeployService(ctx, &options.DeployServiceOptions)
}

// UndeployFunction undeploys a service from an environment
func (ptr *BtrFaaS) UndeployFunction(ctx context.Context, options *faas.UndeployFunctionOptions) error {
	if err := ptr.platform.UndeployService(ctx, &options.UndeployServiceOptions); err != nil {
		return err
	}
	if err := ptr.platform.UndeploySecret(ctx, &deployment.UndeploySecretOptions{
		EnvironmentID: options.EnvironmentID,
		ID:            options.ID + "-key",
	}); err != nil {
		return err
	}
	return ptr.platform.UndeploySecret(ctx, &deployment.UndeploySecretOptions{
		EnvironmentID: options.EnvironmentID,
		ID:            options.ID + "-cert",
	})
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
	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	ca, err := ioutil.ReadFile(filepath.Join(home, ".btrfaas", options.EnvironmentID, "ca-cert.pem"))
	if err != nil {
		return fmt.Errorf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return errors.New("failed to append ca certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName: "fgateway",
		RootCAs:    certPool,
	})
	cli, err := grpc.NewClient(options.GatewayAddress, g.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	return cli.Run(ctx, chain, opts, options.Input, options.Output)
}

func createCallRequest(expr string) (chain []string, opts [][]string, err error) {
	fnExpressions := strings.Split(expr, "|")
	chain = make([]string, len(fnExpressions))
	opts = make([][]string, len(fnExpressions))
	for idx, fnExpression := range fnExpressions {
		parts := strings.Fields(fnExpression)
		if len(parts) < 1 {
			return nil, nil, errors.New("malformed expression")
		}
		chain[idx] = parts[0]
		if len(parts) > 1 {
			opts[idx] = parts[1:]
		}
	}
	return chain, opts, nil
}

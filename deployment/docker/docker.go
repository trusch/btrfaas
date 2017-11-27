package docker

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/frunner/env"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// dockerPlatform implements deployment.Platform with the help of a docker
type dockerPlatform struct {
	cli *client.Client
}

// NewPlatform creates a new Platform instance for local docker development
func NewPlatform() (deployment.Platform, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &dockerPlatform{cli}, nil
}

// PrepareEnvironment prepares an environment to start deploying services
// This should contain all one time setup like creating namespaces/networks etc.
func (p *dockerPlatform) PrepareEnvironment(ctx context.Context, options *deployment.PrepareEnvironmentOptions) error {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	secretRoot := filepath.Join(home, ".btrfaas", options.ID, "secrets")
	if err = os.MkdirAll(secretRoot, 0755); err != nil {
		return err
	}
	name := options.ID + "_network"
	_, err = p.cli.NetworkInspect(ctx, name, false)
	if err != nil {
		_, err = p.cli.NetworkCreate(ctx, name, types.NetworkCreate{
			Driver:     "bridge",
			Attachable: true,
		})
		return err
	}
	return nil
}

// DeployService deploys a service in an environment
func (p *dockerPlatform) DeployService(ctx context.Context, options *deployment.DeployServiceOptions) error {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	secretRoot := filepath.Join(home, ".btrfaas", options.EnvironmentID, "secrets")
	netName := options.EnvironmentID + "_network"
	if options.Labels == nil {
		options.Labels = make(deployment.LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID

	containerConfig := &container.Config{
		Image:        options.Image,
		Labels:       options.Labels,
		Cmd:          options.Cmd,
		Env:          env.Env(options.Env).ToSlice(),
		ExposedPorts: constructExposedPorts(options.Ports),
	}
	ports, err := constructPortMap(options.Ports)
	if err != nil {
		return err
	}
	binds := constructSecretBinds(options.Secrets, secretRoot)
	binds = append(binds, constructVolumeBinds(options.Volumes)...)
	hostConfig := &container.HostConfig{
		AutoRemove:   !deployment.Debug(),
		PortBindings: ports,
		Binds:        binds,
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			netName: {},
		},
	}

	createResp, err := p.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, options.ID)
	if err != nil {
		return err
	}
	return p.cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{})
}

// UndeployService unddeploys a service from an environment
func (p *dockerPlatform) UndeployService(ctx context.Context, options *deployment.UndeployServiceOptions) error {
	d := 5 * time.Second
	return p.cli.ContainerStop(ctx, options.ID, &d)
}

// ListServices returns a list of all deployed services
func (p *dockerPlatform) ListServices(ctx context.Context, options *deployment.ListServicesOptions) ([]*deployment.ServiceInfo, error) {
	if options.Labels == nil {
		options.Labels = make(deployment.LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID

	args := filters.NewArgs()
	for key, val := range options.Labels {
		args.Add("label", key+"="+val)
	}

	opts := types.ContainerListOptions{
		Filters: args,
	}

	resp, err := p.cli.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*deployment.ServiceInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &deployment.ServiceInfo{
			ID:        val.Names[0][1:],
			Image:     val.Image,
			Labels:    val.Labels,
			Cmd:       strings.Split(val.Command, " "),
			CreatedAt: time.Unix(val.Created, 0),
			Endpoint:  val.NetworkSettings.Networks[options.EnvironmentID+"_network"].IPAddress,
			Scale:     1,
		}
	}
	return result, nil
}

// DeploySecret deploys a secret in an environment
func (p *dockerPlatform) DeploySecret(ctx context.Context, options *deployment.DeploySecretOptions) error {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	secretRoot := filepath.Join(home, ".btrfaas", options.EnvironmentID, "secrets")
	return ioutil.WriteFile(filepath.Join(secretRoot, options.ID), options.Value, 0600)
}

// GetSecret returns the secret value
func (p *dockerPlatform) GetSecret(ctx context.Context, options *deployment.GetSecretOptions) ([]byte, error) {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	secretRoot := filepath.Join(home, ".btrfaas", options.EnvironmentID, "secrets")
	return ioutil.ReadFile(filepath.Join(secretRoot, options.ID))
}

// UndeploySecret unddeploys a secret from an environment
func (p *dockerPlatform) UndeploySecret(ctx context.Context, options *deployment.UndeploySecretOptions) error {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	secretRoot := filepath.Join(home, ".btrfaas", options.EnvironmentID, "secrets")
	return os.Remove(filepath.Join(secretRoot, options.ID))
}

// ListSecrets returns a list of all deployed secrets
func (p *dockerPlatform) ListSecrets(ctx context.Context, options *deployment.ListSecretsOptions) ([]*deployment.SecretInfo, error) {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	secretRoot := filepath.Join(home, ".btrfaas", options.EnvironmentID, "secrets")
	resp, err := ioutil.ReadDir(secretRoot)
	if err != nil {
		return nil, err
	}
	result := make([]*deployment.SecretInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &deployment.SecretInfo{
			ID: val.Name(),
		}
	}
	return result, nil
}

func constructSecretBinds(secrets deployment.LabelSet, secretRoot string) []string {
	res := make([]string, len(secrets))
	idx := 0
	for secretID, path := range secrets {
		res[idx] = filepath.Join(secretRoot, secretID) + ":" + path
		idx++
	}
	return res
}

// ScaleService scales the service
func (p *dockerPlatform) ScaleService(ctx context.Context, options *deployment.ScaleServiceOptions) error {
	return errors.New("not supported in plain docker")
}

// TeardownEnvironment cleans the environment completely
func (p *dockerPlatform) TeardownEnvironment(ctx context.Context, options *deployment.TeardownEnvironmentOptions) error {
	services, err := p.ListServices(ctx, &deployment.ListServicesOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		log.Print("can not list services")
		return err
	}
	for _, service := range services {
		if err = p.UndeployService(ctx, &deployment.UndeployServiceOptions{
			EnvironmentID: options.ID,
			ID:            service.ID,
		}); err != nil {
			log.Printf("can not undeploy %v", service.ID)
			return err
		}
	}
	secrets, err := p.ListSecrets(ctx, &deployment.ListSecretsOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		log.Print("can not list secrets")
		return err
	}
	for _, secret := range secrets {
		if err = p.UndeploySecret(ctx, &deployment.UndeploySecretOptions{
			EnvironmentID: options.ID,
			ID:            secret.ID,
		}); err != nil {
			log.Printf("can not undeploy %v", secret.ID)
			return err
		}
	}
	return p.cli.NetworkRemove(ctx, options.ID+"_network")
}

func constructPortMap(ports []*deployment.PortConfig) (nat.PortMap, error) {
	specs := make([]string, len(ports))
	i := 0
	for _, portConfig := range ports {
		specs[i] = fmt.Sprintf("0.0.0.0:%v:%v", portConfig.HostPort, portConfig.ContainerPort)
		i++
	}
	_, res, err := nat.ParsePortSpecs(specs)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func constructExposedPorts(ports []*deployment.PortConfig) nat.PortSet {
	res := make(nat.PortSet)
	for _, portConfig := range ports {
		if portConfig.Type == "host" {
			res[nat.Port(fmt.Sprintf("%v/tcp", portConfig.ContainerPort))] = struct{}{}
		}
	}
	return res
}

func constructVolumeBinds(volumes []*deployment.VolumeConfig) []string {
	var res []string
	for _, cfg := range volumes {
		if cfg.Type == "host" {
			res = append(res, cfg.Source+":"+cfg.Target)
		}
	}
	return res
}

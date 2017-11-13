package deployment

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

	"github.com/trusch/btrfaas/frunner/env"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const secretRoot = "/var/lib/btrfaas/secrets/"

// DockerPlatform implements deployment.Platform with the help of a docker
type DockerPlatform struct {
	cli *client.Client
}

// NewDockerPlatform creates a new Platform instance for local docker development
func NewDockerPlatform() (Platform, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &DockerPlatform{cli}, nil
}

// PrepareEnvironment prepares an environment to start deploying services
// This should contain all one time setup like creating namespaces/networks etc.
func (p *DockerPlatform) PrepareEnvironment(ctx context.Context, options *PrepareEnvironmentOptions) error {
	if err := os.MkdirAll(secretRoot, 0755); err != nil {
		return err
	}
	name := options.ID + "_network"
	_, err := p.cli.NetworkInspect(ctx, name, false)
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
func (p *DockerPlatform) DeployService(ctx context.Context, options *DeployServiceOptions) error {
	netName := options.EnvironmentID + "_network"
	if options.Labels == nil {
		options.Labels = make(LabelSet)
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
	binds := constructSecretBinds(options.Secrets)
	binds = append(binds, constructVolumeBinds(options.Volumes)...)
	hostConfig := &container.HostConfig{
		AutoRemove:   true,
		PortBindings: ports,
		Binds:        binds,
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			netName: &network.EndpointSettings{},
		},
	}

	createResp, err := p.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, options.ID)
	if err != nil {
		return err
	}
	if err = p.cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

// UndeployService unddeploys a service from an environment
func (p *DockerPlatform) UndeployService(ctx context.Context, options *UndeployServiceOptions) error {
	d := 5 * time.Second
	return p.cli.ContainerStop(ctx, options.ID, &d)
}

// ListServices returns a list of all deployed services
func (p *DockerPlatform) ListServices(ctx context.Context, options *ListServicesOptions) ([]*ServiceInfo, error) {
	if options.Labels == nil {
		options.Labels = make(LabelSet)
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
	result := make([]*ServiceInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &ServiceInfo{
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
func (p *DockerPlatform) DeploySecret(ctx context.Context, options *DeploySecretOptions) error {
	return ioutil.WriteFile(filepath.Join(secretRoot, options.ID), []byte(options.Value), 0600)
}

// UndeploySecret unddeploys a secret from an environment
func (p *DockerPlatform) UndeploySecret(ctx context.Context, options *UndeploySecretOptions) error {
	return os.Remove(filepath.Join(secretRoot, options.ID))
}

// ListSecrets returns a list of all deployed secrets
func (p *DockerPlatform) ListSecrets(ctx context.Context, options *ListSecretsOptions) ([]*SecretInfo, error) {
	resp, err := ioutil.ReadDir(secretRoot)
	if err != nil {
		return nil, err
	}
	result := make([]*SecretInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &SecretInfo{
			ID: val.Name(),
		}
	}
	return result, nil
}

func constructSecretBinds(secrets LabelSet) []string {
	res := make([]string, len(secrets))
	idx := 0
	for secretID, path := range secrets {
		res[idx] = filepath.Join(secretRoot, secretID) + ":" + path
		idx++
	}
	return res
}

// ScaleService scales the service
func (p *DockerPlatform) ScaleService(ctx context.Context, options *ScaleServiceOptions) error {
	return errors.New("not supported in plain docker")
}

// TeardownEnvironment cleans the environment completely
func (p *DockerPlatform) TeardownEnvironment(ctx context.Context, options *TeardownEnvironmentOptions) error {
	services, err := p.ListServices(ctx, &ListServicesOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		log.Print("can not list services")
		return err
	}
	for _, service := range services {
		if err = p.UndeployService(ctx, &UndeployServiceOptions{
			EnvironmentID: options.ID,
			ID:            service.ID,
		}); err != nil {
			log.Printf("can not undeploy %v", service.ID)
			return err
		}
	}
	secrets, err := p.ListSecrets(ctx, &ListSecretsOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		log.Print("can not list secrets")
		return err
	}
	for _, secret := range secrets {
		if err = p.UndeploySecret(ctx, &UndeploySecretOptions{
			EnvironmentID: options.ID,
			ID:            secret.ID,
		}); err != nil {
			log.Printf("can not undeploy %v", secret.ID)
			return err
		}
	}
	return p.cli.NetworkRemove(ctx, options.ID+"_network")
}

func constructPortMap(ports map[uint16]uint16) (nat.PortMap, error) {
	specs := make([]string, len(ports))
	i := 0
	for k, v := range ports {
		specs[i] = fmt.Sprintf("0.0.0.0:%v:%v", k, v)
		i++
	}
	_, res, err := nat.ParsePortSpecs(specs)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func constructExposedPorts(ports map[uint16]uint16) nat.PortSet {
	res := make(nat.PortSet)
	for _, containerPort := range ports {
		res[nat.Port(fmt.Sprintf("%v/tcp", containerPort))] = struct{}{}
	}
	return res
}

func constructVolumeBinds(volumes []*VolumeConfig) []string {
	var res []string
	for _, cfg := range volumes {
		if cfg.Type == "host" {
			res = append(res, cfg.Source+":"+cfg.Target)
		}
	}
	return res
}

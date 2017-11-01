package deployment

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/trusch/btrfaas/frunner/env"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// DockerPlatform implements deployment.Platform with the help of a local docker connection
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
	name := options.EnvironmentID + "_network"
	_, err := p.cli.NetworkInspect(ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		log.Debug("network not found, creating new")
		_, err = p.cli.NetworkCreate(ctx, name, types.NetworkCreate{Driver: "overlay"})
		return err
	}
	return nil
}

// DeployService deploys a service in an environment
func (p *DockerPlatform) DeployService(ctx context.Context, options *DeployServiceOptions) error {
	netName := options.EnvironmentID + "_network"

	service := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: options.ServiceID,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   options.Image,
				Command: options.Cmd,
				Env:     env.Env(options.Env).ToSlice(),
				Secrets: p.constructSecretReferences(ctx, options.Secrets),
			},
			Networks: []swarm.NetworkAttachmentConfig{
				swarm.NetworkAttachmentConfig{Target: netName},
			},
		},
	}
	_, err := p.cli.ServiceCreate(ctx, service, types.ServiceCreateOptions{})
	return err
}

// UndeployService unddeploys a service from an environment
func (p *DockerPlatform) UndeployService(ctx context.Context, options *UndeployServiceOptions) error {
	return p.cli.ServiceRemove(ctx, options.ServiceID)
}

// ListServices returns a list of all deployed services
func (p *DockerPlatform) ListServices(ctx context.Context, options *ListServicesOptions) ([]*ServiceInfo, error) {
	resp, err := p.cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}
	result := make([]*ServiceInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &ServiceInfo{
			DeployServiceOptions: DeployServiceOptions{
				ServiceID: val.Spec.Name,
				Image:     val.Spec.TaskTemplate.ContainerSpec.Image,
			},
			CreatedAt: val.CreatedAt,
		}
	}
	return result, nil
}

// DeploySecret deploys a secret in an environment
func (p *DockerPlatform) DeploySecret(ctx context.Context, options *DeploySecretOptions) error {
	_, err := p.cli.SecretCreate(ctx, swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name: options.SecretID,
		},
		Data: []byte(options.Value),
	})
	return err
}

// UndeploySecret unddeploys a secret from an environment
func (p *DockerPlatform) UndeploySecret(ctx context.Context, options *UndeploySecretOptions) error {
	return p.cli.SecretRemove(ctx, options.SecretID)
}

func (p *DockerPlatform) constructSecretReferences(ctx context.Context, list []string) []*swarm.SecretReference {
	res := make([]*swarm.SecretReference, len(list))
	secrets, err := p.cli.SecretList(ctx, types.SecretListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for idx, id := range list {
		res[idx] = &swarm.SecretReference{
			SecretName: id,
			File: &swarm.SecretReferenceFileTarget{
				Name: "/run/secret/" + id,
				UID:  "0",
				GID:  "0",
				Mode: 0600,
			},
		}
		for _, s := range secrets {
			if s.Spec.Name == id {
				res[idx].SecretID = s.ID
			}
		}
	}
	return res
}

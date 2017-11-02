package deployment

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trusch/btrfaas/frunner/env"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// SwarmPlatform implements deployment.Platform with the help of a docker swarm
type SwarmPlatform struct {
	cli *client.Client
}

// NewSwarmPlatform creates a new Platform instance for local docker development
func NewSwarmPlatform() (Platform, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &SwarmPlatform{cli}, nil
}

// PrepareEnvironment prepares an environment to start deploying services
// This should contain all one time setup like creating namespaces/networks etc.
func (p *SwarmPlatform) PrepareEnvironment(ctx context.Context, options *PrepareEnvironmentOptions) error {
	name := options.ID + "_network"
	_, err := p.cli.NetworkInspect(ctx, name, false)
	if err != nil {
		log.Debug("network not found, creating new")
		_, err = p.cli.NetworkCreate(ctx, name, types.NetworkCreate{
			Driver:     "overlay",
			Attachable: true,
		})
		return err
	}
	return nil
}

// DeployService deploys a service in an environment
func (p *SwarmPlatform) DeployService(ctx context.Context, options *DeployServiceOptions) error {
	netName := options.EnvironmentID + "_network"
	if options.Labels == nil {
		options.Labels = make(LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID

	secrets, err := p.constructSecretReferences(ctx, options.Secrets)
	if err != nil {
		return err
	}
	service := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   options.ID,
			Labels: options.Labels,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image:   options.Image,
				Labels:  options.Labels,
				Command: options.Cmd,
				Env:     env.Env(options.Env).ToSlice(),
				Secrets: secrets,
			},
			Networks: []swarm.NetworkAttachmentConfig{
				swarm.NetworkAttachmentConfig{Target: netName},
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeVIP,
			Ports: createPortConfigs(options.Ports),
		},
	}
	_, err = p.cli.ServiceCreate(ctx, service, types.ServiceCreateOptions{})
	return err
}

// UndeployService unddeploys a service from an environment
func (p *SwarmPlatform) UndeployService(ctx context.Context, options *UndeployServiceOptions) error {
	return p.cli.ServiceRemove(ctx, options.ID)
}

// ListServices returns a list of all deployed services
func (p *SwarmPlatform) ListServices(ctx context.Context, options *ListServicesOptions) ([]*ServiceInfo, error) {
	if options.Labels == nil {
		options.Labels = make(LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID

	args := filters.NewArgs()
	for key, val := range options.Labels {
		args.Add("label", key+"="+val)
	}

	// search swarm services
	opts := types.ServiceListOptions{
		Filters: args,
	}
	resp, err := p.cli.ServiceList(ctx, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*ServiceInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &ServiceInfo{
			ID:        val.Spec.Name,
			Image:     val.Spec.TaskTemplate.ContainerSpec.Image,
			Labels:    val.Spec.Labels,
			Cmd:       val.Spec.TaskTemplate.ContainerSpec.Command,
			Env:       envToLabelSet(val.Spec.TaskTemplate.ContainerSpec.Env),
			Secrets:   secretListToLabelSet(val.Spec.TaskTemplate.ContainerSpec.Secrets),
			CreatedAt: val.CreatedAt,
			Endpoint:  val.Endpoint.VirtualIPs[0].Addr,
			Scale:     *val.Spec.Mode.Replicated.Replicas,
		}
	}
	return result, nil
}

// DeploySecret deploys a secret in an environment
func (p *SwarmPlatform) DeploySecret(ctx context.Context, options *DeploySecretOptions) error {
	if options.Labels == nil {
		options.Labels = make(LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID
	_, err := p.cli.SecretCreate(ctx, swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name:   options.ID,
			Labels: options.Labels,
		},
		Data: []byte(options.Value),
	})
	return err
}

// UndeploySecret unddeploys a secret from an environment
func (p *SwarmPlatform) UndeploySecret(ctx context.Context, options *UndeploySecretOptions) error {
	return p.cli.SecretRemove(ctx, options.ID)
}

// ListSecrets returns a list of all deployed secrets
func (p *SwarmPlatform) ListSecrets(ctx context.Context, options *ListSecretsOptions) ([]*SecretInfo, error) {
	if options.Labels == nil {
		options.Labels = make(LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID

	args := filters.NewArgs()
	for key, val := range options.Labels {
		args.Add("label", key+"="+val)
	}

	// search swarm services
	opts := types.SecretListOptions{
		Filters: args,
	}
	resp, err := p.cli.SecretList(ctx, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*SecretInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &SecretInfo{
			ID:     val.Spec.Name,
			Labels: val.Spec.Labels,
		}
	}
	return result, nil
}

func (p *SwarmPlatform) constructSecretReferences(ctx context.Context, list map[string]string) ([]*swarm.SecretReference, error) {
	res := make([]*swarm.SecretReference, len(list))
	secrets, err := p.cli.SecretList(ctx, types.SecretListOptions{})
	if err != nil {
		return nil, err
	}
	idx := 0
	for secretID, path := range list {
		res[idx] = &swarm.SecretReference{
			SecretName: secretID,
			File: &swarm.SecretReferenceFileTarget{
				Name: path,
				UID:  "0",
				GID:  "0",
				Mode: 0600,
			},
		}
		for _, s := range secrets {
			if s.Spec.Name == secretID {
				res[idx].SecretID = s.ID
			}
		}
		idx++
	}
	return res, nil
}

// ScaleService scales the service
func (p *SwarmPlatform) ScaleService(ctx context.Context, options *ScaleServiceOptions) error {
	service, _, err := p.cli.ServiceInspectWithRaw(ctx, options.ServiceID, types.ServiceInspectOptions{})
	if err != nil {
		return err
	}
	spec := service.Spec
	spec.Mode.Replicated.Replicas = &options.Scale
	_, err = p.cli.ServiceUpdate(ctx, options.ServiceID, service.Version, spec, types.ServiceUpdateOptions{})
	return err
}

// TeardownEnvironment cleans the environment completely
func (p *SwarmPlatform) TeardownEnvironment(ctx context.Context, options *TeardownEnvironmentOptions) error {
	services, err := p.ListServices(ctx, &ListServicesOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		return err
	}
	for _, service := range services {
		if err = p.UndeployService(ctx, &UndeployServiceOptions{
			EnvironmentID: options.ID,
			ID:            service.ID,
		}); err != nil {
			return err
		}
	}
	secrets, err := p.ListSecrets(ctx, &ListSecretsOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		if err = p.UndeploySecret(ctx, &UndeploySecretOptions{
			EnvironmentID: options.ID,
			ID:            secret.ID,
		}); err != nil {
			return err
		}
	}
	return p.cli.NetworkRemove(ctx, options.ID+"_network")
}

func envToLabelSet(env []string) LabelSet {
	res := make(LabelSet)
	for _, val := range env {
		parts := strings.Split(val, "=")
		if len(parts) > 1 {
			res[parts[0]] = parts[1]
		}
	}
	return res
}

func secretListToLabelSet(secrets []*swarm.SecretReference) LabelSet {
	res := make(LabelSet)
	for _, val := range secrets {
		res[val.SecretName] = val.File.Name
	}
	return res
}

func createPortConfigs(portmap map[uint16]uint16) []swarm.PortConfig {
	res := make([]swarm.PortConfig, len(portmap))
	i := 0
	for hostPort, containerPort := range portmap {
		res[i] = swarm.PortConfig{
			Protocol:      swarm.PortConfigProtocolTCP,
			TargetPort:    uint32(containerPort),
			PublishedPort: uint32(hostPort),
			PublishMode:   swarm.PortConfigPublishModeIngress,
		}
	}
	return res
}

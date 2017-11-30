package swarm

import (
	"context"
	"errors"
	"strings"

	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/frunner/env"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// swarmPlatform implements deployment.Platform with the help of a docker swarm
type swarmPlatform struct {
	cli *client.Client
}

// NewPlatform creates a new Platform instance for local docker development
func NewPlatform() (deployment.Platform, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &swarmPlatform{cli}, nil
}

// PrepareEnvironment prepares an environment to start deploying services
// This should contain all one time setup like creating namespaces/networks etc.
func (p *swarmPlatform) PrepareEnvironment(ctx context.Context, options *deployment.PrepareEnvironmentOptions) error {
	name := options.ID + "_network"
	_, err := p.cli.NetworkInspect(ctx, name, false)
	if err != nil {
		_, err = p.cli.NetworkCreate(ctx, name, types.NetworkCreate{
			Driver:     "overlay",
			Attachable: true,
		})
		return err
	}
	return nil
}

// DeployService deploys a service in an environment
func (p *swarmPlatform) DeployService(ctx context.Context, options *deployment.DeployServiceOptions) error {
	netName := options.EnvironmentID + "_network"
	if options.Labels == nil {
		options.Labels = make(deployment.LabelSet)
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
				Mounts:  createMounts(options.Volumes),
			},
			Networks: []swarm.NetworkAttachmentConfig{
				{Target: netName},
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeDNSRR,
			Ports: createPortConfigs(options.Ports),
		},
	}
	_, err = p.cli.ServiceCreate(ctx, service, types.ServiceCreateOptions{})
	return err
}

// UndeployService unddeploys a service from an environment
func (p *swarmPlatform) UndeployService(ctx context.Context, options *deployment.UndeployServiceOptions) error {
	return p.cli.ServiceRemove(ctx, options.ID)
}

// ListServices returns a list of all deployed services
func (p *swarmPlatform) ListServices(ctx context.Context, options *deployment.ListServicesOptions) ([]*deployment.ServiceInfo, error) {
	if options.Labels == nil {
		options.Labels = make(deployment.LabelSet)
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
	result := make([]*deployment.ServiceInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &deployment.ServiceInfo{
			ID:        val.Spec.Name,
			Image:     val.Spec.TaskTemplate.ContainerSpec.Image,
			Labels:    val.Spec.Labels,
			Cmd:       val.Spec.TaskTemplate.ContainerSpec.Command,
			Env:       envToLabelSet(val.Spec.TaskTemplate.ContainerSpec.Env),
			Secrets:   secretListToLabelSet(val.Spec.TaskTemplate.ContainerSpec.Secrets),
			CreatedAt: val.CreatedAt,
			// Endpoint:  val.Endpoint.VirtualIPs[0].Addr,
			Scale: *val.Spec.Mode.Replicated.Replicas,
		}
	}
	return result, nil
}

// DeploySecret deploys a secret in an environment
func (p *swarmPlatform) DeploySecret(ctx context.Context, options *deployment.DeploySecretOptions) error {
	if options.Labels == nil {
		options.Labels = make(deployment.LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID
	_, err := p.cli.SecretCreate(ctx, swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name:   options.ID,
			Labels: options.Labels,
		},
		Data: options.Value,
	})
	return err
}

// UndeploySecret unddeploys a secret from an environment
func (p *swarmPlatform) UndeploySecret(ctx context.Context, options *deployment.UndeploySecretOptions) error {
	return p.cli.SecretRemove(ctx, options.ID)
}

// GetSecret unddeploys a secret from an environment
func (p *swarmPlatform) GetSecret(ctx context.Context, options *deployment.GetSecretOptions) ([]byte, error) {
	args := filters.NewArgs()
	args.Add("label", "btrfaas_env="+options.EnvironmentID)
	opts := types.SecretListOptions{
		Filters: args,
	}
	resp, err := p.cli.SecretList(ctx, opts)
	if err != nil {
		return nil, err
	}
	for _, secret := range resp {
		if secret.Spec.Name == options.ID {
			return secret.Spec.Data, nil
		}
	}
	return nil, errors.New("no such secret")
}

// ListSecrets returns a list of all deployed secrets
func (p *swarmPlatform) ListSecrets(ctx context.Context, options *deployment.ListSecretsOptions) ([]*deployment.SecretInfo, error) {
	if options.Labels == nil {
		options.Labels = make(deployment.LabelSet)
	}
	options.Labels["btrfaas_env"] = options.EnvironmentID

	args := filters.NewArgs()
	for key, val := range options.Labels {
		args.Add("label", key+"="+val)
	}

	// search swarm secrets
	opts := types.SecretListOptions{
		Filters: args,
	}
	resp, err := p.cli.SecretList(ctx, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*deployment.SecretInfo, len(resp))
	for idx, val := range resp {
		result[idx] = &deployment.SecretInfo{
			ID:     val.Spec.Name,
			Labels: val.Spec.Labels,
		}
	}
	return result, nil
}

func (p *swarmPlatform) constructSecretReferences(ctx context.Context, list map[string]string) ([]*swarm.SecretReference, error) {
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
func (p *swarmPlatform) ScaleService(ctx context.Context, options *deployment.ScaleServiceOptions) error {
	service, _, err := p.cli.ServiceInspectWithRaw(ctx, options.ID, types.ServiceInspectOptions{})
	if err != nil {
		return err
	}
	spec := service.Spec
	spec.Mode.Replicated.Replicas = &options.Scale
	_, err = p.cli.ServiceUpdate(ctx, options.ID, service.Version, spec, types.ServiceUpdateOptions{})
	return err
}

// TeardownEnvironment cleans the environment completely
func (p *swarmPlatform) TeardownEnvironment(ctx context.Context, options *deployment.TeardownEnvironmentOptions) error {
	services, err := p.ListServices(ctx, &deployment.ListServicesOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		return err
	}
	for _, service := range services {
		if err = p.UndeployService(ctx, &deployment.UndeployServiceOptions{
			EnvironmentID: options.ID,
			ID:            service.ID,
		}); err != nil {
			return err
		}
	}
	secrets, err := p.ListSecrets(ctx, &deployment.ListSecretsOptions{
		EnvironmentID: options.ID,
	})
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		if err = p.UndeploySecret(ctx, &deployment.UndeploySecretOptions{
			EnvironmentID: options.ID,
			ID:            secret.ID,
		}); err != nil {
			return err
		}
	}
	return p.cli.NetworkRemove(ctx, options.ID+"_network")
}

func envToLabelSet(env []string) deployment.LabelSet {
	res := make(deployment.LabelSet)
	for _, val := range env {
		parts := strings.Split(val, "=")
		if len(parts) > 1 {
			res[parts[0]] = parts[1]
		}
	}
	return res
}

func secretListToLabelSet(secrets []*swarm.SecretReference) deployment.LabelSet {
	res := make(deployment.LabelSet)
	for _, val := range secrets {
		res[val.SecretName] = val.File.Name
	}
	return res
}

func createPortConfigs(configs []*deployment.PortConfig) []swarm.PortConfig {
	res := make([]swarm.PortConfig, len(configs))
	i := 0
	for _, cfg := range configs {
		if cfg.Type == "host" {
			res[i] = swarm.PortConfig{
				Protocol:      swarm.PortConfigProtocolTCP,
				TargetPort:    uint32(cfg.ContainerPort),
				PublishedPort: uint32(cfg.HostPort),
				PublishMode:   swarm.PortConfigPublishModeHost,
			}
			i++
		}
	}
	return res[:i]
}

func createMounts(volumes []*deployment.VolumeConfig) []mount.Mount {
	var mounts []mount.Mount
	for _, cfg := range volumes {
		if cfg.Type == "host" {
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: cfg.Source,
				Target: cfg.Target,
			})
		}
	}
	return mounts
}

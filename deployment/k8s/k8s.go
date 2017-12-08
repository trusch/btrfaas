package k8s

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/trusch/btrfaas/deployment"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

// k8sPlatform implements deployment.Platform with the help of a kubernetes
type k8sPlatform struct {
	cli *kubernetes.Clientset
}

// NewPlatform creates a new Platform instance for local docker development
func NewPlatform() (deployment.Platform, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return &k8sPlatform{clientset}, nil
}

// PrepareEnvironment prepares an environment to start deploying services
// This should contain all one time setup like creating namespaces/networks etc.
func (p *k8sPlatform) PrepareEnvironment(ctx context.Context, options *deployment.PrepareEnvironmentOptions) error {
	nsClient := p.cli.CoreV1().Namespaces()
	ns := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: options.ID}}
	_, err := nsClient.Create(ns)
	return err
}

// DeployService deploys a service in an environment
func (p *k8sPlatform) DeployService(ctx context.Context, options *deployment.DeployServiceOptions) error {
	deploymentsClient := p.cli.AppsV1beta1().Deployments(options.EnvironmentID)
	if options.Labels == nil {
		options.Labels = make(map[string]string)
	}
	options.Labels["name"] = options.ID
	one := int32(1)
	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.ID,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &one,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: options.Labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:         options.ID,
							Image:        options.Image,
							Ports:        constructContainerPorts(options.Ports),
							VolumeMounts: constructVolumeMounts(options),
							Env:          constructEnv(options),
							Command:      options.Cmd,
						},
					},
					Volumes: constructVolumes(options),
				},
			},
		},
	}
	if _, err := deploymentsClient.Create(deployment); err != nil {
		return err
	}

	serviceClient := p.cli.CoreV1().Services(options.EnvironmentID)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.ID,
		},
		Spec: apiv1.ServiceSpec{
			Ports:    constructServicePorts(options.Ports),
			Selector: map[string]string{"name": options.ID},
			Type:     getServiceType(options.Ports),
		},
	}
	if service.Spec.Type == apiv1.ServiceTypeClusterIP {
		service.Spec.ClusterIP = "None"
	}
	_, err := serviceClient.Create(service)
	return err
}

// UndeployService unddeploys a service from an environment
func (p *k8sPlatform) UndeployService(ctx context.Context, options *deployment.UndeployServiceOptions) error {
	deploymentsClient := p.cli.AppsV1beta1().Deployments(options.EnvironmentID)
	if err := deploymentsClient.Delete(options.ID, &metav1.DeleteOptions{}); err != nil {
		return err
	}
	servicesClient := p.cli.CoreV1().Services(options.EnvironmentID)
	return servicesClient.Delete(options.ID, &metav1.DeleteOptions{})
}

// ListServices returns a list of all deployed services
func (p *k8sPlatform) ListServices(ctx context.Context, options *deployment.ListServicesOptions) ([]*deployment.ServiceInfo, error) {
	deploymentsClient := p.cli.AppsV1beta1().Deployments(options.EnvironmentID)
	list, err := deploymentsClient.List(metav1.ListOptions{
		LabelSelector: buildLabelSelector(options.Labels),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*deployment.ServiceInfo, len(list.Items))
	for idx, depl := range list.Items {
		res[idx] = &deployment.ServiceInfo{
			ID:        depl.Name,
			Image:     depl.Spec.Template.Spec.Containers[0].Image,
			Labels:    depl.Labels,
			Cmd:       depl.Spec.Template.Spec.Containers[0].Command,
			Scale:     uint64(*depl.Spec.Replicas),
			CreatedAt: depl.CreationTimestamp.Time,
		}
	}
	return res, nil
}

// DeploySecret deploys a secret in an environment
func (p *k8sPlatform) DeploySecret(ctx context.Context, options *deployment.DeploySecretOptions) error {
	secretClient := p.cli.CoreV1().Secrets(options.EnvironmentID)
	_, err := secretClient.Create(&apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.ID,
		},
		Data: map[string][]byte{"value": options.Value},
	})
	return err
}

// GetSecret returns the secret value
func (p *k8sPlatform) GetSecret(ctx context.Context, options *deployment.GetSecretOptions) ([]byte, error) {
	secretClient := p.cli.CoreV1().Secrets(options.EnvironmentID)
	secret, err := secretClient.Get(options.ID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret.Data["value"], nil
}

// UndeploySecret unddeploys a secret from an environment
func (p *k8sPlatform) UndeploySecret(ctx context.Context, options *deployment.UndeploySecretOptions) error {
	secretClient := p.cli.CoreV1().Secrets(options.EnvironmentID)
	return secretClient.Delete(options.ID, &metav1.DeleteOptions{})
}

// ListSecrets returns a list of all deployed secrets
func (p *k8sPlatform) ListSecrets(ctx context.Context, options *deployment.ListSecretsOptions) ([]*deployment.SecretInfo, error) {
	secretClient := p.cli.CoreV1().Secrets(options.EnvironmentID)
	list, err := secretClient.List(metav1.ListOptions{
		LabelSelector: buildLabelSelector(options.Labels),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*deployment.SecretInfo, len(list.Items))
	for idx, secret := range list.Items {
		res[idx] = &deployment.SecretInfo{
			ID:     secret.Name,
			Labels: secret.Labels,
		}
	}
	return res, nil
}

// ScaleService scales the service
func (p *k8sPlatform) ScaleService(ctx context.Context, options *deployment.ScaleServiceOptions) error {
	deploymentsClient := p.cli.AppsV1beta1().Deployments(options.EnvironmentID)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := deploymentsClient.Get(options.ID, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		scale := int32(options.Scale)
		result.Spec.Replicas = &scale
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
}

// TeardownEnvironment cleans the environment completely
func (p *k8sPlatform) TeardownEnvironment(ctx context.Context, options *deployment.TeardownEnvironmentOptions) error {
	nsClient := p.cli.CoreV1().Namespaces()
	return nsClient.Delete(options.ID, &metav1.DeleteOptions{})
}

func constructContainerPorts(ports []*deployment.PortConfig) []apiv1.ContainerPort {
	res := make([]apiv1.ContainerPort, 0)
	for _, cfg := range ports {
		res = append(res, apiv1.ContainerPort{
			ContainerPort: int32(cfg.Container),
		})
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func constructServicePorts(ports []*deployment.PortConfig) []apiv1.ServicePort {
	res := make([]apiv1.ServicePort, 0)
	for _, cfg := range ports {
		res = append(res, apiv1.ServicePort{
			Name: fmt.Sprintf("port-%v", cfg.Container),
			Port: int32(cfg.Host),
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(cfg.Container),
			},
		})
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func getServiceType(ports []*deployment.PortConfig) apiv1.ServiceType {
	res := apiv1.ServiceTypeClusterIP
	for _, cfg := range ports {
		if cfg.Type == "host" {
			res = apiv1.ServiceTypeNodePort
			return res
		}
	}
	return res
}

func constructVolumeMounts(cfg *deployment.DeployServiceOptions) []apiv1.VolumeMount {
	res := make([]apiv1.VolumeMount, len(cfg.Volumes)+len(cfg.Secrets))
	i := 0
	for _, volume := range cfg.Volumes {
		res[i] = apiv1.VolumeMount{
			Name:      fmt.Sprintf("volume-%v", i),
			MountPath: volume.Target,
		}
		i++
	}
	secretKeys := make([]string, len(cfg.Secrets))
	j := 0
	for key := range cfg.Secrets {
		secretKeys[j] = key
		j++
	}
	sort.Strings(secretKeys)
	for _, key := range secretKeys {
		mountPath := cfg.Secrets[key]
		res[i] = apiv1.VolumeMount{
			Name:      fmt.Sprintf("volume-%v", i),
			MountPath: mountPath,
			ReadOnly:  true,
		}
		i++
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func constructVolumes(cfg *deployment.DeployServiceOptions) []apiv1.Volume {
	res := make([]apiv1.Volume, len(cfg.Volumes)+len(cfg.Secrets))
	i := 0
	for _, volume := range cfg.Volumes {
		res[i] = apiv1.Volume{
			Name: fmt.Sprintf("volume-%v", i),
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: volume.Source,
				},
			},
		}
		i++
	}
	secretKeys := make([]string, len(cfg.Secrets))
	j := 0
	for key := range cfg.Secrets {
		secretKeys[j] = key
		j++
	}
	sort.Strings(secretKeys)
	for _, secretName := range secretKeys {
		res[i] = apiv1.Volume{
			Name: fmt.Sprintf("volume-%v", i),
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}
		i++
	}
	return res
}

func constructEnv(cfg *deployment.DeployServiceOptions) []apiv1.EnvVar {
	res := make([]apiv1.EnvVar, len(cfg.Env))
	i := 0
	for k, v := range cfg.Env {
		res[i] = apiv1.EnvVar{
			Name:  k,
			Value: v,
		}
		i++
	}
	return res
}

func buildLabelSelector(labels deployment.LabelSet) string {
	res := ""
	for k, v := range labels {
		res += k + " = " + v + " , "
	}
	if len(res) > 0 {
		return res[:len(res)-3]
	}
	return res
}

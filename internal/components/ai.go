package components

import (
	"context"
	"errors"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/operantai/woodpecker/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operantai/woodpecker/internal/output"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const (
	defaultWoodpeckerAIImage = "ghcr.io/operantai/woodpecker/woodpecker-ai:latest"
)

type AI struct{}

func (ai *AI) Type() string {
	return "woodpecker-ai-verifier"
}

func (ai *AI) Description() string {
	return "Enables running Experiments against AI providers"
}

func (ai *AI) Install(ctx context.Context, config *Config) error {
	if config.Image == "" {
		config.Image = defaultWoodpeckerAIImage
	}

	switch config.Namespace {
	case "local":
		client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
		defer client.Close()
		images, err := client.ImageList(ctx, image.ListOptions{})
		if err != nil {
			return err
		}

		imageExists := false
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == config.Image {
					imageExists = true
					output.WriteInfo("Using local image: %s", config.Image)
					break
				}
			}
		}

		if !imageExists {
			_, err = client.ImagePull(ctx, config.Image, image.PullOptions{})
			if err != nil {
				return err
			}
		}

		resp, err := client.ContainerCreate(ctx, &container.Config{
			Image: config.Image,
			Tty:   false,
			ExposedPorts: nat.PortSet{
				"8000/tcp": struct{}{},
			},
		}, &container.HostConfig{
			PortBindings: nat.PortMap{
				"8000/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "8000",
					},
				},
			},
		}, nil, nil, config.Type)
		if err != nil {
			return err
		}
		if err := client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return err
		}
	default:
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: config.Type,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": config.Type,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": config.Type,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            config.Type,
								Image:           config.Image,
								ImagePullPolicy: corev1.PullIfNotPresent,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 8080,
									},
								},
							},
						},
					},
				},
			},
		}
		_, err = client.Clientset.AppsV1().Deployments(config.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: config.Type,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": config.Type,
				},
				Ports: []corev1.ServicePort{
					{
						Port:       8080,
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		}
		_, err = client.Clientset.CoreV1().Services(config.Namespace).Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ai *AI) Uninstall(ctx context.Context, config *Config) error {
	switch config.Namespace {
	case "local":
		client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
		defer client.Close()

		containers, err := client.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			return err
		}

		var woodpeckerID string
		for _, container := range containers {
			if slices.Contains(container.Names, "/"+config.Type) {
				woodpeckerID = container.ID
				break
			}
		}

		if woodpeckerID == "" {
			return errors.New("Could not find woodpecker-ai container to remove")
		}

		return client.ContainerRemove(ctx, woodpeckerID, container.RemoveOptions{Force: true})
	default:
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}
		err = client.Clientset.AppsV1().Deployments(config.Namespace).Delete(ctx, config.Type, metav1.DeleteOptions{})
		if err != nil {
			return err
		}

		err = client.Clientset.CoreV1().Services(config.Namespace).Delete(ctx, config.Type, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

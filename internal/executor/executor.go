package executor

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/operantai/secops-chaos/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/utils/pointer"
)

const (
	addr string = "127.0.0.1"
)

type RemoteExecutorConfig struct {
	Name       string
	Namespace  string
	Image      string
	Parameters RemoteExecutor
	Addr       string
	StopCh     chan struct{}
	ReadyCh    chan struct{}
	ErrCh      chan error
	Out        *bytes.Buffer
	ErrOut     *bytes.Buffer
	LocalPort  int32
}

type RemoteExecutor struct {
	ServiceAccountName string
	TargetPort         int32
	ImageParameters    []string
}
type RemoteExecuteAPI struct {
	Image              string   `yaml:"image"`
	ImageParameters    []string `yaml:"image_parameters"`
	ServiceAccountName string   `yaml:"service_account_name"`
	Target             Target   `yaml:"target"`
}
type Target struct {
	Port int32  `yaml:"target_port"`
	Path string `yaml:"path"`
}

// Executor configurations are meant to be used to execute remote commands on a pod in a cluster.
func NewExecutorConfig(name, namespace, image string, imageParameters []string, serviceAccountName string, targetPort int32) *RemoteExecutorConfig {
	return &RemoteExecutorConfig{
		Name:      name,
		Namespace: namespace,
		Image:     image,
		Parameters: RemoteExecutor{
			ServiceAccountName: serviceAccountName,
			TargetPort:         targetPort,
			ImageParameters:    imageParameters,
		},
		StopCh:  make(chan struct{}, 1),
		ReadyCh: make(chan struct{}, 1),
		ErrCh:   make(chan error, 1),
		Out:     new(bytes.Buffer),
		ErrOut:  new(bytes.Buffer),
		Addr:    addr,
	}
}

func (r *RemoteExecutorConfig) Deploy(ctx context.Context, client *kubernetes.Clientset) error {
	envVar := prepareImageParameters(r.Parameters.ImageParameters)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": r.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": r.Name,
						"app":        r.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            r.Name,
							Image:           r.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.Parameters.TargetPort,
								},
							},
							Env: envVar,
						},
					},
				},
			},
		},
	}
	params := r.Parameters
	if params.ServiceAccountName != "" {
		deployment.Spec.Template.Spec.ServiceAccountName = params.ServiceAccountName
	}

	_, err := client.AppsV1().Deployments(r.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": r.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port: r.Parameters.TargetPort,
				},
			},
		},
	}

	_, err = client.CoreV1().Services(r.Namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func (r *RemoteExecutorConfig) OpenLocalPort(ctx context.Context, client *k8s.Client) (*portforward.PortForwarder, error) {
	clientset := client.Clientset
	// Deployments can not be port forwarded to directly, this is similar to how kubectl does it
	pods, err := clientset.CoreV1().Pods(r.Namespace).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", r.Name)})
	if err != nil {
		return nil, err
	}

	// Currently only supports one replica in deployment
	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("Deployment failed to deploy pods")
	}

	// Build the port forwarder from restconfig
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", r.Namespace, pods.Items[0].Name)
	hostIP := strings.TrimPrefix(strings.TrimPrefix(client.RestConfig.Host, "http://"), "https://")
	url := url.URL{Scheme: "https", Path: path, Host: hostIP}
	transport, upgrader, err := spdy.RoundTripperFor(client.RestConfig)
	if err != nil {
		return nil, err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url)
	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", r.Parameters.TargetPort)}, r.StopCh, r.ReadyCh, r.Out, r.ErrOut)
	if err != nil {
		return nil, err
	}

	return forwarder, nil
}

func (r *RemoteExecutorConfig) Cleanup(ctx context.Context, client *kubernetes.Clientset) error {
	err := client.AppsV1().Deployments(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return client.CoreV1().Services(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
}

func prepareImageParameters(imageParameters []string) []corev1.EnvVar {
	var envVar []corev1.EnvVar
	for _, param := range imageParameters {
		parts := strings.Split(param, "=")
		envVar = append(envVar, corev1.EnvVar{
			Name:  parts[0],
			Value: parts[1],
		})
	}

	return envVar
}

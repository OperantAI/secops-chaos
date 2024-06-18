/*
Copyright 2023 Operant AI
*/
package k8s

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type portForwarder struct {
	ctx       context.Context
	addr      string
	k8s       *Client
	stopCh    chan struct{}
	readyCh   chan struct{}
	errCh     chan error
	out       *bytes.Buffer
	errOut    *bytes.Buffer
	localPort int32
}

func (c *Client) NewPortForwarder(ctx context.Context) *portForwarder {
	return &portForwarder{
		ctx:     ctx,
		k8s:     c,
		stopCh:  make(chan struct{}, 1),
		readyCh: make(chan struct{}, 1),
		errCh:   make(chan error, 1),
		out:     new(bytes.Buffer),
		errOut:  new(bytes.Buffer),
	}
}

func (pf *portForwarder) Addr() string {
	return "127.0.0.1"
}

// Forward fowards a local port to a given namespace, label selector, and port
func (pf *portForwarder) Forward(namespace, selector string, port int) (*portforward.ForwardedPort, error) {
	clientset := pf.k8s.Clientset
	pods, err := clientset.CoreV1().Pods(namespace).List(pf.ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("No pods found")
	}

	// Build the port forwarder from restconfig
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, pods.Items[0].Name)
	hostIP := strings.TrimPrefix(strings.TrimPrefix(pf.k8s.RestConfig.Host, "http://"), "https://")
	url := url.URL{Scheme: "https", Path: path, Host: hostIP}
	transport, upgrader, err := spdy.RoundTripperFor(pf.k8s.RestConfig)
	if err != nil {
		return nil, err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url)
	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", port)}, pf.stopCh, pf.readyCh, pf.out, pf.errOut)
	if err != nil {
		return nil, err
	}

	go func() {
		err = forwarder.ForwardPorts()
		if err != nil {
			pf.errCh <- fmt.Errorf("Failed to open port: %w", err)
		}
	}()

	select {
	case <-pf.readyCh:
		break
	case err := <-pf.errCh:
		return nil, err
	}

	ports, err := forwarder.GetPorts()
	if err != nil {
		return nil, fmt.Errorf("Failed to get forwarder ports: %w", err)
	}
	return &ports[0], nil
}

func (pf *portForwarder) Stop() {
	close(pf.stopCh)
}

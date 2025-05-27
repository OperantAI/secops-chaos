package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CheckNamespaceExists(ctx context.Context, namespace string) error {
	_, err := c.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Could not find namespace: %w", err)
	}
	return nil
}

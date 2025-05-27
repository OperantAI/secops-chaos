package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CheckForSecret(ctx context.Context, namespace, secretName string) error {
	_, err := c.Clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Could not find secret: %w", err)
	}
	return nil
}

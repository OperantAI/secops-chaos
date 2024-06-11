/*
Copyright 2023 Operant AI
*/
package k8s

import (
	k8sVersion "k8s.io/apimachinery/pkg/version"
)

func (c *Client) GetK8sVersion() (*k8sVersion.Info, error) {
	version, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	return version, nil
}

/*
Copyright 2023 Operant AI
*/
package k8s

import (
	k8sVersion "k8s.io/apimachinery/pkg/version"

	"k8s.io/client-go/kubernetes"
)

func GetK8sVersion(client *kubernetes.Clientset) (*k8sVersion.Info, error) {
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	return version, nil
}

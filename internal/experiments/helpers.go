package experiments

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/operantai/woodpecker/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const tmpFileDir = "/tmp/woodpecker"

func createTempFile(experimentType, experiment string) (*os.File, error) {
	if _, err := os.Stat(tmpFileDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(tmpFileDir, 0700); err != nil {
				return nil, err
			}
		}
	}
	file, err := os.CreateTemp(tmpFileDir, fmt.Sprintf("%s-%s", experimentType, experiment))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getTempFileContentsForExperiment(experimentType, experiment string) ([][]byte, error) {
	var contents [][]byte
	files, err := getTempFilesForExperiment(experimentType, experiment)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	return contents, nil
}

func getTempFilesForExperiment(experimentType, experiment string) ([]string, error) {
	d, err := os.Open(tmpFileDir)
	if err != nil {
		return nil, err
	}
	files, err := d.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	var fullPaths []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), fmt.Sprintf("%s-%s", experimentType, experiment)) {
			fullPaths = append(fullPaths, filepath.Join(tmpFileDir, file.Name()))
		}
	}
	return fullPaths, nil
}

func removeTempFilesForExperiment(experimentType, experiment string) error {
	files, err := getTempFilesForExperiment(experimentType, experiment)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}

	}
	return nil
}

const WoodpeckerAI = "woodpecker-ai"

func isWoodpeckerAIDockerComponentPresent(ctx context.Context, client *dockerClient.Client) bool {
	containers, err := client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return false
	}
	for _, container := range containers {
		if slices.Contains(container.Names, WoodpeckerAI) {
			return true
		}
	}
	return false
}

func isWoodpeckerAIK8sComponentPresent(ctx context.Context, client *k8s.Client, namespace string) bool {
	_, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, WoodpeckerAI, metav1.GetOptions{})
	return err == nil
}

package experiments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const tmpFileDir = "/tmp/secops-chaos"

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

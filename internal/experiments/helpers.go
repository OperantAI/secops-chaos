package experiments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const tmpFileDir = "/tmp/secops-chaos"

func createTempFile(experimentType, experiment string) (*os.File, error) {
	file, err := os.CreateTemp(tmpFileDir, fmt.Sprintf("%s:%s", experimentType, experiment))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getTempFileContentsForExperiment(experimentType, experiment string) ([][]byte, error) {
	var contents [][]byte
	d, err := os.Open(tmpFileDir)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	files, err := d.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), fmt.Sprintf("%s:%s", experimentType, experiment)) {
			fp := filepath.Join(tmpFileDir, file.Name())
			content, err := os.ReadFile(fp)
			if err != nil {
				return nil, err
			}
			contents = append(contents, content)
		}
	}
	return contents, nil
}

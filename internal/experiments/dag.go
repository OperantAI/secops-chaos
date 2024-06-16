package experiments

import (
	"log"
	"sync"
	"time"

	"github.com/heimdalr/dag"
	"github.com/operantai/secops-chaos/internal/output"
)

type experimentVisitor struct {
	*Runner
	*sync.WaitGroup
}

func (v *experimentVisitor) Visit(vertex dag.Vertexer) {
	_, value := vertex.Vertex()
	experimentName := value.(string)
	experiment, ok := v.Experiments[experimentName]
	if !ok {
		output.WriteError("Experiment not found in DAG: %s", experimentName)
		return
	}
	dependencies := experiment.DependsOn()
	for _, depName := range dependencies {
		if !v.waitForDependency(depName) {
			output.WriteError("Timeout waiting for dependency %s of experiment %s to finish", depName, experimentName)
			return
		}
	}

	v.runExperiment(experimentName)
}

func (v *experimentVisitor) waitForDependency(depName string) bool {
	v.Runner.Mutex.Lock()
	defer v.Runner.Mutex.Unlock()

	for !v.Executed[depName] {
		v.Runner.Mutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		v.Runner.Mutex.Lock()

		select {
		case <-v.Runner.ctx.Done():
			return false
		default:
		}
	}

	return true
}

func (v *experimentVisitor) runExperiment(id string) {
	v.Add(1)
	go func() {
		output.WriteInfo("Running experiment: %s", id)
		defer v.Done()
		err := v.Runner.Experiments[id].Run(v.Runner.ctx, v.Runner.Client)
		if err != nil {
			log.Fatalf("Error running experiment %s: %v", id, err)
		}
		v.Runner.Mutex.Lock()
		defer v.Runner.Mutex.Unlock()
		v.Executed[id] = true
	}()
}

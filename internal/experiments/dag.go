package experiments

import (
	"github.com/heimdalr/dag"
	"github.com/operantai/secops-chaos/internal/output"
)

type DAGStatus int64

const (
	PendingStatus   DAGStatus = 0
	RunningStatus   DAGStatus = 1
	CompletedStatus DAGStatus = 2
)

type experimentVisitor struct {
	*Runner
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
		if !v.waitForDependency(experiment.Name(), depName) {
			output.WriteError("Timeout waiting for dependency %s of experiment %s to finish", depName, experimentName)
			return
		}
	}
	v.runExperiment(experimentName)
}

func (v *experimentVisitor) waitForDependency(experiment, depName string) bool {
	output.WriteInfo("Experiment %s waiting for dependency to finish: %s", experiment, depName)
	// for {
	// 	depStatus, ok := v.ExperimentStatus.Load(depName)
	// 	if !ok {
	// 		output.WriteError("Experiment %s dependency %s not found", experiment, depName)
	// 	}
	// 	if depStatus == CompletedStatus {
	// 		return true
	// 	}

	// 	select {
	// 	case <-time.After(100 * time.Millisecond):
	// 	case <-v.Runner.ctx.Done():
	// 		return false
	// 	}
	// }
	v.Runner.Cond.L.Lock()
	defer v.Runner.Cond.L.Unlock()
	for {
		if depStatus, ok := v.ExperimentStatus.Load(depName); ok && depStatus == CompletedStatus {
			return true
		}
		v.Runner.Cond.Wait()
	}
}

func (v *experimentVisitor) runExperiment(id string) {
	v.Runner.WaitGroup.Add(1)
	go func() {
		output.WriteInfo("Running experiment: %s", id)
		defer v.Runner.WaitGroup.Done()
		v.ExperimentStatus.Store(id, RunningStatus)
		err := v.Runner.Experiments[id].Run(v.Runner.ctx, v.Runner.Client)
		if err != nil {
			output.WriteError("Error running experiment %s: %s", id, err)
		}
		v.ExperimentStatus.Store(id, CompletedStatus)
		output.WriteInfo("Experiment completed: %s", id)
		v.Runner.Cond.Broadcast()
	}()
}

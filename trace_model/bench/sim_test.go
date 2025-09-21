package bench

import (
	"testing"

	"github.com/dfquaresma/faas-simulator/trace_model/runner"
)

const tracePath = "../azure/sample-100k-processed.csv"
const outputPath = "../results/"
const forwardLatency = 1

func BenchmarkSimulator(b *testing.B) {
	technique := []string{"GCI"}
	percentileThreshould := []string{"p99"}
	replicaIdleTime := []int{-1}
	runner.Sim(tracePath, outputPath, technique, percentileThreshould, replicaIdleTime, forwardLatency)
}

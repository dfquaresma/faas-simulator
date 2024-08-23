package bench

import (
	"testing"

	"github.com/dfquaresma/faas-simulator/runner"
)

const tracePath = "../azure/inv2021-processed.csv"
const outputPath = "../results/"

var forwardLatency = []int{1}

func BenchmarkSimulator(b *testing.B) {
	technique := []string{"RequestHedgingDefault"}
	hasOracle := []string{"true"}
	percentileThreshould := []string{"p99"}
	replicaIdleTime := []int{-1}
	runner.Sim(tracePath, outputPath, technique, hasOracle, percentileThreshould, replicaIdleTime, forwardLatency)
}

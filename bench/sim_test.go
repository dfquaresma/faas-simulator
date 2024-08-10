package bench

import (
	"testing"

	"github.com/dfquaresma/faas-simulator/runner"
)

const tracePath = "../azure/sample.csv"
const outputPath = "../results/"

var forwardLatency = []int{0}

func BenchmarkSimulatorRequestHedgingOptOracleP99Inf(b *testing.B) {
	technique := []string{"RequestHedgingOpt"}
	hasOracle := []string{"true"}
	percentileThreshould := []string{"p99"}
	replicaIdleTime := []int{-1}
	runner.Sim(tracePath, outputPath, technique, hasOracle, percentileThreshould, replicaIdleTime, forwardLatency)
}

func BenchmarkSimulatorRequestHedgingOptNoOracleP99Inf(b *testing.B) {
	technique := []string{"RequestHedgingOpt"}
	hasOracle := []string{"false"}
	percentileThreshould := []string{"p99"}
	replicaIdleTime := []int{-1}
	runner.Sim(tracePath, outputPath, technique, hasOracle, percentileThreshould, replicaIdleTime, forwardLatency)
}

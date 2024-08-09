package main

import (
	"log"

	"github.com/dfquaresma/faas-simulator/runner"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		return
	}

	tracePath := viper.GetString("tracePath")
	outputPath := viper.GetString("outputPath")

	techniques := viper.GetStringSlice("resourceProvisioner.technique")
	hasOracle := viper.GetStringSlice("resourceProvisioner.hasOracle")
	tailLatencyProbs := viper.GetStringSlice("resourceProvisioner.tailLatencyProb")

	idletimes := viper.GetIntSlice("resourceProvisioner.idletime")
	forwardLatencies := viper.GetIntSlice("resourceProvisioner.forwardLatency")

	runner.Sim(tracePath, outputPath, techniques, hasOracle, tailLatencyProbs, idletimes, forwardLatencies)
}

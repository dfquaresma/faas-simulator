package main

import (
	"log"
	"strconv"

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
	tailLatencyProbs := viper.GetStringSlice("resourceProvisioner.tailLatencyProb")

	idletimes := viper.GetIntSlice("resourceProvisioner.idletime")
	forwardLatencies := viper.GetIntSlice("resourceProvisioner.forwardLatency")

	skipColdStartStrSlc := viper.GetStringSlice("resourceProvisioner.skipColdStart")
	skipColdStartOps := make([]bool, len(skipColdStartStrSlc))
	for i, s := range skipColdStartStrSlc {
		val, err := strconv.ParseBool(s)
		if err != nil {
			panic(err)
		}
		skipColdStartOps[i] = val
	}

	runner.Sim(tracePath, outputPath, techniques, tailLatencyProbs, idletimes, forwardLatencies, skipColdStartOps)
}

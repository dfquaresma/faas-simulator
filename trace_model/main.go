package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dfquaresma/faas-simulator/trace_model/runner"
	"github.com/spf13/viper"
)

func main() {
	start := time.Now()
	runner.Sim(getConfigs("generator"))
	generator_total_time := time.Since(start)

	start = time.Now()
	runner.Sim(getConfigs("azure"))
	azure_total_time := time.Since(start)

	fmt.Printf("Generator TotalTime: %s, Azure Totaltime: %s", generator_total_time, azure_total_time)
}

func getConfigs(s string) (string, string, []string, []string, []int, int) {
	viper.SetConfigFile("config.json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		panic("Failed to read config file")
	}

	tracePath := viper.GetString(s + ".tracePath")
	outputPath := viper.GetString(s + ".outputPath")
	techniques := viper.GetStringSlice(s + ".resourceProvisioner.technique")
	tailLatencyProbs := viper.GetStringSlice(s + ".resourceProvisioner.tailLatencyProb")

	idletimes := viper.GetIntSlice(s + ".resourceProvisioner.idletime")
	forwardLatencies := viper.GetInt(s + ".forwardLatency")

	return tracePath, outputPath, techniques, tailLatencyProbs, idletimes, forwardLatencies
}

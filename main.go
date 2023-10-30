package main

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/dfquaresma/faas-simulator/common"
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
	rows := readInput(tracePath)

	invocations, err := common.NewInvocations(rows)
	if err != nil {
		panic(err)
	}

	cfg := common.Config{
		ColdstartLatency: viper.GetFloat64("resourceProvisioner.coldstartLatency"),
		ForwardLatency:   viper.GetFloat64("resourceProvisioner.forwardLatency"),
		Idletime:         viper.GetFloat64("resourceProvisioner.idletime"),
		TailLatency:      viper.GetFloat64("replica.tailLatency"),
		TailLatencyProb:  viper.GetFloat64("replica.tailLatencyProb"),
		Technique:        viper.GetString("technique"),
	}
	selector := common.NewSelector(cfg)
	replayer := common.NewReplayer(invocations, selector)
	replayer.Run()

	outputPath := viper.GetString("outputPath")
	writeOutput(outputPath+"-invocations.csv", invocations.GetOutPut())
	writeOutput(outputPath+"-replicas.csv", selector.GetOutPut())
}

func readInput(tracePath string) [][]string {
	input, err := os.Open(tracePath)
	if err != nil {
		panic(err)
	}
	defer input.Close()

	r := csv.NewReader(input)
	rows, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	return rows[1:]
}

func writeOutput(outputPath string, data [][]string) {
	output, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	for _, record := range data {
		err := writer.Write(record)
		if err != nil {
			panic(err)
		}
	}
}

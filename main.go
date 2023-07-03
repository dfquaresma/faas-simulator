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
	outputPath := viper.GetString("outputPath")
	tailLatencyProb := viper.GetString("simulationSettings.tailLatencyProb")
	Coldstart := viper.GetString("simulationSettings.Coldstart")
	TailLatency := viper.GetString("simulationSettings.TailLatency")
	Idletime := viper.GetString("simulationSettings.Idletime")
	log.Printf("Config: %s, %s, %s, %s, %s, %s", tracePath, outputPath, tailLatencyProb, Coldstart, TailLatency, Idletime)

	rows := readInput(tracePath)

	invocations, err := common.NewInvocations(rows)
	if err != nil {
		panic(err)
	}
	selector := common.NewSelector()
	replayer := common.NewReplayer(invocations, selector)
	replayer.Run()

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
	return rows
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

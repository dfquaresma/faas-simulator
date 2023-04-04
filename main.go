package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/dfquaresma/faas-simulator/common"
)

var (
	tracePath  = flag.String("trace", "", "Comma-separated trace file to reproduce in simulation")
	outputPath = flag.String("output", "", "file path to output results")
)

func main() {
	flag.Parse()

	validateParams(*tracePath, *outputPath)

	rows := readInput(*tracePath)

	invocations := newInvocations(rows)
	selector := newSelector()
	replayer := newReplayer(invocations, selector)
	replayer.Run()

	writeOutput(*outputPath + "invocations.csv", invocations.getOutPut())
	writeOutput(*outputPath + "replicas.csv", selector.getOutPut())

}

func validateParams(tracePath, outputPath string) {
	if tracePath == "" {
		log.Fatalf("A trace path must be given!")
	}
	if outputPath == "" {
		log.Fatalf("An output path must be given!")
	}
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

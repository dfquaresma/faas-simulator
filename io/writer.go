package io

import (
	"encoding/csv"
	"os"
)

func WriteOutput(outputPath, simulationName string, data [][]string) {
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	filePath := outputPath + simulationName
	output, err := os.Create(filePath)
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

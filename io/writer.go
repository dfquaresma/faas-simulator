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

func WriteOutputHeaderRow(outputPath, simulationName string, rows []string) {
	_ = os.Remove(outputPath + simulationName)
	WriteOutputByRow(outputPath, simulationName, rows)
}

func WriteOutputByRow(outputPath, simulationName string, rows []string) {
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	filePath := outputPath + simulationName
	// Open the file in append mode, create if not exists
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(rows)
	if err != nil {
		panic(err)
	}
}

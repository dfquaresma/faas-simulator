package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/dfquaresma/faas-simulator/latency_model/distuv"
	"github.com/dfquaresma/faas-simulator/trace_model/io"
)

func main() {
	firstStart := time.Now()
	viper.SetConfigFile("config.json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		return
	}

	requests_count := viper.GetInt("requests_count")
	outputPath := viper.GetString("outputPath")
	functions := viper.GetStringSlice("functions")

	sim_results := [][]string{{"app", "func", "end_timestamp", "duration"}}
	viper.SetConfigFile("../synthetic_functions.json")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		return
	}
	for _, f := range functions {
		start := time.Now()
		fmt.Printf("Running for function %s...", f)
		interarrival_distname := viper.GetString(f + ".interarrival_distribution")
		servicetime_distname := viper.GetString(f + ".servicetime_distribution")
		sim_results = append(
			sim_results,
			generate(requests_count, f, interarrival_distname, servicetime_distname)...,
		)
		fmt.Printf(" Finished. Time Running: %s\n", time.Since(start))
	}
	io.WriteOutput(outputPath, "generated-trace.csv", sim_results)
	fmt.Printf("\nTotal Time of Simulation: %s\n", time.Since(firstStart))
}

func generate(requests_count int, f, interarrival_distname, servicetime_distname string) [][]string {
	interarrival_dist := distuv.NewDistribution(f, interarrival_distname)
	servicetime_dist := distuv.NewDistribution(f, servicetime_distname)
	if interarrival_dist == nil || servicetime_dist == nil {
		panic(fmt.Sprintf("Either %s or %s for %s is not valid", interarrival_distname, servicetime_distname, f))
	}

	ts := 0.0
	generated_app := servicetime_distname + "_" + interarrival_distname + "-app"
	generated_func := f
	workload := [][]string{}
	for i := 0; i < requests_count; i++ {
		ts = ts + interarrival_dist.NextValue()
		duration := servicetime_dist.NextValue()
		end_timestamp := ts + duration
		workload = append(workload, []string{
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(duration, 'f', -1, 64),
		})
	}
	return workload
}

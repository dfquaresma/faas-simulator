package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/spf13/viper"

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

	sim_results := [][]string{{
		"technique", "app", "func",
		"end_timestamp", "response_time", "total_time_running_functions",
		"service_time", "copy_service_time", "delay"},
	}
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
	io.WriteOutput(outputPath, "simple_model-full-results.csv", sim_results)
	fmt.Printf("\nTotal Time of Simulation: %s\n", time.Since(firstStart))
}

func generate(requests_count int, f, interarrival_distname, servicetime_distname string) [][]string {
	interarrival_dist := NewDistribution(f, interarrival_distname)
	servicetime_dist := NewDistribution(f, servicetime_distname)
	if interarrival_dist == nil || servicetime_dist == nil {
		panic(fmt.Sprintf("Either %s or %s for %s is not valid", interarrival_distname, servicetime_distname, f))
	}

	ts := 0.0
	generated_app := servicetime_distname + "_" + interarrival_distname + "-app"
	generated_func := f
	workload := [][]string{}
	for i := 0; i < requests_count; i++ {
		ts = ts + interarrival_dist.NextValue()
		service_time := servicetime_dist.NextValue()
		copy_service_time := servicetime_dist.NextValue()

		workload = append(workload, getBaseline(service_time, ts, generated_app, generated_func))
		workload = append(workload, getHedgedRequestNoDelay(service_time, copy_service_time, ts, generated_app, generated_func))
		workload = append(workload, getHedgedRequestDelayP95(servicetime_dist, service_time, copy_service_time, ts, generated_app, generated_func))
		workload = append(workload, getHedgedRequestDelayP99(servicetime_dist, service_time, copy_service_time, ts, generated_app, generated_func))
		workload = append(workload, getPerfectHedgedRequest(service_time, copy_service_time, ts, generated_app, generated_func))
		workload = append(workload, getNearPerfectHedgedRequest(service_time, copy_service_time, ts, generated_app, generated_func))

		//workload = append(workload, getNaiveHedgedNoDelay(service_time, copy_service_time, ts, generated_app, generated_func))
		//workload = append(workload, getDelayedNaiveHedged(servicetime_dist, service_time, copy_service_time, ts, generated_app, generated_func))
	}
	return workload
}

func getBaseline(service_time, ts float64, generated_app, generated_func string) []string {
	response_time := service_time
	end_timestamp := ts + response_time
	total_time_running_functions := response_time // vanilla scenario, send just one and that's all
	return []string{
		"baseline",
		generated_app,
		generated_func,
		strconv.FormatFloat(end_timestamp, 'f', -1, 64),
		strconv.FormatFloat(response_time, 'f', -1, 64),
		strconv.FormatFloat(total_time_running_functions, 'f', -1, 64),
		strconv.FormatFloat(service_time, 'f', -1, 64),
		"0",
		"0",
	}
}

func getHedgedRequestNoDelay(service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	return getHedgedRequest(service_time, copy_service_time, 0.0, ts, "hedged_requests_nodelay", generated_app, generated_func)
}

func getHedgedRequestDelayP95(servicetime_dist *Distribution, service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	p95 := servicetime_dist.GetPercentile(0.95)
	return getHedgedRequest(service_time, copy_service_time, p95, ts, "hedged_requests_p95", generated_app, generated_func)
}

func getHedgedRequestDelayP99(servicetime_dist *Distribution, service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	p99 := servicetime_dist.GetPercentile(0.99)
	return getHedgedRequest(service_time, copy_service_time, p99, ts, "hedged_requests_p99", generated_app, generated_func)
}

func getNearPerfectHedgedRequest(service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	// hedge with cancellation, but only consider copy if it is worth
	delay := service_time + 1.0
	if copy_service_time < service_time {
		delay = 0.0 // if copy is faster, send it right away
	}
	return getHedgedRequest(service_time, copy_service_time, delay, ts, "near_perfect_hedged_requests", generated_app, generated_func)
}

func getPerfectHedgedRequest(service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	// if copy_service_time is faster, we switch it with original to cancel the original service_time and run only the copy one.
	shouldSwap := copy_service_time < service_time
	if shouldSwap {
		service_time, copy_service_time = copy_service_time, service_time
	}
	// always set delay above function to be sent in order to avoid sending a copy
	delay := service_time + 1.0

	res := getHedgedRequest(service_time, copy_service_time, delay, ts, "perfect_hedged_requests", generated_app, generated_func)
	if shouldSwap {
		// swap values in response to proper output sim stats
		res[6], res[7] = res[7], res[6]
	}
	return res
}

func getHedgedRequest(service_time, copy_service_time, delay, ts float64, name, generated_app, generated_func string) []string {
	// hedge with cancellation
	response_time := math.Min(service_time, delay+copy_service_time)
	end_timestamp := ts + response_time
	total_time_running_functions := response_time
	if response_time > delay {
		delta := response_time - delay
		total_time_running_functions = delay + 2*delta // add additinal time spent running function after delay up to first finish
	}
	return []string{
		name,
		generated_app,
		generated_func,
		strconv.FormatFloat(end_timestamp, 'f', -1, 64),
		strconv.FormatFloat(response_time, 'f', -1, 64),
		strconv.FormatFloat(total_time_running_functions, 'f', -1, 64),
		strconv.FormatFloat(service_time, 'f', -1, 64),
		strconv.FormatFloat(copy_service_time, 'f', -1, 64),
		strconv.FormatFloat(delay, 'f', -1, 64),
	}
}

func getNaiveHedgedNoDelay(service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	return getNaiveHedged(service_time, copy_service_time, 0, ts, "naive_hedge", generated_app, generated_func)
}

func getDelayedNaiveHedged(servicetime_dist *Distribution, service_time, copy_service_time, ts float64, generated_app, generated_func string) []string {
	return getNaiveHedged(service_time, copy_service_time, servicetime_dist.GetPercentile(0.95), ts, "delayed_naive_hedge", generated_app, generated_func)
}

func getNaiveHedged(service_time, copy_service_time, delay, ts float64, name, generated_app, generated_func string) []string {
	// hedge with no cancellation
	response_time := math.Min(service_time, delay+copy_service_time)
	end_timestamp := ts + response_time
	total_time_running_functions := response_time
	if response_time > delay {
		total_time_running_functions = service_time + copy_service_time // if a second is sent, process both completely
	}
	return []string{
		name,
		generated_app,
		generated_func,
		strconv.FormatFloat(end_timestamp, 'f', -1, 64),
		strconv.FormatFloat(response_time, 'f', -1, 64),
		strconv.FormatFloat(total_time_running_functions, 'f', -1, 64),
		strconv.FormatFloat(service_time, 'f', -1, 64),
		strconv.FormatFloat(copy_service_time, 'f', -1, 64),
		strconv.FormatFloat(delay, 'f', -1, 64),
	}
}

package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"

	"github.com/dfquaresma/faas-simulator/io"
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
	interarrival_dist := newDistribution(f, interarrival_distname)
	servicetime_dist := newDistribution(f, servicetime_distname)
	if interarrival_dist == nil || servicetime_dist == nil {
		panic(fmt.Sprintf("Either %s or %s for %s is not valid", interarrival_distname, servicetime_distname, f))
	}

	ts := 0.0
	generated_app := servicetime_distname + "_" + interarrival_distname + "-app"
	generated_func := f
	workload := [][]string{}
	for i := 0; i < requests_count; i++ {
		ts = ts + interarrival_dist.nextValue()
		duration := servicetime_dist.nextValue()
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

type distribution struct {
	dist         string
	latency      float64
	tail_latency float64
	b            distuv.Bernoulli
	ln           distuv.LogNormal
	ps           distuv.Poisson
	wb           distuv.Weibull
}

func newDistribution(f, dist string) *distribution {
	switch dist {
	case "constant":
		return &distribution{
			dist:    dist,
			latency: viper.GetFloat64(f + ".distributions.constant.latency"),
		}
	case "bernoulli":
		return &distribution{
			dist:         dist,
			latency:      viper.GetFloat64(f + ".distributions.bernoulli.latency"),
			tail_latency: viper.GetFloat64(f + ".distributions.bernoulli.tailLatency"),
			b: distuv.Bernoulli{
				P:   viper.GetFloat64(f + ".distributions.bernoulli.prob"),
				Src: rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "poisson":
		return &distribution{
			dist: dist,
			ps: distuv.Poisson{
				Lambda: viper.GetFloat64(f + ".distributions.poisson.lambda"),
				Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "weibull":
		return &distribution{
			dist: dist,
			wb: distuv.Weibull{
				K:      viper.GetFloat64(f + ".distributions.weibull.k"),
				Lambda: viper.GetFloat64(f + ".distributions.weibull.lambda"),
				Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "lognormal":
		return &distribution{
			dist: dist,
			ln: distuv.LogNormal{
				Mu:    viper.GetFloat64(f + ".distributions.logNormal.mu"),
				Sigma: viper.GetFloat64(f + ".distributions.logNormal.sigma"),
				Src:   rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	default:
		return nil
	}
}

func (d *distribution) nextValue() float64 {
	switch d.dist {
	case "bernoulli":
		if d.b.Rand() == 0 {
			return d.latency
		} else {
			return d.tail_latency
		}
	case "weibull":
		return d.wb.Rand()
	case "lognormal":
		return d.ln.Rand()
	default:
		return d.latency
	}
}

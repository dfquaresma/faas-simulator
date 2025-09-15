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
	viper.SetConfigFile("generator_config.json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		return
	}

	requests_count := viper.GetInt("requests_count")
	interarrival_dist := viper.GetStringSlice("interarrival_distribution")
	latency_dist := viper.GetStringSlice("latency_distribution")
	outputPath := viper.GetString("outputPath")

	for i, dist := range interarrival_dist {
		generate(requests_count, dist, latency_dist[i], outputPath)
	}
}

func generate(requests_count int, interarrival_dist, latency_dist, outputPath string) {
	interarrival := newDistribution(interarrival_dist)
	latency := newDistribution(latency_dist)
	if interarrival == nil || latency == nil {
		panic(fmt.Sprintf("Either %s or %s is not valid", interarrival_dist, latency_dist))
	}

	ts := 0.0
	generated_app := interarrival_dist + "-" + latency_dist + "-app"
	generated_func := interarrival_dist + "-" + latency_dist + "-func"
	workload := [][]string{{"app", "func", "end_timestamp", "duration"}}
	for i := 0; i < requests_count; i++ {
		ts = ts + interarrival.nextValue()
		duration := latency.nextValue()
		end_timestamp := ts + duration
		workload = append(workload, []string{
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(duration, 'f', -1, 64),
		})
	}

	outputName := "latency_" + latency_dist + "-arrival_" + interarrival_dist + ".csv"
	io.WriteOutput(outputPath, outputName, workload)
}

type distribution struct {
	dist         string
	latency      float64
	tail_latency float64
	prob         float64
	ln           distuv.LogNormal
	ps           distuv.Poisson
	wb           distuv.Weibull
	rng          *rand.Rand
}

func newDistribution(dist string) *distribution {
	switch dist {
	case "constant":
		return &distribution{
			dist:    dist,
			latency: viper.GetFloat64("distributions.constant.latency"),
		}
	case "bernoulli":
		return &distribution{
			dist:         dist,
			latency:      viper.GetFloat64("distributions.bernoulli.latency"),
			tail_latency: viper.GetFloat64("distributions.bernoulli.tailLatency"),
			prob:         viper.GetFloat64("distributions.bernoulli.prob"),
			rng:          rand.New(rand.NewSource(uint64(time.Now().Nanosecond()))),
		}
	case "poisson":
		return &distribution{
			dist: dist,
			ps: distuv.Poisson{
				Lambda: viper.GetFloat64("distributions.poisson.lambda"),
				Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "weibull":
		return &distribution{
			dist: dist,
			wb: distuv.Weibull{
				K:      viper.GetFloat64("distributions.weibull.k"),
				Lambda: viper.GetFloat64("distributions.weibull.lambda"),
				Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "logNormal":
		return &distribution{
			dist: dist,
			ln: distuv.LogNormal{
				Mu:    viper.GetFloat64("distributions.logNormal.mu"),
				Sigma: viper.GetFloat64("distributions.logNormal.sigma"),
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
		latency := d.latency
		if d.rng.Float64() >= 1-d.prob {
			latency = d.tail_latency
		}
		return latency
	case "poisson":
		return d.ps.Rand()
	case "weibull":
		return d.wb.Rand()
	case "logNormal":
		return d.ln.Rand()
	default:
		return d.latency
	}
}

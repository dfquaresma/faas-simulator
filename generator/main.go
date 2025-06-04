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

	interarrival_dist := viper.GetString("interarrival_distribution")
	latency_dist := viper.GetString("latency_distribution")
	interarrival := newDistribution(interarrival_dist)
	latency := newDistribution(latency_dist)
	if interarrival == nil || latency == nil {
		panic(fmt.Sprintf("Either %s or %s is not valid", interarrival_dist, latency_dist))
	}

	ts := 0.0
	generated_app := viper.GetString("app_id")
	generated_func := viper.GetString("func_id")
	requests_count := viper.GetInt("requests_count")
	workload := [][]string{[]string{"app", "func", "end_timestamp", "duration"}}
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

	outputPath := viper.GetString("outputPath")
	outputName := viper.GetString("outputName")
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
	case "constant_with_tail":
		return &distribution{
			dist:         dist,
			latency:      viper.GetFloat64("distributions.constant_with_tail.latency"),
			tail_latency: viper.GetFloat64("distributions.constant_with_tail.tailLatency"),
			prob:         viper.GetFloat64("distributions.constant_with_tail.prob"),
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
	case "constant_with_tail":
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

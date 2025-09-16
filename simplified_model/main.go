package main

import (
	"fmt"
	"log"
	"math"
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
	workload := [][]string{{"technique", "app", "func", "end_timestamp", "duration", "total_duration"}}
	for i := 0; i < requests_count; i++ {
		ts = ts + interarrival.nextValue()

		duration := latency.nextValue()
		minDuration := duration
		end_timestamp := ts + minDuration
		total := minDuration // vanilla scenario, send just one and that's all
		workload = callAppend(
			workload,
			"baseline",
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(minDuration, 'f', -1, 64),
			strconv.FormatFloat(total, 'f', -1, 64),
		)

		copy_duration := latency.nextValue()
		minDuration = math.Min(duration, copy_duration)
		end_timestamp = ts + minDuration
		total = duration + copy_duration // send two and process both completely
		workload = callAppend(
			workload,
			"naive_hedge",
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(minDuration, 'f', -1, 64),
			strconv.FormatFloat(total, 'f', -1, 64),
		)

		total = 2 * minDuration // send two but cancels the longer after the shorter finishes
		workload = callAppend(
			workload,
			"hedged_nodelay_cancellation",
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(minDuration, 'f', -1, 64),
			strconv.FormatFloat(total, 'f', -1, 64),
		)

		delay := latency.getP95()
		minDuration = math.Min(duration, delay+copy_duration)
		end_timestamp = ts + minDuration
		total = minDuration
		if minDuration >= delay {
			total = duration + copy_duration // two invocs sent, count them both completely
		}
		workload = callAppend(
			workload,
			"hedge_delayed_nocancellation",
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(minDuration, 'f', -1, 64),
			strconv.FormatFloat(total, 'f', -1, 64),
		)

		total = minDuration
		if minDuration > delay {
			total += minDuration - delay // duplicates only after delay until the first finish
		}
		workload = callAppend(
			workload,
			"hedged_requests",
			generated_app,
			generated_func,
			strconv.FormatFloat(end_timestamp, 'f', -1, 64),
			strconv.FormatFloat(minDuration, 'f', -1, 64),
			strconv.FormatFloat(total, 'f', -1, 64),
		)
	}

	outputName := "latency_" + latency_dist + "-arrival_" + interarrival_dist + ".csv"
	io.WriteOutput(outputPath, outputName, workload)
}

func callAppend(list_to_append [][]string, technique, appId, funcId, ts, dur, totaldur string) [][]string {
	return append(list_to_append, []string{technique, appId, funcId, ts, dur, totaldur})
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
			b: distuv.Bernoulli{
				P:   viper.GetFloat64("distributions.bernoulli.prob"),
				Src: rand.NewSource(uint64(time.Now().Nanosecond())),
			},
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
		if d.b.Rand() == 0 {
			return d.latency
		} else {
			return d.tail_latency
		}
	case "weibull":
		return d.wb.Rand()
	case "logNormal":
		return d.ln.Rand()
	default:
		return d.latency
	}
}

func (d *distribution) getP95() float64 {
	switch d.dist {
	case "bernoulli":
		if d.b.Quantile(0.95) == 0 {
			return d.latency
		} else {
			return d.tail_latency
		}
	case "weibull":
		return d.wb.Quantile(0.95)
	case "logNormal":
		return d.ln.Quantile(0.95)
	default:
		return d.latency
	}
}

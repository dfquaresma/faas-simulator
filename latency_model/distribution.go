package main

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type Distribution struct {
	dist         string
	latency      float64
	tail_latency float64
	b            distuv.Bernoulli
	ln           distuv.LogNormal
	ps           distuv.Poisson
	wb           distuv.Weibull
}

func NewDistribution(f, dist string) *Distribution {
	switch dist {
	case "constant":
		return &Distribution{
			dist:    dist,
			latency: viper.GetFloat64(f + ".distributions.constant.latency"),
		}
	case "bernoulli":
		return &Distribution{
			dist:         dist,
			latency:      viper.GetFloat64(f + ".distributions.bernoulli.latency"),
			tail_latency: viper.GetFloat64(f + ".distributions.bernoulli.tailLatency"),
			b: distuv.Bernoulli{
				P:   viper.GetFloat64(f + ".distributions.bernoulli.prob"),
				Src: rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "poisson":
		return &Distribution{
			dist: dist,
			ps: distuv.Poisson{
				Lambda: viper.GetFloat64(f + ".distributions.poisson.lambda"),
				Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "weibull":
		return &Distribution{
			dist: dist,
			wb: distuv.Weibull{
				K:      viper.GetFloat64(f + ".distributions.weibull.k"),
				Lambda: viper.GetFloat64(f + ".distributions.weibull.lambda"),
				Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
			},
		}
	case "lognormal":
		return &Distribution{
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

func (d *Distribution) NextValue() float64 {
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

func (d *Distribution) GetPercentile(p float64) float64 {
	switch d.dist {
	case "bernoulli":
		if d.b.Quantile(p) == 0 {
			return d.latency
		} else {
			return d.tail_latency
		}
	case "weibull":
		return d.wb.Quantile(p)
	case "logNormal":
		return d.ln.Quantile(p)
	default:
		return d.latency
	}
}

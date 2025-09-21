package model

type coldStart struct {
	dur float64
}

func newColdStart(d float64) *coldStart {
	return &coldStart{
		dur: d,
	}
}

func (c *coldStart) getColdStart() float64 {
	return c.dur
}

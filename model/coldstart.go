package model

type coldstart struct {
	dur float64
}

func newColdStart(d float64) *coldstart {
	return &coldstart{
		dur: d,
	}
}

func (c *coldstart) getColdStart() float64 {
	return c.dur
}

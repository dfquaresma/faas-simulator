package model

type Config struct {
	ForwardLatency  float64
	Idletime        float64
	TailLatency     float64
	TailLatencyProb string
	Technique       string
}

type traceEntry struct {
	appID    string
	funcID   string
	rows     int64
	startTS  float64
	duration float64
	endTS    float64

	coldStart   *coldStart
	tailLatency *tailLatency
}

type invocationMetadata struct {
	invocationId string

	responseTime          float64
	techniqueResponseTime float64

	forwardedTs float64
	processedTs float64
	shedTimes   int64

	srcInvoc *Invocation
}

type percentile struct {
	p50   float64
	p95   float64
	p99   float64
	p999  float64
	p9999 float64
	p100  float64
}

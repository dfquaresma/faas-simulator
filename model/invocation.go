package model

import (
	"fmt"
	"strconv"
	"strings"
)

type Invocation struct {
	te traceEntry
	im invocationMetadata
}

type traceEntry struct {
	appID    string
	funcID   string
	duration float64
	endTS    float64
	startTS  float64

	coldStart              float64
	mu                     float64
	sigma                  float64
	percentile             percentile
	tlProb                 string
	tail_latency_threshold float64
}

type invocationMetadata struct {
	invocationId          string
	forwardedTs           []float64
	processedTs           []float64
	responseTime          float64
	techniqueResponseTime float64
	hopIds                []string
	hopDelays             []float64

	is_copy       bool
	is_cold_start bool
	fowardTimes   float64
}

type percentile struct {
	p50   float64
	p95   float64
	p99   float64
	p999  float64
	p9999 float64
	p100  float64
}

func NewInvocation(id string, te traceEntry) *Invocation {
	return &Invocation{
		te: te,
		im: invocationMetadata{
			invocationId: id,
			is_copy:      false,
		},
	}
}

func CopyInvocation(i *Invocation) *Invocation {
	return &Invocation{
		te: traceEntry{
			appID:                  i.te.appID,
			funcID:                 i.te.funcID,
			duration:               i.te.duration,
			endTS:                  i.te.endTS,
			startTS:                i.te.startTS,
			percentile:             i.te.percentile,
			tail_latency_threshold: i.te.tail_latency_threshold,
		},
		im: invocationMetadata{
			invocationId: i.im.invocationId,
			forwardedTs:  i.im.forwardedTs,
			processedTs:  i.im.processedTs,
			responseTime: i.im.responseTime,
			hopIds:       []string{},
			hopDelays:    []float64{},
			is_copy:      true,
		},
	}
}

func ToTraceEntry(row []string, tlProb string) (*traceEntry, error) {
	// Row expected format: func,duration,startts,app,endts
	appID := row[0]
	funcID := row[1]

	startTS, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing start_timestamp in row (%v): %q", row, err)
	}

	duration, err := strconv.ParseFloat(row[3], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing duration in row (%v): %q", row, err)
	}

	endTS, err := strconv.ParseFloat(row[4], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing end_timestamp in row (%v): %q", row, err)
	}

	mu, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing mu in row (%v): %q", row, err)
	}

	sigma, err := strconv.ParseFloat(row[6], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing sigma in row (%v): %q", row, err)
	}

	p50, err := strconv.ParseFloat(row[7], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p50 in row (%v): %q", row, err)
	}

	p95, err := strconv.ParseFloat(row[8], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p95 in row (%v): %q", row, err)
	}

	p99, err := strconv.ParseFloat(row[9], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p99 in row (%v): %q", row, err)
	}

	p999, err := strconv.ParseFloat(row[10], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p999 in row (%v): %q", row, err)
	}

	p9999, err := strconv.ParseFloat(row[11], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p9999 in row (%v): %q", row, err)
	}

	p100, err := strconv.ParseFloat(row[12], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p100 in row (%v): %q", row, err)
	}

	var tail_latency_threshold float64
	switch tlProb {
	case "p95":
		tail_latency_threshold = p95
	case "p99":
		tail_latency_threshold = p99
	case "p999":
		tail_latency_threshold = p999
	case "p9999":
		tail_latency_threshold = p9999
	}

	return &traceEntry{
		appID:     appID,
		funcID:    funcID,
		duration:  duration,
		endTS:     endTS,
		startTS:   startTS,
		coldStart: p100,
		mu:        mu,
		sigma:     sigma,
		percentile: percentile{
			p50:   p50,
			p95:   p95,
			p99:   p99,
			p999:  p999,
			p9999: p9999,
			p100:  p100,
		},
		tlProb:                 tlProb,
		tail_latency_threshold: tail_latency_threshold,
	}, nil
}

func (i *Invocation) AddProcessedTs(pt float64) {
	i.im.processedTs = append(i.im.processedTs, pt)
}

func (i *Invocation) UpdateResponse(hopResponse float64, replicaID string) {
	i.im.hopIds = append(i.im.hopIds, replicaID)
	i.im.hopDelays = append(i.im.hopDelays, hopResponse)
	i.im.responseTime += hopResponse
}

func (i *Invocation) hasHops() bool {
	return len(i.im.hopIds) != 0 && len(i.im.hopDelays) != 0
}

func (i *Invocation) IsTailLatency() bool {
	return i.te.duration > i.te.tail_latency_threshold
}

func (i *Invocation) IsCopy() bool {
	return i.im.is_copy
}

func (i *Invocation) IsColdStart() bool {
	return i.im.is_cold_start
}

func (i *Invocation) SetAsColdStart() {
	i.im.is_cold_start = true
}

func (i *Invocation) IncrementFowardTimes() {
	i.im.fowardTimes = i.im.fowardTimes + 1
}

func (i *Invocation) GetFowardTimes() float64 {
	return i.im.fowardTimes
}

func (i *Invocation) GetAppID() string {
	return i.te.appID
}

func (i *Invocation) GetProcessedTs() []float64 {
	return i.im.processedTs
}

func (i *Invocation) GetResponseTime() float64 {
	return i.im.responseTime
}

func (i *Invocation) GetHops() []string {
	return i.im.hopIds
}

func (i *Invocation) GetFuncID() string {
	return i.te.funcID
}

func (i *Invocation) GetDuration() float64 {
	return i.te.duration
}

func (i *Invocation) GetTailLatencyThreshold() float64 {
	return i.te.tail_latency_threshold
}

func (i *Invocation) GetMU() float64 {
	return i.te.mu
}

func (i *Invocation) GetSigma() float64 {
	return i.te.sigma
}

func (i *Invocation) GetColdStart() float64 {
	return i.te.coldStart
}

func (i *Invocation) GetStartTS() float64 {
	return i.te.startTS
}

func (i *Invocation) GetID() string {
	return i.getDatasetID() + i.GetAppID() + i.GetFuncID()
}

func (i *Invocation) getDatasetID() string {
	return i.im.invocationId
}

func (i *Invocation) GetLastHop() string {
	return i.im.hopIds[len(i.im.hopIds)-1]
}

func (i *Invocation) GetLastHopResponse() float64 {
	return i.im.hopDelays[len(i.im.hopDelays)-1]
}

func (i *Invocation) GetOutPut() []string {
	forwardedTsStr := make([]string, len(i.im.forwardedTs))
	for i, f := range i.im.forwardedTs {
		forwardedTsStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	processedTsStr := make([]string, len(i.im.processedTs))
	for i, f := range i.im.processedTs {
		processedTsStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	hopDelaysStr := make([]string, len(i.im.hopDelays))
	for i, f := range i.im.hopDelays {
		hopDelaysStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}

	return []string{
		i.te.appID,
		i.te.funcID,
		strconv.FormatFloat(i.te.duration, 'f', -1, 64),
		strconv.FormatFloat(i.te.endTS, 'f', -1, 64),
		strconv.FormatFloat(i.te.startTS, 'f', -1, 64),
		i.im.invocationId,

		strings.Join(forwardedTsStr, ";"),
		strings.Join(processedTsStr, ";"),
		strconv.FormatFloat(i.im.responseTime, 'f', -1, 64),
		strconv.FormatFloat(i.im.techniqueResponseTime, 'f', -1, 64),

		strconv.FormatFloat(i.te.tail_latency_threshold, 'f', -1, 64),
		strconv.FormatFloat(i.im.fowardTimes, 'f', -1, 64),
		strings.Join(i.im.hopIds, ";"),
		strings.Join(hopDelaysStr, ";"),
	}
}

func (i *Invocation) SetDuration(nd float64) {
	i.te.duration = nd
}

func (i *Invocation) SetForwardedTs(ft float64) {
	i.im.forwardedTs = append(i.im.forwardedTs, ft)
}

func (i *Invocation) UpdateTechniqueResponseTime(rt float64) {
	i.im.techniqueResponseTime = rt
}

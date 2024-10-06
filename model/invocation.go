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
	appID                  string
	funcID                 string
	duration               float64
	endTS                  float64
	startTS                float64
	percentile             percentile
	tlProb                 string
	tail_latency_threshold float64
	is_tail_latency        bool
}

type invocationMetadata struct {
	datasetId       string
	forwardedTs     float64
	processedTs     []float64
	responseTime    float64
	hops            []string
	hopResponses    []float64
	rh_forwardedTs  float64
	rh_processedTs  []float64
	rh_responseTime float64
	rh_hops         []string
	rh_hopResponses []float64
	is_copy         bool
	shed_times      int
}

type percentile struct {
	p90  float64
	p95  float64
	p99  float64
	p100 float64
}

func NewInvocation(id string, te traceEntry) *Invocation {
	return &Invocation{
		te: te,
		im: invocationMetadata{
			datasetId: id,
			is_copy:   false,
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
			is_tail_latency:        i.te.is_tail_latency,
		},
		im: invocationMetadata{
			datasetId:    i.im.datasetId,
			forwardedTs:  i.im.forwardedTs,
			processedTs:  i.im.processedTs,
			responseTime: i.im.responseTime,
			hops:         []string{},
			hopResponses: []float64{},
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

	p90, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p90 in row (%v): %q", row, err)
	}

	p95, err := strconv.ParseFloat(row[6], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p95 in row (%v): %q", row, err)
	}

	p99, err := strconv.ParseFloat(row[7], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p99 in row (%v): %q", row, err)
	}

	p100, err := strconv.ParseFloat(row[8], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing p100 in row (%v): %q", row, err)
	}

	var tail_latency_threshold float64
	switch tlProb {
	case "p90":
		tail_latency_threshold = p90
	case "p95":
		tail_latency_threshold = p95
	case "p99":
		tail_latency_threshold = p99
	}

	return &traceEntry{
		appID:    appID,
		funcID:   funcID,
		duration: duration,
		endTS:    endTS,
		startTS:  startTS,
		percentile: percentile{
			p90:  p90,
			p95:  p95,
			p99:  p99,
			p100: p100,
		},
		tlProb:                 tlProb,
		tail_latency_threshold: tail_latency_threshold,
		is_tail_latency:        duration > tail_latency_threshold,
	}, nil
}

func (i *Invocation) AddProcessedTs(pt float64) {
	i.im.processedTs = append(i.im.processedTs, pt)
}

func (i *Invocation) ResetResponseTime() {
	i.im.responseTime = 0
}

func (i *Invocation) UpdateHops(replicaID string) {
	i.im.hops = append(i.im.hops, replicaID)
}

func (i *Invocation) removeLastHop() {
	if i.hasHops() {
		i.im.hops = i.im.hops[:len(i.im.hops)-1]
		i.im.hopResponses = i.im.hopResponses[:len(i.im.hopResponses)-1]
	}
}

func (i *Invocation) UpdateHopResponse(hopResponse float64) {
	i.im.hopResponses = append(i.im.hopResponses, hopResponse)
	i.im.responseTime += hopResponse
}

func (i *Invocation) UpdateRhInvocationMetadata(rh_forwardedTs float64, rh_processedTs []float64, rh_responseTime float64, rh_hops []string, rh_hopResponses []float64) {
	i.im.rh_forwardedTs = rh_forwardedTs
	i.im.rh_processedTs = rh_processedTs
	i.im.rh_responseTime = rh_responseTime
	i.im.rh_hops = rh_hops
	i.im.rh_hopResponses = rh_hopResponses
}

func (i *Invocation) hasHops() bool {
	return len(i.im.hops) != 0 && len(i.im.hopResponses) != 0
}

func (i *Invocation) IsTailLatency() bool {
	return i.te.is_tail_latency
}

func (i *Invocation) IsCopy() bool {
	return i.im.is_copy
}

func (i *Invocation) IncrementShedTimes() {
	i.im.shed_times = i.im.shed_times + 1
}

func (i *Invocation) GetShedTimes() int {
	return i.im.shed_times
}

func (i *Invocation) GetAppID() string {
	return i.te.appID
}

func (i *Invocation) GetForwardedTs() float64 {
	return i.im.forwardedTs
}

func (i *Invocation) GetProcessedTs() []float64 {
	return i.im.processedTs
}

func (i *Invocation) GetResponseTime() float64 {
	return i.im.responseTime
}

func (i *Invocation) GetHops() []string {
	return i.im.hops
}

func (i *Invocation) GetHopResponses() []float64 {
	return i.im.hopResponses
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

func (i *Invocation) GetP90() float64 {
	return i.te.percentile.p90
}

func (i *Invocation) GetP95() float64 {
	return i.te.percentile.p95
}

func (i *Invocation) GetP99() float64 {
	return i.te.percentile.p99
}

func (i *Invocation) GetP100() float64 {
	return i.te.percentile.p100
}

func (i *Invocation) GetStartTS() float64 {
	return i.te.startTS
}

func (i *Invocation) GetID() string {
	return i.getDatasetID() + i.GetAppID() + i.GetFuncID()
}

func (i *Invocation) getDatasetID() string {
	return i.im.datasetId
}

func (i *Invocation) GetLastProcessedTs() float64 {
	return i.im.processedTs[len(i.im.processedTs)-1]
}

func (i *Invocation) GetLastHop() string {
	return i.im.hops[len(i.im.hops)-1]
}

func (i *Invocation) GetLastHopResponse() float64 {
	return i.im.hopResponses[len(i.im.hopResponses)-1]
}

func (i *Invocation) GetOutPut() []string {
	hopResponsesStr := make([]string, len(i.im.hopResponses))
	for i, f := range i.im.hopResponses {
		hopResponsesStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	rh_hopResponsesStr := make([]string, len(i.im.rh_hopResponses))
	for i, f := range i.im.rh_hopResponses {
		rh_hopResponsesStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	processedTsStr := make([]string, len(i.im.processedTs))
	for i, f := range i.im.processedTs {
		processedTsStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	rh_processedTsStr := make([]string, len(i.im.rh_processedTs))
	for i, f := range i.im.rh_processedTs {
		rh_processedTsStr[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}

	var tl_threshold_accuracy float64
	switch i.te.tlProb {
	case "p90":
		tl_threshold_accuracy = i.te.tail_latency_threshold / i.te.percentile.p90
	case "p95":
		tl_threshold_accuracy = i.te.tail_latency_threshold / i.te.percentile.p95
	case "p99":
		tl_threshold_accuracy = i.te.tail_latency_threshold / i.te.percentile.p99
	}

	return []string{
		i.te.appID,
		i.te.funcID,
		strconv.FormatFloat(i.te.duration, 'f', -1, 64),
		strconv.FormatFloat(i.te.endTS, 'f', -1, 64),
		strconv.FormatFloat(i.te.startTS, 'f', -1, 64),
		i.im.datasetId,
		strconv.FormatFloat(i.im.forwardedTs, 'f', -1, 64),
		strings.Join(processedTsStr, ";"),
		strconv.FormatFloat(i.im.responseTime, 'f', -1, 64),
		strings.Join(i.im.hops, ";"),
		strings.Join(hopResponsesStr, ";"),
		strconv.FormatFloat(i.im.rh_forwardedTs, 'f', -1, 64),
		strings.Join(rh_processedTsStr, ";"),
		strconv.FormatFloat(i.im.rh_responseTime, 'f', -1, 64),
		strings.Join(i.im.rh_hops, ";"),
		strings.Join(rh_hopResponsesStr, ";"),
		strconv.FormatFloat(tl_threshold_accuracy, 'f', -1, 64),
		strconv.FormatFloat(i.te.tail_latency_threshold, 'f', -1, 64),
		strconv.FormatFloat(i.te.percentile.p90, 'f', -1, 64),
		strconv.FormatFloat(i.te.percentile.p95, 'f', -1, 64),
		strconv.FormatFloat(i.te.percentile.p99, 'f', -1, 64),
		strconv.FormatFloat(i.te.percentile.p100, 'f', -1, 64),
	}
}

func (i *Invocation) SetDuration(nd float64) {
	i.te.duration = nd
	i.te.is_tail_latency = nd > i.te.tail_latency_threshold
}

func (i *Invocation) SetForwardedTs(ft float64) {
	i.im.forwardedTs = ft
}

func (i *Invocation) SetTailLatencieThreshold(threshold float64) {
	i.te.tail_latency_threshold = threshold
	i.te.is_tail_latency = i.te.duration > i.te.tail_latency_threshold
}

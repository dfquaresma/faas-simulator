package model

import (
	"fmt"
	"strconv"
)

type Invocation struct {
	te traceEntry
	im invocationMetadata
}

type traceEntry struct {
	appID      string
	funcID     string
	rows       int64
	startTS    float64
	duration   float64
	endTS      float64
	percentile percentile

	coldStart   float64
	tlProb      string
	tlthreshold float64
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

func newInvocation(id string, te traceEntry) *Invocation {
	return &Invocation{
		te: te,
		im: invocationMetadata{
			invocationId: id,
		},
	}
}

func CopyInvocation(i *Invocation) *Invocation {
	return &Invocation{
		te: traceEntry{
			appID:       i.te.appID,
			funcID:      i.te.funcID,
			rows:        i.te.rows,
			startTS:     i.te.startTS,
			duration:    i.te.duration,
			endTS:       i.te.endTS,
			percentile:  i.te.percentile,
			coldStart:   i.te.coldStart,
			tlProb:      i.te.tlProb,
			tlthreshold: i.te.tlthreshold,
		},
		im: invocationMetadata{
			invocationId: i.im.invocationId,
			responseTime: i.im.responseTime,
			srcInvoc:     i,
		},
	}
}

func toTraceEntry(row []string, tlProb string) (*traceEntry, error) {
	// Row expected format: app,func,rows,startts,duration,endts
	appID := row[0]
	funcID := row[1]
	funcRows, err := strconv.ParseInt(row[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing rows in row (%v): %q", row, err)
	}
	startTS, err := strconv.ParseFloat(row[3], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing start_timestamp in row (%v): %q", row, err)
	}
	duration, err := strconv.ParseFloat(row[4], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing duration in row (%v): %q", row, err)
	}
	endTS, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing end_timestamp in row (%v): %q", row, err)
	}

	p50, p95, p99, p999, p9999, p100, err := extractPercentiles(row)
	if err != nil {
		return nil, fmt.Errorf("Error extracting percentiles in row (%v): %q", row, err)
	}

	var tlthreshold float64
	switch tlProb {
	case "p95":
		tlthreshold = p95
	case "p99":
		tlthreshold = p99
	case "p999":
		tlthreshold = p999
	case "p9999":
		tlthreshold = p9999
	}

	return &traceEntry{
		appID:     appID,
		funcID:    funcID,
		rows:      funcRows,
		duration:  duration,
		endTS:     endTS,
		startTS:   startTS,
		coldStart: p100,
		percentile: percentile{
			p50:   p50,
			p95:   p95,
			p99:   p99,
			p999:  p999,
			p9999: p9999,
			p100:  p100,
		},
		tlProb:      tlProb,
		tlthreshold: tlthreshold,
	}, nil
}

func extractPercentiles(row []string) (float64, float64, float64, float64, float64, float64, error) {
	p50, err := strconv.ParseFloat(row[6], 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Error parsing p50 in row (%v): %q", row, err)
	}
	p95, err := strconv.ParseFloat(row[7], 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Error parsing p95 in row (%v): %q", row, err)
	}
	p99, err := strconv.ParseFloat(row[8], 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Error parsing p99 in row (%v): %q", row, err)
	}
	p999, err := strconv.ParseFloat(row[9], 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Error parsing p999 in row (%v): %q", row, err)
	}
	p9999, err := strconv.ParseFloat(row[10], 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Error parsing p9999 in row (%v): %q", row, err)
	}
	p100, err := strconv.ParseFloat(row[11], 64)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Error parsing p100 in row (%v): %q", row, err)
	}

	return p50, p95, p99, p999, p9999, p100, nil
}

func (i *Invocation) IsTailLatency() bool {
	return i.GetDuration() > i.GetTailLatencyThreshold()
}

func (i *Invocation) IsCopy() bool {
	return i.im.srcInvoc != nil
}

func (i *Invocation) IsColdStart() bool {
	return i.GetDuration() >= i.GetColdStart()
}

func (i *Invocation) UpdateSource(src *Invocation) {
	i.im.srcInvoc = src
}

func (i *Invocation) UpdateShedTimes() {
	i.im.shedTimes += 1
}

func (i *Invocation) UpdateResponse(hopResponse float64) {
	i.im.responseTime += hopResponse
}

func (i *Invocation) UpdateTechniqueResponseTime(iCopy *Invocation) {
	i.im.techniqueResponseTime = iCopy.im.processedTs - i.im.forwardedTs
}

func (i *Invocation) SetProcessedTs(pt float64) {
	i.im.processedTs = pt
}

func (i *Invocation) SetForwardedTs(ft float64) {
	i.im.forwardedTs = ft
}

func (i *Invocation) SetAsColdStart() {
	i.SetDuration(i.GetColdStart())
}

func (i *Invocation) SetDuration(nd float64) {
	i.te.duration = nd
}

func (i *Invocation) GetTailLatencyThreshold() float64 {
	return i.te.tlthreshold
}

func (i *Invocation) GetAppID() string {
	return i.te.appID
}

func (i *Invocation) GetFuncID() string {
	return i.te.funcID
}

func (i *Invocation) getRows() int64 {
	return i.te.rows
}
func (i *Invocation) GetInvocationId() string {
	return i.im.invocationId
}

func (i *Invocation) GetDuration() float64 {
	return i.te.duration
}

func (i *Invocation) GetColdStart() float64 {
	return i.te.coldStart
}

func (i *Invocation) GetStartTS() float64 {
	return i.te.startTS
}

func (i *Invocation) GetSrcInvoc() *Invocation {
	return i.im.srcInvoc
}

func (i *Invocation) GetP999() float64 {
	return i.te.percentile.p999
}

func (i *Invocation) getOutPut() []string {
	return []string{
		i.te.appID,
		i.te.funcID,
		i.im.invocationId,

		strconv.FormatFloat(i.te.endTS, 'f', -1, 64),
		strconv.FormatFloat(i.te.startTS, 'f', -1, 64),
		strconv.FormatFloat(i.te.tlthreshold, 'f', -1, 64),

		strconv.FormatFloat(i.te.duration, 'f', -1, 64),
		strconv.FormatFloat(i.im.responseTime, 'f', -1, 64),
		strconv.FormatFloat(i.im.techniqueResponseTime, 'f', -1, 64),

		strconv.FormatInt(i.im.shedTimes, 10),
	}
}

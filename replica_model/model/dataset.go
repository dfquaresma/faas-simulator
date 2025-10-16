package model

import (
	"fmt"
	"strconv"
)

type Dataset struct {
	iLen        int
	iterator    int
	invocations []Invocation
	latencies   map[string][]float64
}

func NewDataSet(rows [][]string, tlProb string) (*Dataset, error) {
	invocs := make([]Invocation, len(rows))
	latencies := make(map[string][]float64)
	tailLatencyCount := 0
	for id, row := range rows {
		entry, err := toTraceEntry(row, tlProb)
		if err != nil {
			return nil, err
		}
		if entry.duration > entry.tailLatency.getTailLatencyThreshold() {
			tailLatencyCount += 1
		}
		invoc := newInvocation(strconv.Itoa(id), *entry)
		list, exists := latencies[invoc.GetAppID()+invoc.GetFuncID()]
		if !exists {
			list = make([]float64, 0, invoc.getRows())
		}
		latencies[invoc.GetAppID()+invoc.GetFuncID()] = append(list, invoc.GetDuration())
		invocs[id] = *invoc
	}

	fmt.Printf(
		"Number of Invocations: %d\nNumber of Tail Latency Reqs: %d\nPercentage Free of Tail Latency: %f\n\n",
		len(invocs),
		tailLatencyCount,
		1-(float64(tailLatencyCount)/float64(len(invocs))),
	)

	return &Dataset{
		iLen:        len(invocs),
		invocations: invocs,
		latencies:   latencies,
	}, nil
}

func toTraceEntry(row []string, tlProb string) (*traceEntry, error) {
	// Row expected format: app,func,rows,startts,duration,endts,p50,p95,p99,p9999,p100
	appID := row[0]
	funcID := row[1]
	funcRows, err := strconv.ParseInt(row[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing rows in row (%v): %q", row, err)
	}
	duration, err := strconv.ParseFloat(row[4], 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing duration in row (%v): %q", row, err)
	}
	startTS, endTS, err := extractTimestamps(row)
	if err != nil {
		return nil, fmt.Errorf("Error extracting timestamps in row (%v): %q", row, err)
	}

	p50, p95, p99, p999, p9999, p100, err := extractPercentiles(row)
	if err != nil {
		return nil, fmt.Errorf("Error extracting percentiles in row (%v): %q", row, err)
	}

	return &traceEntry{
		appID:     appID,
		funcID:    funcID,
		rows:      funcRows,
		duration:  duration,
		endTS:     endTS,
		startTS:   startTS,
		coldStart: newColdStart(p100),
		tailLatency: newTailLatency(
			percentile{
				p50:   p50,
				p95:   p95,
				p99:   p99,
				p999:  p999,
				p9999: p9999,
				p100:  p100,
			},
			tlProb,
		),
	}, nil
}

func extractTimestamps(row []string) (float64, float64, error) {
	startTS, err := strconv.ParseFloat(row[3], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("Error parsing start_timestamp in row (%v): %q", row, err)
	}
	endTS, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("Error parsing end_timestamp in row (%v): %q", row, err)
	}

	return startTS, endTS, nil
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

func (d *Dataset) GetLatenciesOf(id string) []float64 {
	return d.latencies[id]
}

func (d *Dataset) Next() *Invocation {
	if !d.HasNext() {
		return nil
	}
	index := d.iterator
	d.iterator++
	return &d.invocations[index]
}

func (d *Dataset) HasNext() bool {
	return d.iterator < d.iLen
}

func (d *Dataset) GetSize() int {
	return len(d.invocations)
}

func (d *Dataset) GetOutPut() [][]string {
	res := [][]string{}
	header := []string{
		"appID", "funcID", "invocationID",
		"endTS", "startTS", "tl_threshold",
		"duration", "responseTime", "techniqueResponseTime",
		"shedTimes",
	}
	res = append(res, header)
	for _, inv := range d.invocations {
		res = append(res, inv.getOutPut())
	}
	return res
}

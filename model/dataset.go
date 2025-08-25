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
	coldstart   float64
}

func NewDataSet(rows [][]string, tlProb string) (*Dataset, error) {
	invocs := make([]Invocation, len(rows))
	latencies := make(map[string][]float64)
	tailLatencyCount := 0
	for id, row := range rows {
		entry, err := ToTraceEntry(row, tlProb)
		if err != nil {
			return nil, err
		}
		if entry.duration > entry.tlthreshold {
			tailLatencyCount += 1
		}
		invoc := NewInvocation(strconv.Itoa(id), *entry)
		list, exists := latencies[invoc.GetAppID()+invoc.GetFuncID()]
		if !exists {
			list = make([]float64, 0, invoc.GetRows())
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
		"appID", "funcID", "duration", "endTS", "startTS", "invocationID",
		"forwardedTs", "processedTs", "responseTime", "techniqueResponseTime",
		"tl_threshold", "fowardTimes", "hopsId", "hopDelays"}
	res = append(res, header)
	for _, inv := range d.invocations {
		res = append(res, inv.GetOutPut())
	}
	return res
}

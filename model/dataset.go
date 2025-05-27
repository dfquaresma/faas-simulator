package model

import (
	"fmt"
	"strconv"
)

type Dataset struct {
	iLen        int
	iterator    int
	invocations []Invocation
	coldstart   float64
}

func NewDataSet(rows [][]string, tlProb string, hasOracle bool) (*Dataset, error) {
	invocs := make([]Invocation, len(rows))
	tailLatencyCount := 0
	for id, row := range rows {
		entry, err := ToTraceEntry(row, tlProb, hasOracle)
		if err != nil {
			return nil, err
		}
		if entry.duration > entry.tail_latency_threshold {
			tailLatencyCount += 1
		}
		invoc := NewInvocation(strconv.Itoa(id), *entry)
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
	}, nil
}

func (i *Dataset) Next() *Invocation {
	if !i.HasNext() {
		return nil
	}
	index := i.iterator
	i.iterator++
	return &i.invocations[index]
}

func (i *Dataset) HasNext() bool {
	return i.iterator < i.iLen
}

func (i *Dataset) GetSize() int {
	return len(i.invocations)
}

func (i *Dataset) GetOutPut() [][]string {
	res := [][]string{}
	header := []string{"appID", "funcID", "duration", "endTS", "startTS", "invocationID", "forwardedTs", "processedTs", "responseTime", "hopsId", "hopResponses", "rHforwardedTs", "rHprocessedTs", "rHresponseTime", "rHhopsId", "rHhopResponses", "tl_threshold_accuracy", "tl_threshold", "p90", "p95", "p99", "p100"}
	res = append(res, header)
	for _, inv := range i.invocations {
		res = append(res, inv.GetOutPut())
	}
	return res
}

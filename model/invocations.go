package model

import (
	"strconv"
)

type Invocations struct {
	iLen        int
	iterator    int
	invocations []Invocation
}

func NewInvocations(rows [][]string, tlProb string) (*Invocations, error) {
	invocs := make([]Invocation, len(rows))
	for id, row := range rows {
		traceEntry, err := ToTraceEntry(row, tlProb)
		if err != nil {
			return nil, err
		}
		invoc := NewInvocation(strconv.Itoa(id), *traceEntry)
		invocs = append(invocs, *invoc)
	}

	return &Invocations{
		iLen:        len(invocs),
		invocations: invocs,
	}, nil
}

func (i *Invocations) Next() *Invocation {
	if !i.HasNext() {
		return nil
	}
	index := i.iterator
	i.iterator++
	return &i.invocations[index]
}

func (i *Invocations) HasNext() bool {
	return i.iterator < i.iLen
}

func (i *Invocations) GetSize() int64 {
	return int64(len(i.invocations))
}

func (i *Invocations) GetOutPut() [][]string {
	res := [][]string{}
	header := []string{"appID", "funcID", "duration", "endTS", "startTS", "invocationID", "forwardedTs", "processedTs", "responseTime", "hopsId", "hopResponses", "rHforwardedTs", "rHprocessedTs", "rHresponseTime", "rHhopsId", "rHhopResponses", "tl_threshold_accuracy", "tl_threshold", "p90", "p95", "p99", "p100"}
	res = append(res, header)
	for _, inv := range i.invocations {
		res = append(res, inv.GetOutPut())
	}
	return res
}

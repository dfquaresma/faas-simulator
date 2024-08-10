package model

import (
	"strconv"
)

type Dataset struct {
	iLen        int
	iterator    int
	invocations []Invocation
}

func NewDataSet(rows [][]string, tlProb string) (*Dataset, error) {
	invocs := make([]Invocation, len(rows))
	for id, row := range rows {
		entry, err := ToTraceEntry(row, tlProb)
		if err != nil {
			return nil, err
		}
		invoc := NewInvocation(strconv.Itoa(id), *entry)
		invocs[id] = *invoc
	}

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

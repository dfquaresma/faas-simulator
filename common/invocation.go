package common

import (
	"fmt"
	"strconv"
	"strings"
)

type iInvocation interface {
	getAppID() string
	getFuncID() string
	getDuration() float64
	getP95() float64
	getStartTS() float64
	getEndTS() float64
	getID() string
	getOutPut() []string
	setForwardedTs(ft float64)
	addProcessedTs(ft float64)
	getLastProcessedTs() float64
	setDuration(nd float64)
	updateHops(replicaID string)
	getLastHop() string
	updateHopResponse(hopResponse float64)
	getLastHopResponse() float64
	hasHops() bool
	removeLastHop()
	resetResponseTime()
}

type iInvocations interface {
	next() iInvocation
	hasNext() bool
	getSize() int64
	getOutPut() [][]string
}

type invocations struct {
	iLen        int
	iterator    int
	invocations []invocation
}

type invocation struct {
	te traceEntry
	im invocationMetadata
}

type traceEntry struct {
	appID    string
	funcID   string
	duration float64
	p95      float64
	endTS    float64
	startTS  float64
}

type invocationMetadata struct {
	id              string
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
}

func newInvocation(id string, te traceEntry) *invocation {
	return &invocation{
		te: te,
		im: invocationMetadata{id: id},
	}
}

func copyInvocation(i *invocation) *invocation {
	return &invocation{
		te: traceEntry{
			appID:    i.te.appID,
			funcID:   i.te.funcID,
			duration: i.te.duration,
			p95:      i.te.p95,
			endTS:    i.te.endTS,
			startTS:  i.te.startTS,
		},
		im: invocationMetadata{
			id:           i.im.id,
			forwardedTs:  i.im.forwardedTs,
			processedTs:  i.im.processedTs,
			responseTime: i.im.responseTime,
			hops:         []string{},
			hopResponses: []float64{},
		},
	}
}

func (i *invocation) getAppID() string {
	return i.te.appID
}

func (i *invocation) getFuncID() string {
	return i.te.funcID
}

func (i *invocation) getDuration() float64 {
	return i.te.duration
}

func (i *invocation) getP95() float64 {
	return i.te.p95
}

func (i *invocation) getStartTS() float64 {
	return i.te.startTS
}

func (i *invocation) getEndTS() float64 {
	return i.te.endTS
}

func (i *invocation) getID() string {
	return i.im.id
}

func (i *invocation) setForwardedTs(ft float64) {
	i.im.forwardedTs = ft
}

func (i *invocation) addProcessedTs(pt float64) {
	i.im.processedTs = append(i.im.processedTs, pt)
}

func (i *invocation) getLastProcessedTs() float64 {
	return i.im.processedTs[len(i.im.processedTs)-1]
}

func (i *invocation) setDuration(nd float64) {
	i.te.duration = nd
}

func (i *invocation) updateHops(replicaID string) {
	i.im.hops = append(i.im.hops, replicaID)
}

func (i *invocation) getLastHop() string {
	return i.im.hops[len(i.im.hops)-1]
}

func (i *invocation) updateHopResponse(hopResponse float64) {
	i.im.hopResponses = append(i.im.hopResponses, hopResponse)
	i.im.responseTime += hopResponse
}

func (i *invocation) updateRhInvocationMetadata(rh_forwardedTs float64, rh_processedTs []float64, rh_responseTime float64, rh_hops []string, rh_hopResponses []float64) {
	i.im.rh_forwardedTs = rh_forwardedTs
	i.im.rh_processedTs = rh_processedTs
	i.im.rh_responseTime = rh_responseTime
	i.im.rh_hops = rh_hops
	i.im.rh_hopResponses = rh_hopResponses
}

func (i *invocation) getLastHopResponse() float64 {
	return i.im.hopResponses[len(i.im.hopResponses)-1]
}

func (i *invocation) hasHops() bool {
	return len(i.im.hops) != 0 && len(i.im.hopResponses) != 0
}

func (i *invocation) removeLastHop() {
	if i.hasHops() {
		i.im.hops = i.im.hops[:len(i.im.hops)-1]
		i.im.hopResponses = i.im.hopResponses[:len(i.im.hopResponses)-1]
	}
}

func (i *invocation) resetResponseTime() {
	i.im.responseTime = 0
}

func NewInvocations(rows [][]string) (*invocations, error) {
	invocs := make([]invocation, 0)
	for id, row := range rows {
		traceEntry, err := toTraceEntry(row)
		if err != nil {
			return nil, err
		}
		invoc := newInvocation(strconv.Itoa(id), *traceEntry)
		invocs = append(invocs, *invoc)
	}

	return &invocations{
		iLen:        len(invocs),
		invocations: invocs,
	}, nil
}

func (i *invocations) next() *invocation {
	if !i.hasNext() {
		return nil
	}
	index := i.iterator
	i.iterator++
	return &i.invocations[index]
}

func (i *invocations) hasNext() bool {
	return i.iterator < i.iLen
}

func toTraceEntry(row []string) (*traceEntry, error) {
	// Row expected format: func,duration,startts,app,endts

	AppID := row[0]
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

	p95 := 0.0
	//p95, err := strconv.ParseFloat(row[5], 64)
	//if err != nil {
	//	return nil, fmt.Errorf("Error parsing p95 in row (%v): %q", row, err)
	//}

	return &traceEntry{
		appID:    AppID,
		funcID:   funcID,
		duration: duration,
		p95:      p95,
		endTS:    endTS,
		startTS:  startTS,
	}, nil
}

func (i *invocation) getOutPut() []string {
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

	return []string{
		i.te.appID,
		i.te.funcID,
		strconv.FormatFloat(i.te.duration, 'f', -1, 64),
		strconv.FormatFloat(i.te.endTS, 'f', -1, 64),
		strconv.FormatFloat(i.te.startTS, 'f', -1, 64),
		i.im.id,
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
	}
}

func (i *invocations) GetSize() int64 {
	return int64(len(i.invocations))
}

func (i *invocations) GetOutPut() [][]string {
	res := [][]string{}
	header := []string{"appID", "funcID", "duration", "endTS", "startTS", "invocationID", "forwardedTs", "processedTs", "responseTime", "hopsId", "hopResponses", "rHforwardedTs", "rHprocessedTs", "rHresponseTime", "rHhopsId", "rHhopResponses"}
	res = append(res, header)
	for _, inv := range i.invocations {
		res = append(res, inv.getOutPut())
	}
	return res
}

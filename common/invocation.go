package common

type iInvocation interface {
	getAppID() 		string
	getFuncID() 	string
	getDuration() 	float64
	getStartTS() 	float64
	getEndTS() 		float64

	getID() int64
	setForwardedTs(ft float64)
	setProcessedTs(ft float64)
	updateHops(replicaID string)
	updateHopResponse(hopResponse float64)

	getOutPut() []string
}

type invocation struct {
	te	traceEntry
	im	invocationMetadata
}

type traceEntry struct {
	appID       	string
	funcID 			string
	duration        float64
	endTS			float64
	startTS			float64
}

type invocationMetadata struct {
	id          	int64
	forwardedTs		float64
	processedTs		float64
	responseTime	float64
	hops 			[]string
	hopResponses   	[]float64
}

func newInvocation(id int64, te traceEntry) *invocation {
	return &invocation{
		te: te,
		im: &invocationMetadata{id: id},
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

func (i *invocation) getStartTS() float64 {
	return i.te.startTS
}

func (i *invocation) getEndTS() float64 {
	return i.te.endTS
}

func (i *invocation) getID() int64 {
	return i.im.id
}

func (i *invocation) setForwardedTs(ft float64) {
	i.im.forwardedTs = ft
}

func (i *invocation) setProcessedTs(pt float64) {
	i.im.processedTs = pt
}

func (i *invocation) updateHops(replicaID string) {
	i.im.hops = append(i.im.hops, replicaID)
}

func (i *invocation) updateHopResponse(hopResponse float64) {
	i.im.hopResponses = append(i.im.hopResponses, hopResponse)
	i.im.responseTime += hopResponse
}

type iInvocations interface {
	next() 		 iInvocation
	hasNext()    bool
	getOutPut() [][]string
}

type invocations struct {
	iLen	    int64
	iterator	int64
	invocations []iInvocation
}

func newInvocations(rows [][]string) (*invocations, error) {
	invocs := make([]iInvocation, 0)
	for id, row := range rows {
		traceEntry, err := toTraceEntry(row)
		if err != nil {
			return nil, err
		}
		invoc := newInvocation(id, traceEntry)
		invocs = append(invocs, invoc)
	}

	return &invocations{
		iLen: 		 len(invocs),
		invocations: invocs,
	}
}

func (i *invocations) next() *invocation {
	if !i.hasNext() {
		return nil
	}
	index := i.iterator++
	return i.invocations[index]
}

func (i *invocations) hasNext() bool {
	return i.iterator < i.iLen
}

func toTraceEntry(row []string) (traceEntry, error) {
	// Row expected format: func,duration,startts,app,endts 

	AppID  := row[0]
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

	return traceEntryy{Status: status, ResponseTime: responseTime, Body: body, TsBefore: tsbefore, TsAfter: tsafter}, nil
}

func (i *invocation) getOutPut() string {
	return []string{"", "", ""}
}

func (i *invocations) getOutPut() [][]string {
	res := [][]string{} 
	for _, inv := range i.invocations {
		res = append(res, inv.getOutPut())
	}
	return res
}

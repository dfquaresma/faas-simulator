package sim

type iInvocation interface {
	getAppID() 		string
	getFuncID() 	string
	getDuration() 	float64
	getStartTS() 	float64
	getEndTS() 		float64

	getID() int64
	updateHops(replicaID string)
	updateHopResponse(hopResponse float64)
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

func (i *invocation) updateHops(replicaID string) {
	i.im.hops = append(i.im.hops, replicaID)
}

func (i *invocation) updateHopResponse(hopResponse float64) {
	i.im.hopResponses = append(i.im.hopResponses, hopResponse)
	i.im.responseTime += hopResponse
}

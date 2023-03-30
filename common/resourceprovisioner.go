package common

type iResourceProvisioner interface {
	forward(inv iInvocation)
	response(inv iInvocation)
}

type resourceProvisioner struct {
	*godes.Runner
	arrivalQueue	*godes.FIFOQueue
	arrivalCond		*godes.BooleanControl
	appID			string
	funcID			string
	replicas		[]iReplica
	availableRep	int64
}

func newResourceProvisioner(appID, funcID string) *resourceProvisioner {
	return &resourceProvisioner{
		Runner:			&godes.Runner{},
		arrivalQueue:	godes.NewFIFOQueue("arrival"),
		arrivalCond:	godes.NewBooleanControl(),
		appID:			appID,
		funcID:			funcID,
		replicas:		make([]iReplica),
		availableRep: 	-1,	
	}
}

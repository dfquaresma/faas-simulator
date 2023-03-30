package common

type iReplica interface {
}

type replica struct {
	*godes.Runner
	cond 			 	*godes.BooleanControl
	frp					iResourceProvisioner
	idlenessDeadline 	time.Duration
	id               	string
	terminated       	bool
}

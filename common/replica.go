package sim

type iReplica interface {
}

type replica struct {
	*godes.Runner
}
package sim

type iResourceProvisioner interface {
}

type resourceProvisioner struct {
	*godes.Runner
}


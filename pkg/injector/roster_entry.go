package injector

type RosterEntry struct {
	weight uint
	Client
	status Status
}

func (re RosterEntry) Status() Status {
	return re.status
}

func (re RosterEntry) Weight() uint {
	return re.weight
}

package syntheticuser

type StatusHolder struct {
	status            Status
	statusChangeFuncs []StatusChangeFunc
}

type StatusChangeFunc func(oldStatus, newStatus Status)

func NewStatusHolder() StatusHolder {
	return StatusHolder{
		status:            Ready,
		statusChangeFuncs: make([]StatusChangeFunc, 0),
	}
}

func (sh *StatusHolder) Status() Status {
	return sh.status
}

func (sh *StatusHolder) SetStatus(newStatus Status) {
	oldStatus := sh.status
	sh.status = newStatus

	for _, statusChangeFunc := range sh.statusChangeFuncs {
		statusChangeFunc(oldStatus, newStatus)
	}
}

func (sh *StatusHolder) RegisterStatusChangeFunc(fn StatusChangeFunc) {
	sh.statusChangeFuncs = append(sh.statusChangeFuncs, fn)
}

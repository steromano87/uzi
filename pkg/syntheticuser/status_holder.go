package syntheticuser

import "sync"

type StatusHolder struct {
	status            Status
	statusMutex       sync.RWMutex
	statusChangeFuncs []StatusChangeFunc
}

type StatusChangeFunc func(oldStatus, newStatus Status)

func NewStatusHolder() StatusHolder {
	return StatusHolder{
		status:            Status_READY,
		statusChangeFuncs: make([]StatusChangeFunc, 0),
	}
}

func (sh *StatusHolder) Status() Status {
	sh.statusMutex.RLock()
	defer sh.statusMutex.RUnlock()
	return sh.status
}

func (sh *StatusHolder) SetStatus(newStatus Status) {
	sh.statusMutex.Lock()
	defer sh.statusMutex.Unlock()
	oldStatus := sh.status
	sh.status = newStatus

	for _, statusChangeFunc := range sh.statusChangeFuncs {
		statusChangeFunc(oldStatus, newStatus)
	}
}

func (sh *StatusHolder) RegisterStatusChangeFunc(fn StatusChangeFunc) {
	sh.statusChangeFuncs = append(sh.statusChangeFuncs, fn)
}

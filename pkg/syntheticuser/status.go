package syntheticuser

//go:generate stringer -type=Status

type Status int

const (
	Ready Status = iota
	Starting
	Running
	Stopping
	Stopped
	Error
)

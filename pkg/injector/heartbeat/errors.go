package heartbeat

import "errors"

var ErrHeartbeatRequestTimeout = errors.New("timeout waiting heartbeat request")

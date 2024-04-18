package rest

import "fmt"

type ErrBadHTTPStatus struct {
	Status string
}

func (e ErrBadHTTPStatus) Error() string {
	return fmt.Sprintf("bad HTTP status: %s", e.Status)
}

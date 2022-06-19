package rest

import "fmt"

type ErrInvalidPartialUrl struct {
	Url string
}

func (ipu ErrInvalidPartialUrl) Error() string {
	return fmt.Sprintf("'%s' is an invalid partial URL", ipu.Url)
}

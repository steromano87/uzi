package db

import "fmt"

type ErrUnknownType struct {
	unkType string
}

func (e ErrUnknownType) Error() string {
	return fmt.Sprintf("unknown database type: %s", e.unkType)
}

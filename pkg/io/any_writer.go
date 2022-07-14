package io

type AnyWriter[T any] interface {
	Write(T) error
}

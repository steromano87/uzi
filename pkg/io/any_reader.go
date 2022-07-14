package io

type AnyReader[T any] interface {
	Read(any) (T, error)
}

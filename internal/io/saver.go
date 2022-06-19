package io

type Saver interface {
	Save(any) error
}

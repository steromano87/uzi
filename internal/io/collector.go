package io

type Collector[T any] struct {
	saver Saver
}

func (c Collector[T]) Register(saver Saver) {
	c.saver = saver
}

func (c Collector[T]) Collect(item T) error {
	return c.saver.Save(item)
}

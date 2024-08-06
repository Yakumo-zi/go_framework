package main

import (
	"context"
	"strings"
)

type Selector[T any] struct {
	sb   *strings.Builder
	args []any
}

func (s *Selector[T]) Get(ctx *context.Context) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx *context.Context) ([]T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) Build() *Query {
	//TODO implement me
	panic("implement me")
}

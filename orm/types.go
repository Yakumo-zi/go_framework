package main

import (
	"context"
	"database/sql"
)

type Querier[T any] interface {
	Get(ctx *context.Context) (T, error)
	GetMulti(ctx *context.Context) ([]T, error)
}

type QueryBuilder interface {
	Build() *Query
}
type Executor interface {
	Exec(ctx *context.Context) (sql.Result, error)
}

type Query struct {
	SQL  string
	Args []any
}

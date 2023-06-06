package repository

import (
	"context"

	"github.com/pkg/errors"
)

type Repository[T any] interface {
	Ping(ctx context.Context) (pong string, err error)
	Get(ctx context.Context, key string) (val T, err error)
	Set(ctx context.Context, key string, val T) error
	Exists(ctx context.Context, key string) (bool, error)
	Keys(ctx context.Context) ([]string, error)
	Values(ctx context.Context) ([]T, error)
	Size(ctx context.Context) (int64, error)
	Close() error
}

var ErrNotFound = errors.New("no such record")

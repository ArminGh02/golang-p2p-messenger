package redis

import (
	"context"
	"encoding/json"

	"github.com/ArminGh02/golang-p2p-messenger/internal/stun/repository"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	URL string
}

type Redis[T any] struct {
	client *redis.Client
}

func New[T any](cfg *Config) (*Redis[T], error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing redis url %q", cfg.URL)
	}
	client := redis.NewClient(opt)
	return &Redis[T]{client}, nil
}

func (r *Redis[T]) Ping(ctx context.Context) (pong string, err error) {
	return r.client.Ping(ctx).Result()
}

func (r *Redis[T]) Close() error {
	return r.client.Close()
}

func (r *Redis[T]) Get(ctx context.Context, key string) (val T, err error) {
	res, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		err = repository.ErrNotFound
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "error getting value of key %v", key)
		return
	}

	err = json.Unmarshal([]byte(res), &val)
	return
}

func (r *Redis[T]) Set(ctx context.Context, key string, val T) error {
	b, err := json.Marshal(val) // can't we just store and then cast?
	if err != nil {
		return errors.Wrapf(err, "error marshalling value %v", val)
	}

	return r.client.Set(ctx, key, string(b), 0).Err()
}

func (r *Redis[T]) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.Wrapf(err, "error checking if key %v exists", key)
	}

	return count > 0, nil
}

func (r *Redis[T]) Keys(ctx context.Context) (keys []string, err error) {
	iter := r.client.Scan(ctx, 0, "", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err = iter.Err(); err != nil {
		keys = nil
		return
	}
	return
}

func (r *Redis[T]) Values(ctx context.Context) (values []T, err error) {
	size, err := r.Size(ctx)
	if err != nil {
		return
	}

	values = make([]T, 0, size)

	const count = 10
	var (
		keys   = make([]string, 0, count)
		cursor uint64
	)
	for {
		keys, cursor, err = r.client.Scan(ctx, cursor, "*", count).Result()
		if err != nil {
			return
		}

		for _, key := range keys {
			value, err := r.Get(ctx, key)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}

		if cursor == 0 {
			break
		}
	}
	return
}

func (r *Redis[T]) Size(ctx context.Context) (int64, error) {
	keyCount, err := r.client.DBSize(ctx).Result()
	if err != nil {
		return 0, err
	}

	return keyCount, nil
}

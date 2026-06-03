package testkit

import (
	"context"
	"slices"
	"testing"
)

type fakeRedisClient struct {
	flushed bool
	keys    []string
}

func (f *fakeRedisClient) FlushDB(context.Context) error {
	f.flushed = true
	return nil
}

func (f *fakeRedisClient) Del(_ context.Context, keys ...string) error {
	f.keys = append(f.keys, keys...)
	return nil
}

func TestRedisCleanerFlushesWhenNoKeysConfigured(t *testing.T) {
	client := &fakeRedisClient{}
	cleanup := RedisCleaner[struct{}](
		RedisCleanupOptions[struct{}]{
			Client: func(struct{}, *Bootstrap[struct{}]) RedisClient {
				return client
			},
		},
	)

	if err := cleanup(context.Background(), &Bootstrap[struct{}]{}); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}
	if !client.flushed {
		t.Fatal("Redis client was not flushed")
	}
}

func TestRedisCleanerDeletesKeys(t *testing.T) {
	client := &fakeRedisClient{}
	cleanup := RedisCleaner[struct{}](
		RedisCleanupOptions[struct{}]{
			Client: func(struct{}, *Bootstrap[struct{}]) RedisClient {
				return client
			},
			Keys: []string{"a", "b"},
		},
	)

	if err := cleanup(context.Background(), &Bootstrap[struct{}]{}); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}
	if !slices.Equal(client.keys, []string{"a", "b"}) {
		t.Fatalf("keys = %v, want [a b]", client.keys)
	}
}

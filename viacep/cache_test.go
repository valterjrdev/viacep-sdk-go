package viacep

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestViaCep_MemoryCache_Get(t *testing.T) {
	cache := newMemoryCache()
	cache.mu.Lock()
	cache.data["user:1"] = []byte("invalid data")
	cache.mu.Unlock()

	type dummy struct {
		ID   int
		Name string
		Age  int
	}

	model := dummy{ID: 1, Name: "John Doe", Age: 30}

	err := cache.Set(context.Background(), "user:1", model, 0)
	assert.NoError(t, err)

	t.Run("retrieve value with success", func(t *testing.T) {
		var dest dummy
		ctx := context.Background()
		found := cache.Get(ctx, "user:1", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)
	})

	t.Run("key not found", func(t *testing.T) {
		var dest dummy
		ctx := context.Background()
		found := cache.Get(ctx, "user:nonexistent", &dest)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest)
	})

	t.Run("deserialization failure", func(t *testing.T) {
		cache.mu.Lock()
		cache.data["user:invalid"] = []byte("invalid data")
		cache.mu.Unlock()

		var dest dummy

		ctx := context.Background()
		found := cache.Get(ctx, "user:invalid", &dest)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest)
	})
}

func TestViaCep_MemoryCache_Set(t *testing.T) {
	cache := newMemoryCache()

	type dummy struct {
		ID   int
		Name string
		Age  int
	}

	model := dummy{ID: 1, Name: "John Doe", Age: 30}

	t.Run("set and retrieve successfully", func(t *testing.T) {
		err := cache.Set(context.Background(), "user:1", model, 0)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)
	})

	t.Run("serialization error", func(t *testing.T) {
		testCases := []struct {
			value    any
			expected string
		}{
			{make(chan int), "failed to encode value of type chan int: gob NewTypeObject can't handle type: chan int"},
			{func() {}, "failed to encode value of type func(): gob NewTypeObject can't handle type: func()"},
			{map[chan int]int{}, "failed to encode value of type map[chan int]int: gob NewTypeObject can't handle type: chan int"},
			{struct{ x chan int }{make(chan int)}, "failed to encode value of type struct { x chan int }: gob: type struct { x chan int } has no exported fields"},
		}

		for _, tc := range testCases {
			err := cache.Set(context.Background(), "invalid:", tc.value, 0)
			assert.EqualError(t, err, tc.expected)
		}
	})

	t.Run("TTL expiry", func(t *testing.T) {
		err := cache.Set(context.Background(), "user:2", model, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to set cache with TTL: %v", err)
		}

		var dest dummy
		found := cache.Get(context.Background(), "user:2", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)

		time.Sleep(500 * time.Millisecond)

		var dest2 dummy
		found = cache.Get(context.Background(), "user:2", &dest2)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest2)
	})
}

func TestViaCep_MemoryCache_Delete(t *testing.T) {
	cache := newMemoryCache()

	type dummy struct {
		ID   int
		Name string
		Age  int
	}

	model := dummy{ID: 1, Name: "John Doe", Age: 30}

	err := cache.Set(context.Background(), "user:1", model, 0)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	t.Run("delete key", func(t *testing.T) {
		err := cache.Delete(context.Background(), "user:1")
		if err != nil {
			t.Errorf("Expected no error when deleting existing key, but got: %v", err)
		}

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
	})
}

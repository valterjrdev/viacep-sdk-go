package viacep

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestViaCep_MemoryCache_cacheKey(t *testing.T) {
	t.Run("multiple values", func(t *testing.T) {
		expected := "viacep:93046c72a31da34f3f01241343d00bddc8edc3b386ebaef62f3b5083ec6257d9"
		result := cacheKey("part1", "part2", "part3")
		assert.Equal(t, expected, result)
	})

	t.Run("single value", func(t *testing.T) {
		expected := "viacep:947f187506f7629c81c81879a2cb2256455038e4ac770091d897fa0a8b945e3b"
		result := cacheKey("single")
		assert.Equal(t, expected, result)
	})

	t.Run("empty input", func(t *testing.T) {
		expected := "viacep:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		result := cacheKey()
		assert.Equal(t, expected, result)
	})
}

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

	t.Run("deserialization error", func(t *testing.T) {
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
		assert.NoError(t, err)

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
		err := cache.Set(context.Background(), "user:1", model, 10*time.Millisecond)
		assert.NoError(t, err)

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)

		time.Sleep(40 * time.Millisecond)

		var dest2 dummy
		found = cache.Get(context.Background(), "user:1", &dest2)
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
	assert.NoError(t, err)

	t.Run("delete key", func(t *testing.T) {
		err := cache.Delete(context.Background(), "user:1")
		assert.NoError(t, err)

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
	})
}

func TestViaCep_RedisCache_Get(t *testing.T) {
	type dummy struct {
		ID   int
		Name string
		Age  int
	}

	model := dummy{ID: 1, Name: "John Doe", Age: 30}

	t.Run("retrieve value with success", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(model)
		assert.NoError(t, err)

		mock.ExpectGet("user:1").SetVal(buffer.String())

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("key not found", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		mock.ExpectGet("user:1").RedisNil()

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error get value", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		mock.ExpectGet("user:1").SetErr(errors.New("error"))

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("deserialization error", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		mock.ExpectGet("user:1").SetVal("invalid data")

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
		assert.Equal(t, dummy{ID: 0, Name: "", Age: 0}, dest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViaCep_RedisCache_Set(t *testing.T) {
	type dummy struct {
		ID   int
		Name string
		Age  int
	}

	model := dummy{ID: 1, Name: "John Doe", Age: 30}

	t.Run("set and retrieve successfully", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(model)
		assert.NoError(t, err)

		mock.ExpectSet("user:1", buffer.Bytes(), 0).SetVal("OK")
		mock.ExpectGet("user:1").SetVal(buffer.String())

		err = cache.Set(context.Background(), "user:1", model, 0)
		assert.NoError(t, err)

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error set key", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(model)
		assert.NoError(t, err)

		mock.ExpectSet("user:1", buffer.Bytes(), 0).SetErr(errors.New("error"))

		err = cache.Set(context.Background(), "user:1", model, 0)
		assert.EqualError(t, err, "failed to set value in cache: error")

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("serialization error", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

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

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TTL expiry", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(model)
		assert.NoError(t, err)

		TTL := 10 * time.Millisecond
		mock.ExpectSet("user:1", buffer.Bytes(), TTL).SetVal("OK")
		mock.ExpectGet("user:1").SetVal(buffer.String())

		err = cache.Set(context.Background(), "user:1", model, TTL)
		assert.NoError(t, err)

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.True(t, found)
		assert.Equal(t, model, dest)

		time.Sleep(40 * time.Millisecond)

		var dest2 dummy
		found = cache.Get(context.Background(), "user:1", &dest2)
		assert.False(t, found)
		assert.Equal(t, dummy{}, dest2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViaCep_RedisCache_Delete(t *testing.T) {
	type dummy struct {
		ID   int
		Name string
		Age  int
	}

	model := dummy{ID: 1, Name: "John Doe", Age: 30}

	t.Run("delete key", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(model)
		assert.NoError(t, err)

		mock.ExpectSet("user:1", buffer.Bytes(), 0).SetVal("OK")
		mock.ExpectDel("user:1").SetVal(1)

		err = cache.Set(context.Background(), "user:1", model, 0)
		assert.NoError(t, err)

		err = cache.Delete(context.Background(), "user:1")
		assert.NoError(t, err)

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
	})

	t.Run("error delete key", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		cache := NewRedisCache(client)

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(model)
		assert.NoError(t, err)

		mock.ExpectSet("user:1", buffer.Bytes(), 0).SetVal("OK")
		mock.ExpectDel("user:1").SetErr(errors.New("error"))

		err = cache.Set(context.Background(), "user:1", model, 0)
		assert.NoError(t, err)

		err = cache.Delete(context.Background(), "user:1")
		assert.EqualError(t, err, "failed to delete key from cache: error")

		var dest dummy
		found := cache.Get(context.Background(), "user:1", &dest)
		assert.False(t, found)
	})
}

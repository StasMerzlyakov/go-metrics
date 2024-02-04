package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt64MemStorage(t *testing.T) {
	ms := NewMemoryInt64Storage()
	ms.Add("m", 1)
	ms.Add("m", 2)
	k, v := ms.Get("m")
	assert.True(t, v)
	assert.Equal(t, int64(3), k)

	ms.Set("m", 5)
	k, v = ms.Get("m")
	assert.True(t, v)
	assert.Equal(t, int64(5), k)

	_, v = ms.Get("m2")
	assert.False(t, v)
}

func TestFloat64MemStorage(t *testing.T) {
	ms := NewMemoryFloat64Storage()
	ms.Add("m", 1)
	ms.Add("m", 2)
	k, v := ms.Get("m")
	assert.True(t, v)
	assert.Equal(t, float64(3), k)

	ms.Set("m", 5)
	k, v = ms.Get("m")
	assert.True(t, v)
	assert.Equal(t, float64(5), k)

	ms.Set("m", 7)
	k, v = ms.Get("m")
	assert.True(t, v)
	assert.Equal(t, float64(7), k)

	_, v = ms.Get("m2")
	assert.False(t, v)
}

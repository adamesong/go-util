package redis

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var r *RedisClient

// init sets up the Redis client for all tests in this package.
// It panics if the connection fails, as tests cannot run without Redis.
func init() {
	var err error
	r, err = NewRedisClient("localhost:6379", "", 0)
	if err != nil {
		panic(fmt.Sprintf("test setup failed: could not create redis client: %v", err))
	}
}

// TestSetAndGet tests Set and Get operations.
func TestSetAndGet(t *testing.T) {
	type someStruct struct {
		IntAttr    int
		StringAttr string
	}
	key := "test:setget"
	val := someStruct{IntAttr: 123, StringAttr: "hello"}
	valBytes, err := json.Marshal(val)
	require.NoError(t, err)

	// Test Set
	err = r.Set(key, valBytes, 10*time.Second)
	require.NoError(t, err)

	// Test Get
	retrievedBytes, err := r.Get(key)
	require.NoError(t, err)
	assert.Equal(t, valBytes, retrievedBytes)

	var retrievedVal someStruct
	err = json.Unmarshal(retrievedBytes, &retrievedVal)
	require.NoError(t, err)
	assert.Equal(t, val, retrievedVal)

	// Test Get on non-existent key
	_, err = r.Get("non-existent-key")
	assert.Equal(t, redis.Nil, err)

	// Cleanup
	_, err = r.Delete(key)
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	keys := []string{"test:del:1", "test:del:2"}

	// Ensure keys are not present
	r.Delete(keys...)

	// Set keys
	for _, k := range keys {
		require.NoError(t, r.Set(k, "val", 10*time.Second))
	}

	// Delete them
	deletedCount, err := r.Delete(keys...)
	require.NoError(t, err)
	assert.Equal(t, int64(len(keys)), deletedCount)

	// Verify they are gone
	for _, k := range keys {
		_, err := r.Get(k)
		assert.Equal(t, redis.Nil, err)
	}

	// Test deleting non-existent keys
	deletedCount, err = r.Delete("non-existent-key")
	require.NoError(t, err)
	assert.Equal(t, int64(0), deletedCount)
}

func TestSetNXAndSetXX(t *testing.T) {
	key := "test:nx-xx"

	// Ensure key is not present
	r.Delete(key)

	// SetNX on non-existent key should succeed
	ok, err := r.SetNX(key, "val1", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)

	// SetNX on existing key should fail
	ok, err = r.SetNX(key, "val2", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, ok)

	// SetXX on existing key should succeed
	ok, err = r.SetXX(key, "val3", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)

	// Verify value was updated by SetXX
	val, err := r.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "val3", string(val))

	// Cleanup
	r.Delete(key)

	// SetXX on non-existent key should fail
	ok, err = r.SetXX(key, "val4", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestGetSet(t *testing.T) {
	key := "test:getset"

	// Ensure key is not present
	r.Delete(key)

	// GetSet on non-existent key should return redis.Nil error
	oldVal, err := r.GetSet(key, "new-val")
	assert.Equal(t, redis.Nil, err)
	assert.Empty(t, oldVal)

	// The key should now exist with "new-val"
	currentVal, err := r.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "new-val", string(currentVal))

	// GetSet on existing key should return old value
	oldVal, err = r.GetSet(key, "final-val")
	require.NoError(t, err)
	assert.Equal(t, "new-val", string(oldVal))

	// The key should now have "final-val"
	currentVal, err = r.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "final-val", string(currentVal))

	// Cleanup
	r.Delete(key)
}

func TestTTLAndExpire(t *testing.T) {
	keyNoExpire := "test:ttl:noexpire"
	keyExpire := "test:ttl:expire"
	keyNotExist := "test:ttl:notexist"

	// Cleanup before test
	r.Delete(keyNoExpire, keyExpire, keyNotExist)

	// Key with no expiration
	require.NoError(t, r.Set(keyNoExpire, "val", 0))
	ttl, err := r.TTL(keyNoExpire)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl, "TTL for key with no expiration should be -1")

	// Key with expiration
	require.NoError(t, r.Set(keyExpire, "val", 20*time.Second))
	ttl, err = r.TTL(keyExpire)
	require.NoError(t, err)
	assert.True(t, ttl > 10*time.Second && ttl <= 20*time.Second, "TTL should be around 20s")

	// Expire on existing key
	ok, err := r.Expire(keyExpire, 30*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)
	ttl, err = r.TTL(keyExpire)
	require.NoError(t, err)
	assert.True(t, ttl > 20*time.Second && ttl <= 30*time.Second, "New TTL should be around 30s")

	// TTL on non-existent key
	ttl, err = r.TTL(keyNotExist)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-2), ttl, "TTL for non-existent key should be -2")

	// Expire on non-existent key
	ok, err = r.Expire(keyNotExist, 10*time.Second)
	require.NoError(t, err)
	assert.False(t, ok)

	// Cleanup
	r.Delete(keyNoExpire, keyExpire)
}

func TestLikeDeletes(t *testing.T) {
	keys := []string{
		"test:likedelete:param:1",
		"test:likedelete:paRAm:2",
		"test:likedelete:param_abc:3",
		"test:likedelete:abc_param:4",
		"test:likedelete:123_param_456:5",
		"test:likedelete:something_else:6",
	}

	// Cleanup before test
	r.Delete(keys...)

	// Set all keys
	for _, k := range keys {
		require.NoError(t, r.Set(k, "val", 10*time.Second))
	}

	// Delete keys containing "param" (case-sensitive)
	deletedCount, err := r.LikeDeletes("param")
	require.NoError(t, err)
	assert.Equal(t, int64(4), deletedCount, "should delete 4 keys containing 'param'")

	// Verify keys that should be deleted
	_, err = r.Get("test:likedelete:param:1")
	assert.Equal(t, redis.Nil, err)
	_, err = r.Get("test:likedelete:param_abc:3")
	assert.Equal(t, redis.Nil, err)
	_, err = r.Get("test:likedelete:abc_param:4")
	assert.Equal(t, redis.Nil, err)
	_, err = r.Get("test:likedelete:123_param_456:5")
	assert.Equal(t, redis.Nil, err)


	// Verify keys that should NOT be deleted
	val, err := r.Get("test:likedelete:paRAm:2")
	assert.NoError(t, err)
	assert.Equal(t, "val", string(val))

	val, err = r.Get("test:likedelete:something_else:6")
	assert.NoError(t, err)
	assert.Equal(t, "val", string(val))


	// Cleanup
	r.Delete(keys...)
}

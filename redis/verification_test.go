package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestVerification creates a new Verification instance for testing.
// It relies on the global 'r' client initialized in redis_test.go.
func getTestVerification() *Verification {
	return &Verification{
		Redis:             r,
		EmailCodeLength:   6,
		EmailCodeTimeout:  time.Hour * 1,
		MobileCodeLength:  6,
		MobileCodeTimeout: time.Minute * 5,
	}
}

func TestGetSetUserEmailVerifyCode(t *testing.T) {
	v := getTestVerification()
	doAssertion := assert.New(t)
	userID := "1"
	email := "xxx@gmail.com"
	key := GetUserEmailCacheKey(userID, email)

	// Teardown: ensure key is deleted after test
	defer func() {
		_, _ = r.Delete(key)
	}()
	// Setup: ensure key does not exist before test
	_, _ = r.Delete(key)

	// First time setting the code
	code, err := v.GetSetUserEmailVerifyCode(userID, email)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	d, err := r.TTL(key)
	require.NoError(t, err)
	fmt.Println("first duration: ", d.String())

	// 5 seconds later, set it again
	time.Sleep(time.Second * 5)

	codeSecond, err := v.GetSetUserEmailVerifyCode(userID, email)
	require.NoError(t, err)

	d2, err := r.TTL(key)
	require.NoError(t, err)

	// Because it was reset, the new TTL should be greater than the original TTL minus 5 seconds.
	doAssertion.Greater(d2.Seconds(), (time.Hour - time.Second*5).Seconds())
	fmt.Println("second duration: ", d2.String())

	// The two codes should be identical
	fmt.Println("code1: ", code, " code2: ", codeSecond)
	doAssertion.Equal(code, codeSecond)
}

func TestGetSetEmailVerifyCode(t *testing.T) {
	v := getTestVerification()
	doAssertion := assert.New(t)
	email := "xxx@gmail.com"
	key := GetEmailCacheKey(email)

	// Teardown: ensure key is deleted after test
	defer func() {
		_, _ = r.Delete(key)
	}()
	// Setup: ensure key does not exist before test
	_, _ = r.Delete(key)

	// First time setting the code
	code, err := v.GetSetEmailVerifyCode(email)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	d, err := r.TTL(key)
	require.NoError(t, err)
	fmt.Println("first duration: ", d.String())

	// 5 seconds later, set it again
	time.Sleep(time.Second * 5)

	codeSecond, err := v.GetSetEmailVerifyCode(email)
	require.NoError(t, err)

	d2, err := r.TTL(key)
	require.NoError(t, err)

	// Because it was reset, the new TTL should be greater than the original TTL minus 5 seconds.
	doAssertion.Greater(d2.Seconds(), (time.Hour - time.Second*5).Seconds())
	fmt.Println("second duration: ", d2.String())

	// The two codes should be identical
	fmt.Println("code1: ", code, " code2: ", codeSecond)
	doAssertion.Equal(code, codeSecond)
}

func TestVerifyCode(t *testing.T) {
	v := getTestVerification()
	doAssertion := assert.New(t)
	key := "test:verifycode"
	code := "123456"

	// Teardown
	defer func() {
		_, _ = r.Delete(key)
	}()
	// Setup
	_, _ = r.Delete(key)

	// 1. Set the code
	err := r.Set(key, code, 5*time.Second)
	require.NoError(t, err)

	// 2. Verify with correct code, should succeed
	ok, err := v.VerifyCode(key, code)
	doAssertion.NoError(err)
	doAssertion.True(ok, "verification with correct code should succeed")

	// 3. Check if the key was deleted
	_, err = r.Get(key)
	doAssertion.Equal(redis.Nil, err, "key should be deleted after successful verification")

	// 4. Set it again for another test case
	err = r.Set(key, code, 5*time.Second)
	require.NoError(t, err)

	// 5. Verify with incorrect code, should fail
	ok, err = v.VerifyCode(key, "654321")
	doAssertion.NoError(err)
	doAssertion.False(ok, "verification with incorrect code should fail")

	// 6. Check that the key was NOT deleted
	_, err = r.Get(key)
	doAssertion.NoError(err, "key should not be deleted after failed verification")
}

package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var v = Verification{
	Redis:             &r,
	EmailCodeLength:   6,
	EmailCodeTimeout:  time.Hour * 1,
	MobileCodeLength:  6,
	MobileCodeTimeout: time.Minute * 5,
}

func TestGetSetUserEmailVerifyCode(t *testing.T) {
	doAssertion := assert.New(t)
	userID := "1"
	email := "xxx@gmail.com"

	// 先确保redis中没有此cache key
	key := GetUserEmailCacheKey(userID, email)
	r.Delete(key)

	// 第一次设置
	code, _ := v.GetSetUserEmailVerifyCode(userID, email)
	if d, ttlErr := r.TTL(key); ttlErr != nil {
		t.Error(ttlErr.Error())
	} else {
		fmt.Println("first duration: ", d.String())
	}

	// 5秒后第二次设置
	time.Sleep(time.Second * 5)

	codeSecond, _ := v.GetSetUserEmailVerifyCode(userID, email)
	if d2, ttl2Err := r.TTL(key); ttl2Err != nil {
		t.Error(ttl2Err.Error())
	} else {
		// 由于重新设置，d应该是1小时，而不会是1小时-5秒，所以d应该大于1小时减5秒
		doAssertion.Greater(d2.Seconds(), time.Duration(time.Hour-time.Second*5).Seconds())
		fmt.Println("second duration: ", d2.String())
	}

	// 两次获得的code应该相同
	fmt.Println("code1: ", code, " code2: ", codeSecond)
	doAssertion.Equal(code, codeSecond)

	// tear down
	r.Delete(key)
}

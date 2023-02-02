package redis

import (
	"time"

	"github.com/adamesong/go-util/random"
)

const (
	MOBILE_VERI_PREFIX = "verify_mobile:"
	EMAIL_VERI_PREFIX  = "verify_email:"
)

// GetUserEmailCacheKey 生成userId+email验证码的cache的key，需要userID
func GetUserEmailCacheKey(userID string, email string) string {
	return EMAIL_VERI_PREFIX + ":" + userID + ":" + email
}

// GetEmailCacheKey 生成email验证码的cache的key，不需要userID
func GetEmailCacheKey(email string) string {
	return EMAIL_VERI_PREFIX + ":" + email
}

// GetMobileCacheKey 生成mobile验证码的cache的key
func GetMobileCacheKey(mobile string) string {
	return MOBILE_VERI_PREFIX + mobile
}

// verification struct
type Verification struct {
	Redis             *RedisClient
	EmailCodeLength   int           // ie: 6
	EmailCodeTimeout  time.Duration // ie: time.hour * 24
	MobileCodeLength  int           // ie: 6
	MobileCodeTimeout time.Duration // ie: time.minute *30
}

// SetUserEmailVerifyCode 在缓存中保存Email验证码，有效时间在config.ini中设置，需要userID
func (v *Verification) SetUserEmailVerifyCode(userID string, email string) (string, error) {
	code := random.RandomString(v.EmailCodeLength)
	err := v.Redis.Set(GetUserEmailCacheKey(userID, email), code, v.EmailCodeTimeout)
	return code, err
}

// 先尝试获取缓存中保存的Email验证码，如没有，则新设置，效果同SetUserEmailVerifyCode；如有，则取出此code，并重新设置此code的有效期。
// 这样保证多次设置 UserEmailVerifyCode 获得的code都是相同的
func (v *Verification) GetSetUserEmailVerifyCode(userID, email string) (code string, err error) {
	key := GetUserEmailCacheKey(userID, email)

	// 先获取
	if value, redisErr := v.Redis.Get(key); redisErr != nil {
		// 如没有获取到，redisErr为“redis: nil”，code为空string
		code = ""
	} else {
		code = string(value)
	}

	// 设置
	if code == "" {
		code = random.RandomString(v.EmailCodeLength)
	}

	err = v.Redis.Set(key, code, v.EmailCodeTimeout)
	return code, err
}

// SetEmailVerifyCode 在缓存中保存Email验证码，有效时间在config.ini中设置，不需要userID
func (v *Verification) SetEmailVerifyCode(email string) (string, error) {
	code := random.RandomNumber(v.EmailCodeLength)
	err := v.Redis.Set(GetEmailCacheKey(email), code, v.EmailCodeTimeout)
	return code, err
}

func (v *Verification) GetSetEmailVerifyCode(email string) (code string, err error) {
	key := GetEmailCacheKey(email)

	// 先获取
	if value, redisErr := v.Redis.Get(key); redisErr != nil {
		// 如没有获取到，redisErr为“redis: nil”，code为空string
		code = ""
	} else {
		code = string(value)
	}

	// 设置
	if code == "" {
		code = random.RandomNumber(v.EmailCodeLength)
	}

	err = v.Redis.Set(key, code, v.EmailCodeTimeout)
	return code, err
}

// SetMobileVerifyCode 在缓存中保存Mobile验证码，有效时间在config.ini中设置
func (v *Verification) SetMobileVerifyCode(mobile string) (string, error) {
	code := random.RandomNumber(v.MobileCodeLength)
	err := v.Redis.Set(GetMobileCacheKey(mobile), code, v.MobileCodeTimeout)
	return code, err
}

// 验证缓存中的验证码（email或者mobile)
func (v *Verification) VerifyCode(key, code string) (bool, error) {
	if value, err := v.Redis.Get(key); err != nil {
		return false, err
	} else {
		if string(value) == code {
			return true, err
		} else {
			return false, err
		}
	}
}

// 验证手机验证码
func (v *Verification) VerifyMobile(mobile, code string) (bool, error) {
	key := GetMobileCacheKey(mobile)
	return v.VerifyCode(key, code)
}

// 验证userID+email验证码
func (v *Verification) VerifyUserEmail(userID string, email, code string) (bool, error) {
	key := GetUserEmailCacheKey(userID, email)
	return v.VerifyCode(key, code)
}

// 验证email验证码，不含userID
func (v *Verification) VerifyEmail(email, code string) (bool, error) {
	key := GetEmailCacheKey(email)
	return v.VerifyCode(key, code)
}

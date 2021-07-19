package jwt

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

// JwtSign 结构体
type JwtSign struct {
	SigningKey []byte
}

// 一些常量
var (
	TokenExpired     = errors.New("token is expired")
	TokenNotValidYet = errors.New("token not active yet")
	TokenMalformed   = errors.New("that's not even a token")
	TokenInvalid     = errors.New("couldn't handle this token")
)

// CustomClaims 用于构成payload
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// 新建一个JwtSign实例
func NewJwtSign(jwtSecret string) *JwtSign {
	return &JwtSign{
		[]byte(jwtSecret),
	}
}

// CreateToken 生成一个token
func (j *JwtSign) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken 解析一个token
func (j *JwtSign) ParseToken(tokenString string) (*CustomClaims, error) {
	// 因为我们只使用一个私钥来签署令牌，所以我们也只使用它的公共计数器部分来验证
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		} else {
			return nil, TokenInvalid
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

// RefreshToken 更新一个token
// func (j *JwtSign) RefreshToken(tokenString string) (string, error) {
// 	jwtDuration, _ := time.ParseDuration(conf.JwtDuration)
// 	jwt.TimeFunc = func() time.Time {
// 		return time.Unix(0, 0)
// 	}
// 	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
// 		return j.SigningKey, nil
// 	})
// 	if err != nil {
// 		return "", err
// 	}
// 	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
// 		jwt.TimeFunc = time.Now
// 		claims.StandardClaims.ExpiresAt = time.Now().Add(jwtDuration).Unix()
// 		return j.CreateToken(*claims)
// 	}
// 	return "", TokenInvalid
// }

// GetClaimsFromExpiredToken 从一个过期的token中获取claims
func (j *JwtSign) GetClaimsFromExpiredToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
				// 如果token是过期了，仍然返回claims，为了从过期的token中获取userID
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				if claims, ok := token.Claims.(*CustomClaims); ok { // 这里删掉了token.valid
					return claims, nil
				}
				return nil, TokenExpired
			} else {
				return nil, TokenInvalid
			}
		} else {
			return nil, TokenInvalid
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

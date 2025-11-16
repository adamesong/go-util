package jwt

/*
// !!! Deprecated !!!
此文件已被弃用。
建议用户在自己的程序中实现 JWT 逻辑，或使用后端模板自动生成。
*/

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
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
	jwt.RegisteredClaims
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
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, TokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, TokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, TokenNotValidYet
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
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if claims, ok := token.Claims.(*CustomClaims); ok {
				return claims, nil
			}
			return nil, TokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, TokenMalformed
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, TokenNotValidYet
		}
		return nil, TokenInvalid
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, TokenInvalid
}

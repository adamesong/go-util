package refresh_token

import (
	"encoding/json"
	"time"

	"github.com/adamesong/go-util/logging"

	"github.com/google/uuid"
)

type RefreshTokenConfig struct {
	Duration time.Duration // ie: time.Hour * 720
}

// ['refresh_token', '过期的datetime', 'ip', '设备信息', 'Browser信息']
type RefreshToken struct {
	Token     string    `json:"refresh_token"`
	ExpiresAt time.Time `json:"expires_at"`
	IP        string    `json:"ip"`
	Platform  string    `json:"platform"`
	Browser   string    `json:"browser"`
}

// NotExpired 一个refresh_token是否过了有效期
func (token RefreshToken) Expired() bool {
	// refreshDuration, _ := time.ParseDuration(conf.RefreshTokenDuration)
	return time.Now().After(token.ExpiresAt)
}

// 一组refresh token的结构体，用于存在用户privateInfo上
type RefreshTokens struct {
	Tokens []RefreshToken `json:"refresh_tokens"`
}

func (cfg *RefreshTokenConfig) NewRefreshToken(ip, platform, browser string) RefreshToken {

	refreshToken := RefreshToken{
		Token:     uuid.NewString(),
		ExpiresAt: time.Now().Add(cfg.Duration),
		IP:        ip,
		Platform:  platform,
		Browser:   browser,
	}
	return refreshToken
}

// GetTokenAndUpdate
// 提供用户登录的ip、平台、浏览器(不做判断)，获得refresh_token(string)，并根据ip、平台是否相同，来更新有效的refresh_token list
// 此过程中，会删除过期的refresh_token。
// 如果删除了过期的refresh_token，或者生成了新的refresh_token，则updated返回true(老数据被改变)，否则false。
func (tokens *RefreshTokens) GetTokenAndUpdate(ip, platform, browser string, config *RefreshTokenConfig) ( //newRefreshTokens RefreshTokens,
	refreshToken string, updated bool) {
	updated = false
	newTokens := make([]RefreshToken, 0) // 用于存更新后的数据
	for _, token := range tokens.Tokens {
		if !token.Expired() { // 如果该refresh_token没有过期
			if token.IP == ip && token.Platform == platform { // 如果ip和平台都一致，则返回这个refresh_token
				refreshToken = token.Token
				newTokens = append(newTokens, token)
			} else { // if token.IP != ip || token.Platform != platform  // 如果不一致，则继续存在list中
				newTokens = append(newTokens, token)
			}
		} else { // 如果tokens里的某条token过期了，删除过期的token，即不append到新数据里
			updated = true
		}
	}
	if refreshToken == "" { // 如果循环完了array没有找到符合条件的老token，则生成一个新token，并append到list中
		newToken := config.NewRefreshToken(ip, platform, browser)
		newTokens = append(newTokens, newToken)
		refreshToken = newToken.Token
		updated = true
	}

	tokens.Tokens = newTokens
	return
}

func (tokens *RefreshTokens) GetMarshaledTokens() []byte {
	data, err := json.Marshal(tokens.Tokens)
	if err != nil {
		logging.Fatal(err.Error())
	}
	return data
}

// 仅判断tokens列表中的token是否包含tokenString，不判断这个token是否过期
func (tokens *RefreshTokens) ContainsTokenString(tokenString string) bool {
	for _, token := range tokens.Tokens {
		if token.Token == tokenString {
			return true
		}
	}
	return false
}

// 仅判断tokens列表中的token是否包含tokenString，且判断这个token是否过期
func (tokens *RefreshTokens) ContainsValidTokenString(tokenString string) bool {
	for _, token := range tokens.Tokens {
		if token.Token == tokenString && token.ExpiresAt.After(time.Now()) {
			return true
		}
	}
	return false
}

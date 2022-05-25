package wechat

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/adamesong/go-util/logging"
	"github.com/adamesong/go-util/redis"
	"github.com/adamesong/go-util/signature"
	"github.com/pkg/errors"
)

const (
	AccessTokenCachePrefix = "wechat_public_access_token:" // 用于缓存 access_token(微信公众号接口调用凭证，缓存2小时) 的prefix，与微信公众号AppID组合成为cache key
	JSApiTicketCachePrefix = "wechat_public_jsapi_ticket:" // 用户缓存 jsapi_ticket(公众号用于调用微信 JS 接口的临时票据。缓存2小时) 的prefix，与微信公众号AppID组合成为 cache key
)

type WeichatPublicDev struct {
	AppID       string             // 微信公众号后台-基本配置-公众号开发信息-开发者ID(AppID)
	AppSecret   string             // 微信公众号后台-基本配置-公众号开发信息-开发者密码(AppSecret)
	RedisClient *redis.RedisClient // 获取的Access Token将被缓存到哪里
	// AccessTokenCacheKey string             // access token在缓存中的key
	// JSApiTicketCacheKey string             // jsapi_ticket 在缓存中的key
}

// 将prefix 与appID组合成 access_token的cache key
func (wx *WeichatPublicDev) AccessTokenCacheKey() string {
	return fmt.Sprintf("%s%s", AccessTokenCachePrefix, wx.AppID)
}

// 将prefix 与appID组合成 jsapi_ticket的cache key
func (wx *WeichatPublicDev) JSApiTicketCacheKey() string {
	return fmt.Sprintf("%s%s", JSApiTicketCachePrefix, wx.AppID)
}

// GetAccessToken，access_token是公众号的全局唯一接口调用凭据，公众号调用各接口时都需使用access_token。
// https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Get_access_token.html
func (wx *WeichatPublicDev) GetAccessToken() (token string, err error) {
	// 先从缓存中检查是否有有效的accessToken，
	expDuration, err := wx.RedisClient.TTL(wx.AccessTokenCacheKey())
	if err != nil {
		return "", err
	}

	// 如果有效期大大于20秒，则从缓存中取出这个值，并返回
	if expDuration.Seconds() >= 20 {
		if tokenByte, err := wx.RedisClient.Get(wx.AccessTokenCacheKey()); err != nil {
			return "", err
		} else {
			return string(tokenByte), nil
		}
	}

	// 有效期如果低于20秒，则重新请求微信接口，避免频繁请求导致微信拒绝调用
	return wx.FetchAccessToken()
}

// 从微信获取新的access token（不从缓存中获取）
func (wx *WeichatPublicDev) FetchAccessToken() (token string, err error) {
	if resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", wx.AppID, wx.AppSecret)); err != nil {
		return "", err
	} else {
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return "", err
		} else {

			respMap := make(map[string]interface{})
			if err := json.Unmarshal(body, &respMap); err != nil {
				return "", err
			}

			if token, exist := respMap["access_token"].(string); exist {
				// 获取成功后，存入缓存，并返回
				if err := wx.RedisClient.Set(wx.AccessTokenCacheKey(), token, time.Second*time.Duration(respMap["expires_in"].(float64))); err != nil {
					return "", err
				} else {
					return token, nil
				}
			}

			logging.Error(fmt.Sprintf("Getting wechat public account access_token error: %v %v", respMap["errcode"], respMap["errmsg"]))
			return "", fmt.Errorf("Error: %v %v", respMap["errcode"], respMap["errmsg"])
		}
	}
}

// 获得jsapi_ticket, jsapi_ticket是公众号用于调用微信 JS 接口的临时票据
func (wx *WeichatPublicDev) GetJSApiTicket() (ticket string, err error) {
	// 先从缓存中检查是否有有效的jsapi_ticket
	expDuration, err := wx.RedisClient.TTL(wx.JSApiTicketCacheKey())
	if err != nil {
		return "", err
	}

	// 如果有效期大大于20秒，则从缓存中取出这个值，并返回
	if expDuration.Seconds() >= 20 {
		if ticketByte, err := wx.RedisClient.Get(wx.JSApiTicketCacheKey()); err != nil {
			return "", err
		} else {
			return string(ticketByte), nil
		}
	}

	// 有效期如果低于20秒，则重新请求微信接口，避免频繁请求导致微信拒绝调用
	return wx.FetchJSApiTicket()
}

// 从微信获取新的jsapi_ticket（不从缓存中获取）
func (wx *WeichatPublicDev) FetchJSApiTicket() (ticket string, err error) {

	// 先获得access_token
	accessToken, err := wx.GetAccessToken()
	if err != nil {
		return "", err
	}

	// 用access_token获得jsapi_ticket
	if resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi", accessToken)); err != nil {
		return "", err
	} else {
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return "", err
		} else {
			// {
			// "errcode":0,
			// "errmsg":"ok",
			// "ticket":"bxLdikRXVbTPdHSM05e5u5sUoXNKd8-41ZO3MhKoyN5OfkWITDGgnr2fwJ0m9E8NYzWKVZvdVtaUgWvsdshFKA",
			// "expires_in":7200
			// }
			respMap := make(map[string]interface{})
			if err := json.Unmarshal(body, &respMap); err != nil {
				return "", err
			}

			if respMap["errmsg"] != "ok" {
				return "", errors.Errorf("Error: %v %v", respMap["errcode"], respMap["errmsg"])
			}

			if ticket, exist := respMap["ticket"].(string); exist {

				// 获取成功后，存入缓存，并返回
				if err := wx.RedisClient.Set(wx.JSApiTicketCacheKey(), ticket, time.Second*time.Duration(respMap["expires_in"].(float64))); err != nil {
					return "", err
				} else {
					return ticket, nil
				}
			}
			logging.Error(fmt.Sprintf("Getting wechat public account jsapi_ticket error: %v %v", respMap["errcode"], respMap["errmsg"]))
			return "", fmt.Errorf("Error: %v %v", respMap["errcode"], respMap["errmsg"])
		}
	}
}

// 用于签名的结构体
type SignParams struct {
	Nonce       string `sign:"noncestr"`
	JSApiTicket string `sign:"jsapi_ticket"`
	Timestamp   string `sign:"timestamp"`
	URL         string `sign:"url"`
}

// 签名算法：
// 签名生成规则如下：参与签名的字段包括noncestr（随机字符串）, 有效的jsapi_ticket, timestamp（时间戳）, url（当前网页的URL，不包含#及其后面部分） 。
// 对所有待签名参数按照字段名的ASCII 码从小到大排序（字典序）后，使用 URL 键值对的格式（即key1=value1&key2=value2…）拼接成字符串string1。
// 这里需要注意的是所有参数名均为小写字符。对string1作sha1加密，字段名和字段值都采用原始值，不进行URL 转义。即signature=sha1(string1)。

// noncestr=Wm3WZYTPz0wzccnW
// jsapi_ticket=sM4AOVdWfPE4DxkXGEs8VMCPGGVi4C3VM0P37wVUCFvkVAy_90u5h9nbSlYy3-Sl-HhTdfl2fzFy1AOcHKP7qg
// timestamp=1414587457
// url=http://mp.weixin.qq.com?params=value
func (signParams *SignParams) GetSign() (sign string) {
	strToSign := signature.GetValidStr(*signParams)

	// 方法1:
	signByte := sha1.Sum([]byte(strToSign))
	sign = hex.EncodeToString(signByte[:])

	// 方法2:
	// hash := sha1.New()
	// hash.Write([]byte(strToSign))
	// md := hash.Sum(nil)
	// mdStr := hex.EncodeToString(md)
	// sign = strings.ToLower(mdStr)

	return
}

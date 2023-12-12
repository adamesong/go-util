package signature

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adamesong/go-util/redis"
	"github.com/google/uuid"
)

const (
	ErrorNoQueryParam     = "ErrorNoQueryParam"
	ErrorWrongAppKey      = "ErrorWrongAppKey"
	ErrorNoAppKey         = "ErrorNoAppKey"
	ErrorNoTimestamp      = "ErrorNoTimestamp"
	ErrorWrongTimestamp   = "ErrorWrongTimestamp"
	ErrorInvalidTimestamp = "ErrorInvalidTimestamp"
	ErrorFutureTimestamp  = "ErrorFutureTimestamp"
	ErrorTSExpired        = "ErrorTSExpired"
	ErrorNonceTooShort    = "ErrorNonceTooShort"
	ErrorNonceTooLong     = "ErrorNonceTooLong"
	ErrorNoSignature      = "ErrorNoSignature"
	ErrorWrongSign        = "ErrorWrongSign"
	ErrorNonceExist       = "ErrorNonceExist"
	ErrorCheckNonce       = "ErrorCheckNonce"

	// 默认的签名有效期：
	DEFAULT_SIGN_DURATION = time.Second * 300

	SIGN_NONCE_PREFIX = "sign_nonce:" // API请求时所带的用于计算签名的一次性随机字符串
)

// GetValidStr 提供一个结构体的实例，得到用于生成签名的原始字符串
// 方法参考微信支付：https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=4_3
// 1.参数以字典序排序
// 2.如果参数的值为空不参与签名
// 3.参数名和参数值区分大小写
// 3.参数之间以&连接，is the original value instead of url encoded value，不要转为url encoded value。
// 4.除本package的结构体外，任意结构体都可用于签名，只需结构体中参与签名的参数名加tag: sign:"partner_code"
// 5.sign参数不参与签名，仅将生成的签名与该sign值做校验
// 例如：valid_string = partner_code=xxx&time=xxx&nonce_str=xxx&credential_code=xxx
//
//	例如，提供struct{
//			PartnerCode string `sign:"partner_code"`
//			Time string  `sign:"time"`              // UTC毫秒时间戳，取当前UTC时间的毫秒数时间戳，Long类型，5分钟内有效
//			NonceStr string  `sign:"nonce_str"`
//			CredentialCode string  `sign:"credential_code"`
//
// 注意：struct中的各项都需要是string
func GetValidStr(queryObj interface{}) (validStr string) {
	// 从queryObj(某结构体)中获取用于签名的项
	var strList []string
	s := reflect.TypeOf(queryObj)
	v := reflect.ValueOf(queryObj)
	for i := 0; i < s.NumField(); i++ {
		// 如果请求的该参数不为空，则加入到params中
		if v.Field(i).String() != "" {
			// 如果结构体的某项有tag "sign"，才加入params中
			if paramName := s.Field(i).Tag.Get("sign"); paramName != "" {
				strList = append(strList, paramName+"="+v.Field(i).String())
			}
		}
	}
	// 按字典序排序
	sort.Strings(strList)
	// 拼接
	for i, v := range strList {
		if i != 0 {
			validStr += "&" + v
		} else {
			validStr += v
		}
	}
	return validStr
}

// 签名规则（与下面的func的签名结果不同）
// 1. 拼接API密钥匙 valid_str + "&key=xxxxx"
// 2. SHA256进行签名，并转为Hex小写字符串
func ValidStrToSign(validStr, key string) (sign string) {
	validStr = validStr + "&key=" + key
	hash := sha256.New()
	hash.Write([]byte(validStr))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	sign = strings.ToLower(mdStr)
	return
}

// 签名规则（与下面的func的签名结果不同）
// 1. 拼接API密钥匙 valid_str + "&key=xxxxx"
// 2. HMAC-SHA256进行签名，并转为Hex小写字符串
func ValidStrToSignHMACSHA256(validStr, key string) (sign string) {
	validStr = validStr + "&key=" + key
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(validStr))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	sign = strings.ToLower(mdStr)
	return
}

// 使用HMAC-SHA256算法，传入as(AppSecret)计算签名 sign = base64(HmacSHA256(as,strToSign))
// appSecret: 分配给app或web的密钥，以此作为加密的key。
func StrToSignHMACSHA256Base64(strToSign, appSecret string) (sign string) {
	key := []byte(appSecret)
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(strToSign))
	sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

// Deprecated: use (*SignVerifyOption)GetStrToSign instead.
// 调用api时的签名计算func
// urlPath: 例如/v1/articles/15  不包含query参数
// reqMethod: GET, DELETE, POST, PUT, PATCH
// reqForm: http包中的request.Form，在 调用 _ = c.Request.ParseForm() 之后，参数将会解析到Form中; 测试时可包装成url.Values
// reqForm中需要包含的参数有ak, ts, nc
// reqBody: 如果请求是POST或PUT或PATCH，body中的json_body
// appKeyAndSecret：包含所有appKey和appSecret的map，形式如：{"xxxx(app_key_1)": "xxxx(app_secret_1)", "xxxx(app_key_2)": "xxxx(app_secret_2)"}
// signDuration：timestamp距离现在是否超过有效期，如这里提供0，则用默认值300秒
// strToSign: 计算签名前的字符串
// errCode: 自定义的错误编码
// success: 是否成功获拼接出 strToSign
func GetStrToSign(urlPath, reqMethod string, reqForm url.Values, reqBody []byte, appKeyAndSecret map[string]string, signDuration time.Duration) (strToSign, errCode string, success bool) {

	// ak: appKey,用来识别调用方身份 （不是AppSecret，用来加密生成签名。）
	// ts: timestamp, unix timestamp,10位,秒
	// nc: nonce, nonce,32-50位的一次性随机字符串
	var ak, ts, nc string               // sn:signature, ak:appKey, ts:timestamp, nc:nonce, as:appSecret
	params := make(map[string][]string) // 仅将query的参数存入(参数中包含ak, ts, ns, 不含sn)
	var keys []string                   // 用来存params的key，用于给key按字典序排序

	if reqForm == nil {
		errCode = ErrorNoQueryParam
		return
	}
	// 这里不用reqForm.Get("ak") 是因为 Get("ak")得到的类型是[]string
	ak = strings.Join(reqForm["ak"], "") // AppKey 用来识别调用方身份 （不是AppSecret，用来加密生成签名。）
	// 判断AppKey是否存在，如果不存在，返回失败
	if ak == "" || appKeyAndSecret[ak] == "" {
		errCode = ErrorWrongAppKey
		return
	}
	// else {
	// 	// 如果存在，取出AppSecret（用于后面计算签名）
	// 	as = conf.AppSecretAndKey[ak]
	// }
	// ! 这里与判断签名func不同的是：这里未从reqForm中取sn(signature)
	ts = strings.Join(reqForm["ts"], "") // 表示时间戳，用来验证接口的时效性。
	// 判断是否有timestamp参数
	if ts == "" {
		errCode = ErrorNoTimestamp
		return
	}
	// 判断timestamp是否是合法的数字
	tsSeconds, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		errCode = ErrorInvalidTimestamp
		return
	}
	tsTime := time.Unix(tsSeconds, 0)
	// 如果timestamp在现在之后(即请求还未发生)，则返回失败
	if tsTime.After(time.Now()) {
		errCode = ErrorFutureTimestamp
		return
	}
	// 判断timestamp距离现在是否超过有效期，如果超过，则返回失败
	if signDuration == 0 {
		signDuration = DEFAULT_SIGN_DURATION
	}
	if tsTime.Add(signDuration).Before(time.Now()) {
		errCode = ErrorTSExpired
		return
	}

	nc = strings.Join(reqForm["nc"], "")
	// 判断nc的长度是否长过32个字符，如果短，则返回失败
	if len(nc) < 32 {
		errCode = ErrorNonceTooShort
		return
	} else if len(nc) > 50 {
		errCode = ErrorNonceTooLong
		return
	}

	for k, v := range reqForm {
		// (3). query中的参数剔除无value的参数，剔除sn参数(signature)
		if k != "sn" {
			keys = append(keys, k)
			// v是[]string，如果v无值，则跳过这个参数
			if len(v) == 0 {
				continue
			} else {
				params[k] = v
			}
		}
	}
	// (5).将所有的query参数的参数名按字典序排序
	sort.Strings(keys)

	// 将query参数拼接至用于计算签名的字符串
	for i, key := range keys {
		// for i := 0; i < len(keys); i++ {
		var strPart string
		// 如果query中的参数仅一个，直接拼接，例如:state = 1
		if len(params[key]) == 1 {
			// 这里同时将 queryEscape后的 "+" 替换成 "%20"
			strPart = fmt.Sprintf("%v=%v", key, strings.ReplaceAll(url.QueryEscape(params[key][0]), "+", "%20"))
		} else if len(params[key]) > 1 {
			// 如果query中的某参数有若干值，则先排序后，顺序拼接，例如：space=1&space=2&space=3
			vArray := params[key]
			sort.Strings(vArray)
			for vI, vStr := range vArray {
				// 这里同时将 queryEscape后的 "+" 替换成 "%20"
				percentEcodedStr := strings.ReplaceAll(url.QueryEscape(vStr), "+", "%20")
				if vI == 0 {
					strPart = fmt.Sprintf("%v=%v", key, percentEcodedStr)
				} else {
					strPart = strPart + fmt.Sprintf("&%v=%v", key, percentEcodedStr)
				}
			}
		}
		// 然后再拼接至strToSign
		if i == 0 {
			strToSign = urlPath + "\n" + strPart
		} else {
			strToSign = strToSign + "&" + strPart
		}
	}

	// 如果请求是POST或PUT或PATCH，则还需要处理body中的json_body
	if reqMethod == "POST" || reqMethod == "PUT" || reqMethod == "PATCH" {
		// base64(md5(json_body)) 然后拼接至strToSign
		// md5
		md5JsonBody := md5.New()
		_, _ = io.WriteString(md5JsonBody, string(reqBody))
		md5JsonBodyStr := fmt.Sprintf("%x", md5JsonBody.Sum(nil))
		// base64加密之后，拼接至 strToSign
		strToSign = strToSign + "\n" + base64.StdEncoding.EncodeToString([]byte(md5JsonBodyStr))
	}

	success = true
	return
}

// Deprecated: use (*SignVerifyOption)VerifySign instead.
// 验证调用api的签名是否有效，签名sn已经在reqForm中了，参数名为"sn"
// sign: 通过参数计算出来的签名，用于与请求中的签名sn做对比
func VerifySign(urlPath, reqMethod string, reqForm url.Values, reqBody []byte, appKeyAndSecret map[string]string, signDuration time.Duration, redisClient *redis.RedisClient) (strToSign, errCode, sign string, success bool) {
	strToSign, _, success = GetStrToSign(urlPath, reqMethod, reqForm, reqBody, appKeyAndSecret, signDuration)

	// 获得appKey。这里不再判断appKey是否存在，因为在GetStrToSign()中已经做了判断
	ak := strings.Join(reqForm["ak"], "")
	as := appKeyAndSecret[ak]

	// 如果获得strToSign成功，则计算签名
	if success {
		sign = StrToSignHMACSHA256Base64(strToSign, as)
	}

	sn := strings.Join(reqForm["sn"], "") // 表示签名加密串，用来验证数据的完整性，防止数据篡改。
	// 判断是否有sn，如果没有，则返回失败
	if sn == "" {
		errCode = ErrorNoSignature
		success = false
		return
	}
	// (10). 如果计算出来的签名与request中query中的签名不一致，则返回失败
	if sign != sn {
		errCode = ErrorWrongSign
		success = false
		return
	}

	// 不再判断nc的长度是否长过32个字符，因为在GetStrToSign()中已经做了判断
	nc := strings.Join(reqForm["nc"], "")
	// (11). 如果一致，则从redis中判断以nonce值是否存在（有有效期），如果存在，说明之前已经请求过，返回失败
	// (12). 如果redis中nonce值不存在，说明未重复请求过(nonce过期的问题已经在之前的timestamp处判断过)，则 在缓存中存入此nonce，并返回成功
	if signDuration == 0 {
		signDuration = DEFAULT_SIGN_DURATION
	}
	ncKey := SIGN_NONCE_PREFIX + nc
	cacheSuccess, cacheErr := redisClient.SetNX(ncKey, 1, signDuration)
	if !cacheSuccess {
		errCode = ErrorNonceExist
		success = false
		return
	}
	if cacheErr != nil {
		errCode = ErrorCheckNonce
		success = false
		return
	}

	errCode = ""
	success = true
	return

}

// Deprecated: use (*SignVerifyOption)GetTestSign instead.
// 生成测试用的api signature，并返回签名后的url.Values
func GetTestSign(urlPath, reqMethod string, reqForm url.Values, bodyJson []byte, appKeyForTest string, appKeyAndSecret map[string]string) (sign, signedUri string, signedForm url.Values) {
	ak := appKeyForTest
	as := appKeyAndSecret[ak]
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	uuid, _ := uuid.NewRandom()
	nc := uuid.String()

	reqForm.Add("ak", ak)
	reqForm.Add("ts", ts)
	reqForm.Add("nc", nc)

	strToSign, errCode, success := GetStrToSign(urlPath, reqMethod, reqForm, bodyJson, appKeyAndSecret, DEFAULT_SIGN_DURATION)
	if !success {
		fmt.Println("err in signature: ", errCode)
	}
	sign = StrToSignHMACSHA256Base64(strToSign, as)

	reqForm.Add("sn", sign)
	signedForm = reqForm
	signedUri = urlPath + "?" + signedForm.Encode()
	return
}

// SignBody 签名的body
type SignBody struct {
	UrlPath       string     // 例如/v1/articles/15  不包含query参数
	RequestMethod string     // GET, DELETE, POST, PUT, PATCH
	ReqForm       url.Values // http包中的request.Form，在 调用 _ = c.Request.ParseForm() 之后，参数将会解析到Form中; 测试时可包装成url.Values。需要包含的参数有ak, ts, nc。如不需要每次签名都唯一，可仅包含ak
	ReqBodyJson   []byte     // reqBody: 如果请求是POST或PUT或PATCH，body中的json_body
}

// Signature Option 生成签名时所需的配置
type SignOption struct {
	AppKeyAndSecret map[string]string // 所支持的appKey和对应的appSecret，map key为appKey, value为appSecret
	UniqueSign      bool              // 如果为true，则app key、timestamp和nonce都会参与签名，同时signDuration、redisClient这两项为必要项；如为false，则不考虑ts和nc，仅用ak来参与签名
	SignDuration    time.Duration     // 签名中的timestamp距离现在的有效期，如这里为0，则默认为300秒
}

// Signature Verification Option 验证签名所需的配置
type SignVerifyOption struct {
	AppKeyAndSecret map[string]string  // 所支持的appKey和对应的appSecret，map key为appKey, value为appSecret
	UniqueSign      bool               // 如果为true，则app key、timestamp和nonce都会参与签名，同时signDuration、redisClient这两项为必要项；如为false，则不考虑ts和nc，仅用ak来参与签名
	SignDuration    time.Duration      // 签名中的timestamp距离现在的有效期，如这里为0，则默认为300秒
	RedisClient     *redis.RedisClient // 用于存取nonce的的redis客户端
	RedisKeyPrefix  string             // redis中nonce的key的前章，默认为"sign_nonce_"

}

// 可替代上面的 GetStrToSign() function，与其目的相同，不同之处：
// - 增加了ts和nc不参与签名的签名方式
func (option *SignOption) GetStrToSign(body *SignBody) (strToSign string, errCode string, success bool) {
	// ak: appKey,用来识别调用方身份 （不是AppSecret，用来加密生成签名。）
	// ts: timestamp, unix timestamp,10位,秒
	// nc: nonce, nonce,32-50位的一次性随机字符串
	params := make(map[string][]string) // 仅将query的参数存入(参数中包含ak, ts, ns, 或仅包含ak, 不含sn)
	var keys []string                   // 用来存params的key，用于给key按字典序排序

	if body.ReqForm == nil {
		errCode = ErrorNoQueryParam
		return
	}
	// 这里不用reqForm.Get("ak") 是因为 Get("ak")得到的类型是[]string
	ak := strings.Join(body.ReqForm["ak"], "") // AppKey 用来识别调用方身份 （不是AppSecret，用来加密生成签名。）

	// 判断AppKey是否存在，如果不存在，返回失败
	if ak == "" {
		errCode = ErrorNoAppKey
		return
	}
	// ! 这里与判断签名func不同的是：这里未从reqForm中取sn(signature)
	if option.UniqueSign {
		ts := strings.Join(body.ReqForm["ts"], "") // 表示时间戳，用来验证接口的时效性。
		// 判断是否有timestamp参数
		if ts == "" {
			errCode = ErrorNoTimestamp
			return
		}
		// 判断timestamp是否是合法的数字
		tsSeconds, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			errCode = ErrorInvalidTimestamp
			return
		}
		tsTime := time.Unix(tsSeconds, 0)
		// 如果timestamp在现在之后(即请求还未发生)，则返回失败
		if tsTime.After(time.Now()) {
			errCode = ErrorFutureTimestamp
			return
		}
		// 判断timestamp距离现在是否超过有效期，如果超过，则返回失败
		if option.SignDuration == 0 {
			option.SignDuration = DEFAULT_SIGN_DURATION
		}
		if tsTime.Add(option.SignDuration).Before(time.Now()) {
			errCode = ErrorTSExpired
			return
		}

		nc := strings.Join(body.ReqForm["nc"], "")
		// 判断nc的长度是否长过32个字符，如果短，则返回失败
		if len(nc) < 32 {
			errCode = ErrorNonceTooShort
			return
		} else if len(nc) > 50 {
			errCode = ErrorNonceTooLong
			return
		}
	}
	for k, v := range body.ReqForm {
		// (3). query中的参数剔除无value的参数，剔除sn参数(signature)
		if k != "sn" {
			keys = append(keys, k)
			// v是[]string，如果v无值，则跳过这个参数
			if len(v) == 0 {
				continue
			} else {
				params[k] = v
			}
		}
	}
	// (5).将所有的query参数的参数名按字典序排序
	sort.Strings(keys)

	// 将query参数拼接至用于计算签名的字符串
	for i, key := range keys {
		// for i := 0; i < len(keys); i++ {
		var strPart string
		// 如果query中的参数仅一个，直接拼接，例如:state = 1
		if len(params[key]) == 1 {
			// 这里同时将 queryEscape后的 "+" 替换成 "%20"
			strPart = fmt.Sprintf("%v=%v", key, strings.ReplaceAll(url.QueryEscape(params[key][0]), "+", "%20"))
		} else if len(params[key]) > 1 {
			// 如果query中的某参数有若干值，则先排序后，顺序拼接，例如：space=1&space=2&space=3
			vArray := params[key]
			sort.Strings(vArray)
			for vI, vStr := range vArray {
				// 这里同时将 queryEscape后的 "+" 替换成 "%20"
				percentEcodedStr := strings.ReplaceAll(url.QueryEscape(vStr), "+", "%20")
				if vI == 0 {
					strPart = fmt.Sprintf("%v=%v", key, percentEcodedStr)
				} else {
					strPart = strPart + fmt.Sprintf("&%v=%v", key, percentEcodedStr)
				}
			}
		}
		// 然后再拼接至strToSign
		if i == 0 {
			strToSign = body.UrlPath + "\n" + strPart
		} else {
			strToSign = strToSign + "&" + strPart
		}
	}

	// 如果请求是POST或PUT或PATCH，则还需要处理body中的json_body
	if body.RequestMethod == "POST" || body.RequestMethod == "PUT" || body.RequestMethod == "PATCH" {
		// base64(md5(json_body)) 然后拼接至strToSign
		// md5
		md5JsonBody := md5.New()
		_, _ = io.WriteString(md5JsonBody, string(body.ReqBodyJson))
		md5JsonBodyStr := fmt.Sprintf("%x", md5JsonBody.Sum(nil))
		// base64加密之后，拼接至 strToSign
		strToSign = strToSign + "\n" + base64.StdEncoding.EncodeToString([]byte(md5JsonBodyStr))
	}

	errCode = ""
	success = true

	return

}

// 可替代上面的 GetStrToSign() function，与其目的相同，不同之处：
// 1. 增加了ts和nc不参与签名的签名方式
// 2. 可自定义nonce在缓存中的cache key prefix
func (option *SignVerifyOption) VerifySign(body *SignBody) (success bool, errCode string) {
	signOption := SignOption{
		AppKeyAndSecret: option.AppKeyAndSecret,
		UniqueSign:      option.UniqueSign,
		SignDuration:    option.SignDuration,
	}

	strToSign, errCode, success := signOption.GetStrToSign(body)

	// 如果获得strToSign失败，则直接返回
	if !success {
		return
	}

	// 获得appKey。并判断appKey是否存在
	ak := strings.Join(body.ReqForm["ak"], "")
	if ak == "" || option.AppKeyAndSecret[ak] == "" {
		errCode = ErrorWrongAppKey
		success = false
		return
	}

	as := option.AppKeyAndSecret[ak]

	sign := StrToSignHMACSHA256Base64(strToSign, as)

	sn := strings.Join(body.ReqForm["sn"], "") // 表示签名加密串，用来验证数据的完整性，防止数据篡改。
	// 判断是否有sn，如果没有，则返回失败
	if sn == "" {
		errCode = ErrorNoSignature
		success = false
		return
	}
	// (10). 如果计算出来的签名与request中query中的签名不一致，则返回失败
	if sign != sn {
		errCode = ErrorWrongSign
		success = false
		return
	}

	// 如果是ts和nc都参与的签名验证，则对ts和nc的有效性做判断
	if option.UniqueSign {
		// 不再判断nc的长度是否长过32个字符，因为在GetStrToSign()中已经做了判断
		nc := strings.Join(body.ReqForm["nc"], "")
		// (11). 如果一致，则从redis中判断以nonce值是否存在（有有效期），如果存在，说明之前已经请求过，返回失败
		// (12). 如果redis中nonce值不存在，说明未重复请求过(nonce过期的问题已经在之前的timestamp处判断过)，则 在缓存中存入此nonce，并返回成功
		if option.SignDuration == 0 {
			option.SignDuration = DEFAULT_SIGN_DURATION
		}
		ncKey := SIGN_NONCE_PREFIX + nc
		if option.RedisKeyPrefix != "" {
			ncKey = option.RedisKeyPrefix + ncKey
		}
		cacheSuccess, cacheErr := option.RedisClient.SetNX(ncKey, 1, option.SignDuration)
		if !cacheSuccess {
			errCode = ErrorNonceExist
			success = false
			return
		}
		if cacheErr != nil {
			errCode = ErrorCheckNonce
			success = false
			return
		}
	}
	success = true
	return
}

// 生成测试用的api signature，并返回签名后的url.Values
func (option *SignOption) GetTestSign(body *SignBody, appKeyForTest string) (signedUri, sign string, signedForm url.Values) {
	ak := appKeyForTest
	as := option.AppKeyAndSecret[ak]
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	uuid, _ := uuid.NewRandom()
	nc := uuid.String()

	signedForm = body.ReqForm
	body.ReqForm.Add("ak", ak)

	if option.UniqueSign {
		body.ReqForm.Add("ts", ts)
		body.ReqForm.Add("nc", nc)
	}

	strToSign, errCode, success := option.GetStrToSign(body)

	if !success {
		fmt.Println("err in signature: ", errCode)
		return
	}
	sign = StrToSignHMACSHA256Base64(strToSign, as)

	signedForm.Add("sn", sign)
	signedUri = body.UrlPath + "?" + signedForm.Encode()
	return
}

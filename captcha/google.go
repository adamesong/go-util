package captcha

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Captcha interface {
	Verify() (verified bool, errCodes string)
}

type GoogleCaptcha struct {
	Token     string
	Remoteip  string
	SecretKey string // google reCaptcha Secret Key
}

// https://developers.google.com/recaptcha/docs/verify
// Each reCAPTCHA user response token is valid for two minutes, and can only be verified once to prevent replay attacks. If you need a new token, you can re-run the reCAPTCHA verification.
// 注意：坑：post请求的参数不在body里，不是json
func (c *GoogleCaptcha) Verify() (verified bool, errCodes string) {
	type respBody struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}

	// postUrl := "https://www.google.com/recaptcha/api/siteverify"
	postUrl := "https://www.recaptcha.net/recaptcha/api/siteverify" // https://developers.google.com/recaptcha/docs/faq#can-i-use-recaptcha-globally
	param := url.Values{}
	param.Add("secret", c.SecretKey)
	param.Add("response", c.Token)
	param.Add("remoteip", c.Remoteip)

	// resp, err := http.Post(postUrl, "application/json", bytes.NewBuffer(bodyJson))
	resp, err := http.Post(postUrl+"?"+param.Encode(), "application/json", nil)

	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err.Error()
	}

	var resBody respBody
	if err := json.Unmarshal(respByte, &resBody); err != nil {
		return false, err.Error()
	}
	fmt.Println(strings.Join(resBody.ErrorCodes, ","))
	if resBody.Success {
		return true, strings.Join(resBody.ErrorCodes, ",")
	} else {
		return false, strings.Join(resBody.ErrorCodes, ",")
	}
}

package signature_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/adamesong/go-util/signature"
)

func TestGetStrToSign(t *testing.T) {
	// Prepare the SignVerifyOption instance
	option := &signature.SignOption{
		AppKeyAndSecret: map[string]string{
			"testAppKey": "testAppSecret",
		},
		UniqueSign:   false,
		SignDuration: 300 * time.Second,
	}

	// Prepare the SignBody instance
	reqForm := url.Values{}
	reqForm.Set("ak", "testAppKey")
	body := &signature.SignBody{
		UrlPath:       "/v1/articles/15",
		RequestMethod: "GET",
		ReqForm:       reqForm,
		ReqBodyJson:   nil,
	}

	// Call the GetStrToSign method and check the output
	strToSign, errCode, success := option.GetStrToSign(body)

	// Replace the expectedStrToSign with the expected string
	expectedStrToSign := "/v1/articles/15\nak=testAppKey"
	if !success {
		t.Errorf("Expected success but got false, error code: %s", errCode)
	}

	if strToSign != expectedStrToSign {
		t.Errorf("Expected strToSign: %s, got: %s", expectedStrToSign, strToSign)
	}
}

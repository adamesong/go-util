package signature_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/adamesong/go-util/redis"
	"github.com/adamesong/go-util/signature"
	"github.com/stretchr/testify/assert"
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

func TestVerifySign(t *testing.T) {

	redisClient := redis.RedisClient{Addr: "localhost:6379", Password: "", DB: 0}
	AppKeyAndSecret := map[string]string{
		"testAppKey-abcde1234": "testAppSecret-67890fghijk",
	}
	urlPath := "/v1/my_expert/"

	cases := []struct {
		reqMethod   string
		reqForm     url.Values
		reqBodyJson []byte
		uniqueSign  bool
	}{
		{
			reqMethod:   http.MethodGet,
			reqForm:     url.Values{},
			reqBodyJson: []byte{},
			uniqueSign:  false,
		},
		{
			reqMethod:   http.MethodPost,
			reqForm:     url.Values{},
			reqBodyJson: []byte{},
			uniqueSign:  true,
		},
	}

	doAssertion := assert.New(t)

	for n, tc := range cases {
		fmt.Println("Test case: ", n)

		// Prepare the SignOption instance
		sOption := signature.SignOption{
			AppKeyAndSecret: AppKeyAndSecret,
			UniqueSign:      tc.uniqueSign,
		}

		// signBody
		sBody := signature.SignBody{
			UrlPath:       urlPath,
			RequestMethod: tc.reqMethod,
			ReqForm:       tc.reqForm,
			ReqBodyJson:   tc.reqBodyJson,
		}

		// get test sign
		_, _, signedForm := sOption.GetTestSign(&sBody, "testAppKey-abcde1234")

		// Prepare the SignVerifyOption instance
		vOption := signature.SignVerifyOption{
			AppKeyAndSecret: AppKeyAndSecret,
			UniqueSign:      tc.uniqueSign,
			RedisClient:     &redisClient,
		}

		// signBody in signVerify
		vBody := signature.SignBody{
			UrlPath:       urlPath,
			RequestMethod: tc.reqMethod,
			ReqForm:       signedForm,
			ReqBodyJson:   tc.reqBodyJson,
		}

		// VerifySign
		result, errCode := vOption.VerifySign(&vBody)

		fmt.Println(errCode)
		doAssertion.True(result)
	}
}

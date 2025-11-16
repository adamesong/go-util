package signature_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/adamesong/go-util/redis"
	"github.com/adamesong/go-util/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// Correctly initialize the redis client using the new constructor
	redisClient, err := redis.NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err, "Failed to initialize redis client for test")
	defer redisClient.Close()

	AppKeyAndSecret := map[string]string{
		"testAppKey-abcde1234": "testAppSecret-67890fghijk",
	}
	urlPath := "/v1/my_expert/"

	cases := []struct {
		name        string
		reqMethod   string
		reqForm     url.Values
		reqBodyJson []byte
		uniqueSign  bool
	}{
		{
			name:        "Non-unique sign",
			reqMethod:   http.MethodGet,
			reqForm:     url.Values{},
			reqBodyJson: []byte{},
			uniqueSign:  false,
		},
		{
			name:        "Unique sign",
			reqMethod:   http.MethodPost,
			reqForm:     url.Values{},
			reqBodyJson: []byte{},
			uniqueSign:  true,
		},
	}

	doAssertion := assert.New(t)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
				RedisClient:     redisClient, // Correctly pass the redis client pointer
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

			doAssertion.True(result, "Expected verification to succeed, but it failed with code: %s", errCode)
			doAssertion.Empty(errCode, "Expected no error code on success")
		})
	}
}

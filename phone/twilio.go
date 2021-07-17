package phone

import (
	"encoding/json"
	"fmt"

	"net/http"
	"net/url"
	"strings"

	"github.com/adamesong/go-util/logging"
)

type Twilio struct {
	AccountSID  string
	AuthToken   string
	PhoneNumber string
}

// https://www.twilio.com/blog/2017/09/send-text-messages-golang.html
// SendSMS 通过twilio发一个手机短信
func (twilio *Twilio) SendSMSByTwilio(numberTo string, message string) {
	fmt.Println("to_number:::::", numberTo)
	// 判断手机号的有效性，如果手机号无效，不发送
	_, _, _, isValidNumber := ParsePhone(numberTo)
	fmt.Println("is valid number: ", isValidNumber)
	if !isValidNumber {
		return
	}
	//time.Sleep(10 * time.Second)
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + twilio.AccountSID + "/Messages.json"
	msgData := url.Values{} // created to store and encode the URL parameters
	msgData.Set("To", "+"+numberTo)
	msgData.Set("From", twilio.PhoneNumber)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode()) // created allows this object to be parsed like a string.

	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(twilio.AccountSID, twilio.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	if resp, err := client.Do(req); err != nil {
		logging.Error(err.Error())
	} else {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var data map[string]interface{}
			decoder := json.NewDecoder(resp.Body)
			err := decoder.Decode(&data)
			if err != nil {
				logging.Error(err.Error())
			} else {
				fmt.Println(data["sid"])
			}
		} else {
			var data map[string]interface{}
			decoder := json.NewDecoder(resp.Body)
			err := decoder.Decode(&data)
			if err != nil {
				logging.Error(err.Error())
			} else {
				logging.Info("Twilio Error: ", resp.Status, data)
			}
		}
	}
}

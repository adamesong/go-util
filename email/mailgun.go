package email

import (
	"context"
	"fmt"
	"time"

	"github.com/adamesong/go-util/logging"
	"github.com/adamesong/go-util/map_tool"

	"github.com/mailgun/mailgun-go/v4"
)

type Mailgun struct {
	Domain string // ie: mail.xxx.com
	APIKey string
}

// https://github.com/mailgun/mailgun-go/blob/master/examples/examples.go
func SendTaggedMessage(domain, apiKey string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		"Excited User <test@notice.xxx.com>",
		"Hello",
		"Testing some Mailgun awesomeness!",
		"xxxx@xxxx.ca",
	)

	err := m.AddTag("FooTag", "BarTag", "BlortTag")
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)

	if err != nil {
		fmt.Println("Mailgun error: ", err)
		return "", err
	} else {
		fmt.Println("Mailgun sent: ", id)
		return id, err
	}
}

func (mgObj *Mailgun) SendSimpleMessage(from, subject, text, template, to string) {
	mg := mailgun.NewMailgun(mgObj.Domain, mgObj.APIKey)
	m := mg.NewMessage(
		from,    // "Excited User <mailgun@notice.xxx.ca>",
		subject, // "Hello",
		text,    // "Testing some Mailgun awesomeness!",
		to,      // "xxx@xx.com",
	)

	if template != "" {
		m.SetTemplate(template)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	if err != nil {
		fmt.Println("Mailgun error: ", err)
		logging.Error(err.Error())
	} else {
		fmt.Println("Mailgun sent: ", id)
	}
}

// https://documentation.mailgun.com/en/latest/user_manual.html#batch-sending
// The maximum number of recipients allowed for Batch Sending is 1,000.
//
//	to的形式：{
//	  "user1@example.com" : {"var1": "ABC123456789", "var2": "adfsa"},
//	  "user2@example.com" : {"var1": "ZXY987654321", "var2": "34123"}
//	}
func (mgObj *Mailgun) SendBatchMessageLessThan1k(from, subject, text, template string, to map[string]interface{}) {
	mg := mailgun.NewMailgun(mgObj.Domain, mgObj.APIKey)
	m := mg.NewMessage(from, subject, text)
	// 如果template名称不为空，则设置模板
	if template != "" {
		m.SetTemplate(template)
	}
	for k, v := range to {
		// 如mailgun template里的variable是 {{var}} 则用m.AddTemplateVariable(var, value)来设置所有模板，所有的收件人的邮件内容都是一样的
		// 如mailgun template里的variable是 {%recipient.var%} 则用m.AddRecipientAndVariables(r, {"var":value})来设置，每个收件人的邮件内容都是独特的
		// if template != "" {
		// 	varMap := v.(map[string]interface{})
		// 	for varKey, varValue := range varMap {
		// 		addVariableErr := m.AddTemplateVariable(varKey, varValue)
		// 		if addVariableErr != nil {
		// 			fmt.Println("addVariableErr: ", addVariableErr.Error())
		// 		}
		// 	}
		// 	fmt.Println("varMap: ", varMap)
		// }

		addRecipientErr := m.AddRecipientAndVariables(k, v.(map[string]interface{}))
		// fmt.Println("k: ", k)
		if addRecipientErr != nil {
			fmt.Println("addRecipientErr: ", addRecipientErr.Error())
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	if err != nil {
		fmt.Println("Mailgun error: ", err)
		logging.Error(err.Error())
	} else {
		fmt.Println("Mailgun sent: ", id)
	}
}

// 当大于1000条的时候，自动分成若干个1000条的发出去。
func (mgObj *Mailgun) SendBatchEmail(from, subject, text, template string, to map[string]interface{}) {

	// 将map分为多个小雨1000的子map的array
	mapsArray, _ := map_tool.SplitMapTo(&to, 1000)

	// 对每个子map进行处理
	for _, subMaps := range mapsArray {
		mgObj.SendBatchMessageLessThan1k(from, subject, text, template, subMaps)
	}
}

func SendComplexMessage(domain, apiKey string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		"Excited User <test@notice.xxx.com>",
		"Hello",
		"Testing some Mailgun awesomeness!",
		"foo@example.com",
	)
	m.AddCC("baz@example.com")
	m.AddBCC("bar@example.com")
	m.SetHtml("<html>HTML version of the body</html>")
	m.AddAttachment("files/test.jpg")
	m.AddAttachment("files/test.txt")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	return id, err
}

func SendScheduledMessage(domain, apiKey string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		"Excited User <YOU@YOUR_DOMAIN_NAME>",
		"Hello",
		"Testing some Mailgun awesomeness!",
		"bar@example.com",
	)
	m.SetDeliveryTime(time.Now().Add(5 * time.Minute))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	return id, err
}

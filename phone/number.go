package phone

import (
	"fmt"
	"strconv"

	"github.com/nyaruka/phonenumbers"
)

// ParsePhone 将"1234567"这样的电话号码解析出国家代码等信息。mobile不能带"+"。
// countryCode: 1
// regionCode: CA, US等
// nationalNumber (778) 778-7878
func ParsePhone(mobile string) (countryCode, regionCode, nationalNumber string, isValidNumber bool) {
	// 这里从电话号码中取得区号。Parse()这个方法如果不加defaultRegion，需要将numberToParse的string中包含"+"
	if mobile != "" {
		var number *phonenumbers.PhoneNumber
		number, err := phonenumbers.Parse("+"+mobile, "")
		if err != nil {
			fmt.Println("err", err.Error())
		} else {
			countryCode = strconv.Itoa(int(*number.CountryCode))
			isValidNumber = phonenumbers.IsValidNumber(number)
			nationalNumber = phonenumbers.Format(number, phonenumbers.NATIONAL)
			regionCode = phonenumbers.GetRegionCodeForNumber(number)
		}
	}
	return
}

// Phone 带有国家码的电话信息
type Phone struct {
	CountryCode    string `json:"country_code"`    // 例如：1
	RegionCode     string `json:"region_code"`     // 例如：ca, us
	NationalNumber string `json:"national_number"` // 例如：(778) 778-7878
	IsValidNumber  bool   `json:"is_valid_number"` // 是否有效的电话号码：true, false
}

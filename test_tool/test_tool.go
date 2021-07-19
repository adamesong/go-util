package test_tool

import (
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GetSubTestName 用于在table driven tests中，给子测试命名，queryObj为每一个test cases中的结构体实例
// 注意：需要在struct中加tag `query_name:"xxx"`！！
func GetSubTestName(queryObj interface{}) (subtestName string) {
	s := reflect.TypeOf(queryObj)
	v := reflect.ValueOf(queryObj)
	for i := 0; i < s.NumField(); i++ { // s.NumField() 是这个结构体的Field的个数
		if v.Field(i).Type().String() == "string" {
			if v.Field(i).String() != "" {
				subtestName = subtestName + "_" + s.Field(i).Tag.Get("query_name") + "_" + v.Field(i).String()
			}
		} else if v.Field(i).Type().String() == "int" {
			subtestName = subtestName + "_" + s.Field(i).Tag.Get("query_name") + "_" + strconv.Itoa(int(v.Field(i).Int()))
		} else if v.Field(i).Type().String() == "*int" { // todo 需要修改
			subtestName = subtestName + "_" + s.Field(i).Tag.Get("query_name") + "_" + v.Field(i).String()
		} else if v.Field(i).Type().String() == "[]int" { // todo 需要修改
			subtestName = subtestName + "_" + s.Field(i).Tag.Get("query_name") + "_" + v.Field(i).String()
		} else if v.Field(i).Type().String() == "bool" {
			subtestName = subtestName + "_" + s.Field(i).Tag.Get("query_name") + "_" + strconv.FormatBool(v.Field(i).Bool())
		} else {
			subtestName = subtestName + "_" + s.Field(i).Tag.Get("query_name") + "_" + v.Field(i).String()
		}
	}
	// todo 如果是不含参数的，这里的命名需要改一下
	if subtestName == "" {
		subtestName = "_no_query_param"
	}
	return subtestName
}

// GetParams 用于提供一个query param struct(queryObj)，得到一个用于encode的params。
// 注意：需要在struct中加tag `query_name:"xxx"`！！
// 例如，提供struct{
//		Type string `query_name:"type"`
//		Featured string `query_name:"featured"`
//		Country string `query_name:"country"`
//		Sort string `query_name:"sort"`
//		PageSize string `query_name:"page_size"`
//		Page string `query_name:"page"`}
// 得到的params可以用于params.encode()，进而得到 type=0&country=CA&sort=1
func GetParams(queryObj interface{}) url.Values {
	// 将各query param encode
	params := url.Values{}
	s := reflect.TypeOf(queryObj)
	v := reflect.ValueOf(queryObj)
	for i := 0; i < s.NumField(); i++ {
		// 如果请求的该参数不为空，则加入到query params中（其实不做判断也可以，不影响结果，但这样更重于实际请求的情况）
		if v.Field(i).String() != "" && s.Field(i).Tag.Get("query_name") != "" {
			params.Add(s.Field(i).Tag.Get("query_name"), v.Field(i).String())
		}
	}
	return params
}

// CheckParamWithResult 判断query参数中（例如type=0&country=CA&sort=1）的参数值与返回的json结果中的值是否一致。
// 适用于：responseData中仅有1个实例，例如机构详细信息的返回值（不是机构列表信息）
// 注意：仅适用于返回的结果中的数据项类型是int/string/bool
// 注意：需要在struct中加tag `query_name:"xxx" result_name:"xxx" assert:"xxx"` 其中后两者可选填
// 如果返回值data是结构体，则result_name设置成结构体的field_name(大写)，如果返回data是map，则result_name是map的key_name(小写)
// assert tag如果不加，则默认值assert:"equal", 其他项目为notEmpty
// query_name是指在请求的参数中的参数名，result_name指在返回的json 结果中的参数名
func CheckParamWithResult(t *testing.T, queryObj interface{}, respData interface{}) {
	doAssert := assert.New(t)
	s := reflect.TypeOf(queryObj)  // 获得请求参数有哪些项
	v := reflect.ValueOf(queryObj) // 获得请求参数的这些项的值
	for i := 0; i < s.NumField(); i++ {
		if v.Field(i).String() != "" { // 如果没有请求参数，则不需要进行下列判断
			reqS := s.Field(i).Tag.Get("query_name")     // 该项的请求参数名
			reqV := v.Field(i).String()                  // 该项的请求参数值
			assertMethod := s.Field(i).Tag.Get("assert") // 该项预计值与实际值的判断方法，默认是equal，notEmpty等

			respS := s.Field(i).Tag.Get("result_name") // 该项的返回结果参数名。有可能没有result_name，则不需要判断此项
			if respS != "" {                           // 如果请求的参数没有对应的result_name，则不需要判断assert。Equal
				var respV reflect.Value
				// 如果接口的返回值resp中的Data是model中的一个结构体
				//fmt.Println(color.Red(reflect.TypeOf(respData).Kind().String()))
				if reflect.TypeOf(respData).Kind().String() == "struct" {
					respV = reflect.ValueOf(respData).FieldByName(respS) // 该项的返回结果的值，不一定是string
				} else { // 如果 == "map"，即如果接口的返回值resp中的Data是个map[string]interface{}
					key := reflect.ValueOf(respS)                   // 将map[string]interface{} 的key转成reflect.Value
					respV = reflect.ValueOf(respData).MapIndex(key) // 该项的返回结果的值，不一定是string
				}

				// 下面做判断
				// 判断是否一致
				if assertMethod == "" || assertMethod == "equal" {
					// 由于请求的query param只能是string，所以下面需要将结果中的实际值转换为string
					var respString string
					if reflect.TypeOf(respV.Interface()).String() == "int" { // 如果该项的返回结果的值的类型是int
						respString = strconv.Itoa(respV.Interface().(int)) // 则转换成string
					} else if reflect.TypeOf(respV.Interface()).String() == "string" { // 如果是string
						respString = respV.Interface().(string)
					} else if reflect.TypeOf(respV.Interface()).String() == "bool" { // 如果是bool
						respString = strconv.FormatBool(respV.Interface().(bool))
					} else if reflect.TypeOf(respV.Interface()).String() == "float64" { // 如果是float64 (json返回值中的数字)
						respString = strconv.Itoa(int(respV.Interface().(float64))) // 转换成int(对decimal有风险)
					} else if reflect.TypeOf(respV.Interface()).String() == "uint" { // 如果是uint
						respString = strconv.Itoa(int(respV.Interface().(uint))) // 转换成string
					} else {
						t.Error("类型是：" + reflect.TypeOf(respV.Interface()).String())
					}
					//fmt.Println("请求值：", reqV, " 实际值：", respString)
					// 执行判断：是否equal
					doAssert.Equal(reqV, respString, "过滤 "+reqS+" 时, "+reqV+" 与结果 "+respString+" 不符")

					//	判断该项不为空
				} else if assertMethod == "notEmpty" {
					doAssert.NotEmpty(respV, "过滤 "+reqS+" 时, "+respS+" 为空 ")
				}
			}
		}
	}
}

// CheckParamWithResults 判断query参数中（例如type=0&country=CA&sort=1）的参数值与返回的json结果中的值是否一致。
// 适用于：responseData中是实例的列表，例如机构列表信息的返回值（不是机构详细信息）
// 注意：仅适用于返回的结果中的数据项类型是int/string/bool
// 注意：需要在struct中加tag `query_name:"xxx" result_name:"xxx"`
// 如果返回值data是结构体，则result_name设置成结构体的field_name(大写)，如果返回data是map，则result_name是map的key_name(小写)
// query_name是指在请求的参数中的参数名，result_name指在返回的json 结果中的参数名
func CheckParamWithResults(t *testing.T, queryObj interface{}, respDataArray []interface{}) {
	for _, respData := range respDataArray {
		CheckParamWithResult(t, queryObj, respData)
	}
}

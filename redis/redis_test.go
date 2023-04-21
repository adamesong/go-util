package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var r = RedisClient{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
}

func TestConnectRedis(t *testing.T) {
	client := r.Connect()
	if pong, err := client.Ping(ctx).Result(); err != nil {
		t.Error(err.Error())
	} else {
		assert.Equal(t, "PONG", pong, "redis ping 获得的值"+pong+"不一致")
	}

}

func TestCloseConnect(t *testing.T) {
	client := r.Connect()
	defer closeConnect(client)
	if err := client.Close(); err != nil {
		t.Error(err.Error())
	}
	if pong, err := client.Ping(ctx).Result(); err != nil {
		fmt.Println(pong, err.Error())
	}
}

// TestSetAndGet 测试Set和Get
func TestSetAndGet(t *testing.T) {
	type someStruct struct {
		IntAttr    int
		StringAttr string
	}
	cases := []struct {
		Key       string
		Value     interface{}
		Duration  time.Duration
		ValueType string
		Exist     bool
	}{
		{"not_exists", "not exists value", 5 * time.Second, "string", false},
		{"1234", "数字key的字符串value", 5 * time.Second, "string", true},
		{"keyName", "字符串key的字符串value", 5 * time.Second, "string", true},
		{"array", []string{"a1", "a2"}, 5 * time.Second, "array", true},
		{"struct", someStruct{5, "abc"}, 5 * time.Second, "struct", true},
		{"struct array", []someStruct{{5, "a"}, {3, "b"}},
			30 * time.Second, "struct_array", true},
	}
	doAssertion := assert.New(t)
	for i, tc := range cases {
		t.Run(strconv.Itoa(i)+"_TestSetAndGet_"+tc.ValueType, func(t *testing.T) {
			valueJson, err := json.Marshal(tc.Value)
			if err != nil {
				t.Error(err.Error())
			}
			if tc.Exist == true {
				if err := r.Set(tc.Key, valueJson, tc.Duration); err != nil {
					t.Error(err.Error())
				}
			}
			if value, err := r.Get(tc.Key); err != nil {
				if tc.Exist == false { // 如果Get一个不存在的key，则返回错误redis: nil
					doAssertion.Error(errors.New("redis: nil"), err, "返回的错误类型不一致")
					//fmt.Println(err.Error())
				} else {
					t.Error(err.Error())
				}
			} else {
				if tc.ValueType == "string" {
					var valueGet string
					if err := json.Unmarshal(valueJson, &valueGet); err != nil {
						t.Error(err.Error())
					} else {
						//fmt.Println(valueGet)
						doAssertion.Equal(valueGet, tc.Value, "不一致")
					}

				} else if tc.ValueType == "array" {
					valueGet := new([]string)
					if err := json.Unmarshal(valueJson, valueGet); err != nil {
						t.Error(err.Error())
					} else {
						//fmt.Println(*valueGet)
						doAssertion.Equal(*valueGet, tc.Value, "不一致")
					}
				} else if tc.ValueType == "struct" {
					valueGet := new(someStruct)
					if err := json.Unmarshal(valueJson, valueGet); err != nil {
						t.Error(err.Error())
					} else {
						//fmt.Println(*valueGet)
						doAssertion.Equal(*valueGet, tc.Value, "不一致")
					}
				} else if tc.ValueType == "struct_array" {
					valueGet := new([]someStruct)
					if err := json.Unmarshal(valueJson, valueGet); err != nil {
						t.Error(err.Error())
					} else {
						//fmt.Println(*valueGet)
						doAssertion.Equal(*valueGet, tc.Value, "不一致")
					}
				}
				doAssertion.Equal(valueJson, value, "从redis Get的值与Set的值不一致")
			}
		})

	}
}

func TestDelete(t *testing.T) {
	type kAndV struct {
		Key   string
		Value string
	}
	cases := []struct {
		CaseName  string
		KAndV     []kAndV
		KeyAmount int
		Cached    bool // 是否真的Set到redis里
	}{
		{"key_not_exist", []kAndV{{"test_del_key_0", ""}}, 0, false},
		{"only_1_key_exists", []kAndV{{"test_del_key_1", "value 1"}}, 1,
			true},
		{"there_are_2_keys_exist", []kAndV{
			{"test_del_key_2", "value 2"},
			{"test_del_key_3", "value 3"},
		}, 2, true},
	}
	doAssertion := assert.New(t)
	for i, tc := range cases {
		t.Run(strconv.Itoa(i)+"_"+tc.CaseName, func(t *testing.T) {
			var keys []string
			// 先在缓存中set值
			if tc.Cached {
				for _, kv := range tc.KAndV {
					if err := r.Set(kv.Key, kv.Value, 5*time.Second); err != nil {
						t.Error(err.Error())
					}

				}
			}

			// 分别获得每个testCase中的key array
			for _, kv := range tc.KAndV {
				keys = append(keys, kv.Key)
			}

			if num, err := r.Delete(keys...); err != nil {
				t.Error(err.Error())
			} else {
				doAssertion.Equal(int64(tc.KeyAmount), num,
					"删除"+strconv.Itoa(tc.KeyAmount)+"个存在的key时测试失败")
			}
		})
	}
}

func TestSetNXAndSetXX(t *testing.T) {
	cases := []struct {
		Key      string
		Value    string
		Duration time.Duration
		Exist    bool
	}{
		{"1_key_exists", "some value", 5 * time.Second, true},
		{"1_key_not_exists", "value not exists", 5 * time.Second, false},
	}
	doAssertion := assert.New(t)
	for i, tc := range cases {
		t.Run(strconv.Itoa(i)+"_TestSetNXAndSetXX_"+tc.Key, func(t *testing.T) {
			if tc.Exist == true { // 如果存在，则应该SetNX返回失败，SetXX应成功
				if err := r.Set(tc.Key, tc.Value, tc.Duration); err != nil {
					t.Error(err.Error())
				}
				if result, err := r.SetNX(tc.Key, "another value", tc.Duration); err != nil {
					fmt.Println(err.Error())
				} else {
					doAssertion.Equal(false, result, "SetNX应不成功，但设置成功了")
				}
				if result, err := r.SetXX(tc.Key, "another value", tc.Duration); err != nil {
					t.Error(err.Error())
				} else {
					doAssertion.Equal(true, result, "SetXX应成功，但失败了")
				}
			} else { // 如果不存在，SetXX应返回失败，SetNX应成功
				if result, err := r.SetXX(tc.Key, "another value", tc.Duration); err != nil {
					fmt.Println(err.Error())
				} else {
					doAssertion.Equal(false, result, "SetXX应不成功，但设置成功了")
				}
				if result, err := r.SetNX(tc.Key, "another value", tc.Duration); err != nil {
					t.Error(err.Error())
				} else {
					doAssertion.Equal(true, result, "SetNX应成功，但失败了")
				}
			}
		})
	}
}

func TestGetSet(t *testing.T) {
	cases := []struct {
		Key      string
		Value    string
		Duration time.Duration
		Exist    bool
	}{
		{"2_key_exists", "some value", 5 * time.Second, true},
		{"2_key_not_exists", "value not exists", 5 * time.Second, false},
	}
	doAssertion := assert.New(t)
	for i, tc := range cases {
		t.Run(strconv.Itoa(i)+"_TestGetSet_"+tc.Key, func(t *testing.T) {
			if tc.Exist == true { // 如果存在，则GetSet可以获得老的value，并设置新的value
				if err := r.Set(tc.Key, tc.Value, tc.Duration); err != nil {
					t.Error(err.Error())
				}
				if oldValue, err := r.GetSet(tc.Key, "another value"); err != nil {
					t.Error(err.Error())
				} else {
					doAssertion.Equal(tc.Value, string(oldValue), "老数据与GetSet获得的老数据不一致")
					if newValue, err := r.Get(tc.Key); err != nil {
						t.Error(err.Error())
					} else {
						doAssertion.Equal("another value", string(newValue),
							"新设置的数据与实际新设置的值不一致")
					}
				}

			} else { // 如果本身不存在这个key，则GetSet获取的老数据为空，同时设置新数据，且无过期时间
				if oldValue, err := r.GetSet(tc.Key, "another value"); err != nil {
					fmt.Println(err.Error()) // redis: nil
				} else {

					fmt.Println("老数据:", string(oldValue))
					if newValue, err := r.Get(tc.Key); err != nil {
						t.Error(err.Error())

					} else {
						doAssertion.Equal("another value", string(newValue),
							"新设置的数据与实际新设置的值不一致")
					}
				}
			}
			// 清理测试数据，因为不存在老key时，GetSet设置新值无过期时间
			if _, err := r.Delete(tc.Key); err != nil {
				t.Error(err.Error())
			}
		})

	}
}

func TestTTL(t *testing.T) {
	cases := []struct {
		Key      string
		Value    string
		Duration time.Duration
		Exist    bool
	}{
		{"valid", "valid value", 10 * time.Second, true},
		{"expired", "expired value", 1 * time.Second, true},
		{"key_not_exists", "value not exists", 10 * time.Second, false},
		{"no_expiration", "no expiration value", 0 * time.Second, true}, // 0意味着无有效期
	}
	doAssertion := assert.New(t)
	for i, tc := range cases {
		t.Run(strconv.Itoa(i)+"_TestTTL_"+tc.Key, func(t *testing.T) {
			if tc.Exist == true { // 如果存在，设置value
				if err := r.Set(tc.Key, tc.Value, tc.Duration); err != nil {
					t.Error(err.Error())
				}
			}
			// 获得过期时间
			time.Sleep(2 * time.Second) // 等2秒，此时10秒的key还没过期，1秒的key过期了

			if dur, err := r.TTL(tc.Key); err != nil {
				t.Error(err.Error())
			} else {
				//fmt.Println(tc.Key, dur)

				switch tc.Key {
				case "valid":
					fmt.Println("duration:", dur)
					doAssertion.Greater(int(dur/time.Second), 0, "剩余时间应大于0")
				case "expired":
					doAssertion.Equal(-2*time.Nanosecond, dur, "过期Key不存在时返回值不正确")
				case "key_not_exists":
					doAssertion.Equal(-2*time.Nanosecond, dur, "不存在的key返回值不正确")
				case "no_expiration":
					doAssertion.Equal(-1*time.Nanosecond, dur, "无有效期(永远存在)时的返回值不正确")
				}
			}
			// 清理数据
			if _, err := r.Delete(tc.Key); err != nil {
				t.Error(err.Error())
			}
		})
	}

}

func TestExpire(t *testing.T) {
	cases := []struct {
		Key      string
		Value    string
		Duration time.Duration
		Exist    bool
	}{
		{"test_expire", "value", 5 * time.Second, true},
		{"no_expiration", "no expiration value", 0 * time.Second, true}, // 无有效期
		{"not_exists", "not exists", 5 * time.Second, false},
	}
	doAssertion := assert.New(t)

	for i, tc := range cases {
		t.Run(strconv.Itoa(i)+"_TestExpire_"+tc.Key, func(t *testing.T) {
			// 初始设置5秒的过期时间
			if tc.Exist == true {
				if err := r.Set(tc.Key, tc.Value, tc.Duration); err != nil {
					t.Error(err.Error())
				}
			}

			// 重设timeout时间（比原有duration增加5秒，无过期时间的设为5秒）
			if result, err := r.Expire(tc.Key, tc.Duration+5*time.Second); err != nil {
				t.Error(err.Error())
			} else {
				if tc.Key == "not_exists" { // 如果不存在，用expire设置时间会返回false
					doAssertion.Equal(false, result)
				} else {
					// 此时获得过期时间应该在0至5秒之间
					if dur, err := r.TTL(tc.Key); err != nil {
						t.Error(err.Error())
					} else { // 获得的新过期时间应该比原来的过期时间更大
						doAssertion.Greater(int(dur/time.Second), int(tc.Duration/time.Second))
					}
				}
			}

			// 清理数据
			if _, err := r.Delete(tc.Key); err != nil {
				t.Error(err.Error())
			}
		})
	}
}

func TestLikeDeletes(t *testing.T) {
	cases := []struct {
		Key      string
		Value    string
		Duration time.Duration
	}{
		{"param", "value", 5 * time.Second},
		{"paRAm", "value", 5 * time.Second},                             // 大写
		{"param_abc", "value", 5 * time.Second},                         // 后面有关键词
		{"abc_param", "value", 5 * time.Second},                         // 前面有关键词
		{"123_param_456", "value", 5 * time.Second},                     // 前后都有关键词
		{"param_no_expiration", "no expiration value", 0 * time.Second}, // 无有效期
	}
	// 创建6个
	amount := 0
	for _, tc := range cases {
		if err := r.Set(tc.Key, tc.Value, tc.Duration); err != nil {
			t.Error(err.Error())
		}
		amount += 1
	}
	// like删除，验证删除数量是不是6个
	if num, err := r.LikeDeletes("param"); err != nil {
		assert.Equal(t, amount, int(num), "删除数量与创建数量不一致")
	}

	for _, tc := range cases {
		if value, err := r.Get(tc.Key); err != nil {
			assert.Equal(t, "", string(value))
		}
	}
}

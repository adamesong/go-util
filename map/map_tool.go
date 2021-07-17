package map

import (
	"reflect"
)

// DeepCopy 深度拷贝一个map
func DeepCopyMap(originalMap map[string]interface{}) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})
	for k, v := range originalMap {
		newMap[k] = v
	}
	return
}

// 将结构体转换为map
// obj为某个结构体的实例
// tag为结构体里某项的tag，例如`json:"id"`的"json"
// 注意：如果用于存为redis的hashSet，嵌套的struct转为map可能有问题
// 注意：如果存到redis，time.Time也有问题
func StructToMap(obj interface{}, tag string) map[string]interface{} {
	t := reflect.TypeOf(obj)  // 如：model.Article
	v := reflect.ValueOf(obj) // 值
	maps := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		if tag != "" {
			maps[t.Field(i).Tag.Get(tag)] = v.Field(i).Interface()
		} else {
			maps[t.Field(i).Name] = v.Field(i).Interface()
		}
	}
	return maps
}

// 将一个map的key:value分成最大数量length个的若干个map 组成的array，分成了groups个
// 当maps为空map时，groups = 0
func SplitMapTo(maps *map[string]interface{}, length int) (splitMapArray []map[string]interface{}, groups int) {
	var newMap map[string]interface{}
	i := 0
	for k, v := range *maps {
		// 每length个元素，就新建一个子map，从第0个元素开始
		if i%length == 0 {
			newMap = make(map[string]interface{})
		}
		newMap[k] = v

		// 每length个元素，就将这个子map append到splitMapArray里
		// 最后一组不满length个，也append到里面
		if (i-length+1)%length == 0 || len(*maps) == i+1 {
			splitMapArray = append(splitMapArray, newMap)
			groups += 1
		}

		i += 1
	}
	return
}

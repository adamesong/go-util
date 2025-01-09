package struct_tool

import (
	"reflect"
	"strconv"
)

// SetUpdateValue 将updateInfoObj这个结构体实例中的各项更新至originObj中各同名项中。
// 1. updateInfoObj 中的各项如果不是零值，且与originObj中的对应项值不同，才会更新至originObj
// 2. originObjPtr 是 originObj的指针；updateInfoObjPtr 是 updateInfoObj的指针
// updateInfoObj中的各项如果不是zero value，则将originObj中同名的项的值更新
// ! 注意：如果updateInfoObj中有些项刻意置为零值，在这里不会改变，需要额外单独赋值给originObj相应的项。例如，将state置为0，或将某字段置空“”
// changed: 表示原始值与info中有不同的值
// originMap: 如果changed == true，orginMap中列出变动前的项和原始值
// changedMap: 如果changed == true，changedMap中列出后的项和值
// deprecated, 建议使用SetUpdateValueWithOption
func SetUpdateValue(originObjPtr, updateInfoObjPtr interface{}) (changed bool, originMap, changedMap map[string]interface{}) {
	originStruct := reflect.TypeOf(originObjPtr)
	originValue := reflect.ValueOf(originObjPtr)
	originMap = map[string]interface{}{}

	infoStruct := reflect.TypeOf(updateInfoObjPtr)
	infoValue := reflect.ValueOf(updateInfoObjPtr)
	changedMap = map[string]interface{}{}

	for i := 0; i < infoStruct.Elem().NumField(); i++ { // NumField() 这个结构体的Field的个数；因为传的是指针，所以这里要用Elem()来获取指针指向的元素。https://learnku.com/articles/51004

		// 如果infoValue中的某项不是零值（如果是零值，则跳过，进入下一个）
		if !infoValue.Elem().Field(i).IsZero() {
			// 取出这项的Name
			fieldName := infoStruct.Elem().Field(i).Name
			// 从orginalStruct中找到同名的项
			originField, exists := originStruct.Elem().FieldByName(fieldName)
			if !exists {
				continue // 如果 originObj 中不存在该字段，跳过
			}
			// 如果info中的值与origin的值不同，则将该项的值设为info中的值 (注意：reflect.Value无法用==来判断是否相同，需要用 .Interface())
			if !reflect.DeepEqual(originValue.Elem().FieldByName(originField.Name).Interface(), infoValue.Elem().Field(i).Interface()) {

				// 记录下变动项的原始值
				originMap[fieldName] = originValue.Elem().FieldByName(originField.Name).Interface()
				// 记录下变动之后的值
				changedMap[fieldName] = infoValue.Elem().Field(i).Interface()

				// fmt.Println("与原址不同")
				originValue.Elem().FieldByName(originField.Name).Set(infoValue.Elem().Field(i))

				changed = true // 表示有内容变了
			}
		}
	}
	// 此时设置完之后，originObjPtr 指针所指向的 originObj的值已经被更新
	return
}

type SetUpdateValueOption struct {
	// 是否将零值更新到originObj中,零值也包含空指针nil
	CanSetZero bool
}

// SetUpdateValue2 将updateInfoObj这个结构体实例中的各项更新至originObj中各同名项中。
// 1. originObjPtr 是 originObj的指针；updateInfoObjPtr 是 updateInfoObj的指针
// 2. 如果 canSetZero为false, updateInfoObj的各项如果不是零值，且与originObj中的对应项值不同，才会更新至originObj；如果canSetZero为true，updateInfoObj的各项如果是零值，则将orginObj中的对应项更新为零值

// changed: 表示原始值被更新
// originMap: 如果changed == true，orginMap中列出变动前的项和原始值
// changedMap: 如果changed == true，changedMap中列出后的项和值
func SetUpdateValueWithOption(originObjPtr, updateInfoObjPtr interface{}, option *SetUpdateValueOption) (changed bool, originMap, changedMap map[string]interface{}) {
	// option不可为nil，必须有option
	if option == nil {
		panic("option cannot be nil")
	}

	// 初始化返回值
	originMap = map[string]interface{}{}
	changedMap = map[string]interface{}{}

	// 获取 reflect 的类型和值
	originValue := reflect.ValueOf(originObjPtr)
	updateValue := reflect.ValueOf(updateInfoObjPtr)

	// 必须是指针类型
	if originValue.Kind() != reflect.Ptr || updateValue.Kind() != reflect.Ptr {
		panic("both originObjPtr and updateInfoObjPtr must be pointers to structs")
	}

	// 获取指向的元素
	originValue = originValue.Elem()
	updateValue = updateValue.Elem()

	// 确保是结构体
	if originValue.Kind() != reflect.Struct || updateValue.Kind() != reflect.Struct {
		panic("both originObjPtr and updateInfoObjPtr must be pointers to structs")
	}

	// 遍历 updateInfoObj 的字段
	updateType := updateValue.Type()
	for i := 0; i < updateType.NumField(); i++ {
		field := updateType.Field(i)
		fieldName := field.Name

		// 获取 update 和 origin 的字段值
		updateFieldValue := updateValue.Field(i)
		originFieldValue := originValue.FieldByName(fieldName)

		// 检查 origin 中是否存在该字段, 如果orgin中不存在update的字段，跳过该字段
		if !originFieldValue.IsValid() {
			continue
		}

		// 处理指针字段和非指针字段
		var updateValueInterface interface{}
		if updateFieldValue.Kind() == reflect.Ptr {
			// 指针字段，获取指针指向的值
			if updateFieldValue.IsNil() {
				updateValueInterface = nil
			} else {
				updateValueInterface = updateFieldValue.Elem().Interface()
			}
		} else {
			// 非指针字段，直接获取值
			updateValueInterface = updateFieldValue.Interface()
		}

		// 判断是否更新字段
		shouldUpdate := false
		if option.CanSetZero {
			// 如果允许设置零值，则直接更新字段
			shouldUpdate = true
		} else {
			// 如果不允许设置零值，则需要判断是否为零值
			if !updateFieldValue.IsZero() {
				shouldUpdate = true
			}
		}

		// 处理 nil 的特殊情况 (字段为指针，允许设置零值的情况下，updateFieldValue是nil，则将原值也设为nil)
		if shouldUpdate && updateFieldValue.Kind() == reflect.Ptr && updateFieldValue.IsNil() && option.CanSetZero {
			// 如果允许设置零值，且字段为 nil，则更新
			originMap[fieldName] = originFieldValue.Interface()
			changedMap[fieldName] = nil
			originFieldValue.Set(reflect.Zero(originFieldValue.Type())) // 将字段设置为 nil
			changed = true
			continue
		}

		// 如果字段需要更新，并且值与原始值不同
		if shouldUpdate && !reflect.DeepEqual(originFieldValue.Interface(), updateValueInterface) {
			// 记录变更前的值
			originMap[fieldName] = originFieldValue.Interface()
			// 记录变更后的值
			changedMap[fieldName] = updateValueInterface

			// 更新字段值
			if originFieldValue.Kind() == reflect.Ptr {
				// 更新指针字段
				if updateValueInterface == nil {
					originFieldValue.Set(reflect.Zero(originFieldValue.Type()))
				} else {
					newValue := reflect.New(originFieldValue.Type().Elem())
					newValue.Elem().Set(reflect.ValueOf(updateValueInterface))
					originFieldValue.Set(newValue)
				}
			} else {
				// 更新非指针字段
				originFieldValue.Set(reflect.ValueOf(updateValueInterface))
			}

			// 标记有变更
			changed = true
		}
	}

	return
}

// GetStringFromObj obj为结构体实例
// 与test_tool中的GetSubTestName类似
// 返回值格式：keyName_keyValueString_keyName_keyValueString
func GetStringFromObj(obj interface{}) (keyString string) {
	s := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	first := 0
loop:
	for i := 0; i < s.NumField(); i++ { // s.NumField() 是这个结构体的Field的个数
		// 如果某项是零值，则跳过此项
		if v.Field(i).IsZero() {
			continue loop
		} else {
			// 否则，记录非0值的个数
			first += 1
		}
		// 如果不是第一个元素，则在前面加"_"
		if first != 1 {
			keyString += "_"
		}
		// 获得struct的项的名称和类型
		k := s.Field(i).Name
		vType := v.Field(i).Type().String()
		// 处理struct的项的值，处理成string
		if vType == "string" {
			keyString += k + "_" + v.Field(i).String()
		} else if vType == "int" {
			keyString += k + "_" + strconv.Itoa(int(v.Field(i).Int()))
		} else if vType == "uint" {
			keyString += k + "_" + strconv.Itoa(int(v.Field(i).Uint()))
		} else if vType == "bool" {
			keyString += k + "_" + strconv.FormatBool(v.Field(i).Bool())
		} else if vType == "[]string" {
			for j, sub := range v.Field(i).Interface().([]string) {
				if j != 0 {
					keyString += "_" + sub
				} else {
					keyString += k + "_" + sub
				}
			}
		} else if vType == "[]int" {
			for j, sub := range v.Field(i).Interface().([]int) {
				if j != 0 {
					keyString += "_" + strconv.Itoa(sub)
				} else {
					keyString += k + "_" + strconv.Itoa(sub)
				}
			}
		} else if vType == "[]uint" {
			for j, sub := range v.Field(i).Interface().([]uint) {
				if j != 0 {
					keyString += "_" + strconv.Itoa(int(sub))
				} else {
					keyString += k + "_" + strconv.Itoa(int(sub))
				}
			}
		}
	}
	return
}

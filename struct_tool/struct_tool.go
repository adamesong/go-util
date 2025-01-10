package struct_tool

import (
	"errors"
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
	// 是否将空指针(nil)更新到originObj中
	CanSetNil bool
}

// SetUpdateValueWithOption 将updateInfoObj的值更新到originObj中。
// - originObjPtr: 原始对象的指针。
// - updateInfoObjPtr: 更新信息的对象指针。
// - option: 更新选项。
// 返回值:
// - changed: 是否有任何字段被更新。
// - originMap: 记录被更新字段的原始值。
// - changedMap: 记录被更新字段的更新值。
// originObj 和 updateInfoObj 可能是两个不同的struct的instance。这两个struct的field名称可能相同，只有updateInfoObj中的field名称在originObj中存在时，才将会被更新。
// 允许处理字段类型不一致（值与指针互转）的场景，并根据 CanSetNil 决定是否更新空指针。
// 场景1: 例如originObj中可能有个Age int, updateInfoObj中可能有Age *int，那么如果后者不是nil，则将后者指针指向的值更新到前者。如果后者是nil，则无值，无论是否设置CanSetNil，都不会将前者设为零值。
// 场景2: 如果originObj中Age *int， updateInfoObj中Age int，则将后者的值更新至前者。
// 场景3: 如果orginObj中Age *int, updateInfoObj中 Age *int，则后者的Age指针根据CanSetNil的设置情况决定是否更新到originObj的Age
// 如果CanSetNil为false，updateInfoObj中的各项目的值除了nil以外都将更新至originObj，nil以外的零值也更新；
// 如果CanSetNil为true，updateInfoObj中的各项目的值包括nil和其他类型的零值都将更新至originObj
// ! 所以，将此function用于go-gin bind query或bind json时，将updateInfo struct中的可能会出现零值的field用指针表示，并设置CanSetNil=false，这样能防止未提供的更新项被误更新为零值
func SetUpdateValueWithOption(originObjPtr, updateInfoObjPtr interface{}, option *SetUpdateValueOption) (changed bool, originMap, changedMap map[string]interface{}, err error) {
	// option不可为nil，必须有option
	if option == nil {
		// panic("option cannot be nil")
		err = errors.New("option cannot be nil")
		return
	}

	// 初始化返回值
	originMap = make(map[string]interface{})
	changedMap = make(map[string]interface{})

	// 获取 reflect 的类型和值
	originValue := reflect.ValueOf(originObjPtr)
	updateValue := reflect.ValueOf(updateInfoObjPtr)

	// 必须是指针类型
	if originValue.Kind() != reflect.Ptr || updateValue.Kind() != reflect.Ptr {
		// panic("both originObjPtr and updateInfoObjPtr must be pointers to structs")
		err = errors.New("both originObjPtr and updateInfoObjPtr must be pointers to structs")
		return
	}

	// 获取指向的元素
	originValue = originValue.Elem()
	updateValue = updateValue.Elem()

	// 确保是结构体
	if originValue.Kind() != reflect.Struct || updateValue.Kind() != reflect.Struct {
		// panic("both originObjPtr and updateInfoObjPtr must be pointers to structs")
		err = errors.New("both originObjPtr and updateInfoObjPtr must be pointers to structs")
		return
	}

	// 遍历 updateInfoObj 的字段
	updateType := updateValue.Type()
	for i := 0; i < updateType.NumField(); i++ {
		field := updateType.Field(i)
		fieldName := field.Name

		// 获取 update 和 origin 的字段值
		updateFieldValue := updateValue.Field(i)
		originFieldValue := originValue.FieldByName(fieldName)

		// 如果 origin 中不存在该字段，跳过
		if !originFieldValue.IsValid() {
			continue
		}

		// 判断类型是否一致
		isOriginPtr := originFieldValue.Kind() == reflect.Ptr
		isUpdatePtr := updateFieldValue.Kind() == reflect.Ptr

		// 处理两边都是指针的情况
		if isOriginPtr && isUpdatePtr {
			if updateFieldValue.IsNil() {
				if option.CanSetNil {
					originMap[fieldName] = originFieldValue.Interface()
					changedMap[fieldName] = nil
					originFieldValue.Set(reflect.Zero(originFieldValue.Type()))
					changed = true
				}
			} else {
				originMap[fieldName] = originFieldValue.Interface()
				changedMap[fieldName] = updateFieldValue.Interface()
				originFieldValue.Set(updateFieldValue)
				changed = true
			}
			continue
		}

		// 处理origin是值，update是指针
		if !isOriginPtr && isUpdatePtr {
			if !updateFieldValue.IsNil() {
				updateValue := updateFieldValue.Elem().Interface()
				originMap[fieldName] = originFieldValue.Interface()
				changedMap[fieldName] = updateValue
				originFieldValue.Set(reflect.ValueOf(updateValue))
				changed = true
			}
			continue
		}

		// 处理origin是指针，update是值
		if isOriginPtr && !isUpdatePtr {
			updateValue := updateFieldValue.Interface()
			originMap[fieldName] = originFieldValue.Interface()
			changedMap[fieldName] = updateValue
			newValue := reflect.New(originFieldValue.Type().Elem())
			newValue.Elem().Set(reflect.ValueOf(updateValue))
			originFieldValue.Set(newValue)
			changed = true
			continue
		}

		// 处理两边都是值的情况
		updateValue := updateFieldValue.Interface()
		if !reflect.DeepEqual(originFieldValue.Interface(), updateValue) {
			originMap[fieldName] = originFieldValue.Interface()
			changedMap[fieldName] = updateValue
			originFieldValue.Set(reflect.ValueOf(updateValue))
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

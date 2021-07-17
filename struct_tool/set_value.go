package struct_tool

import (
	"reflect"
)

// SetUpdateValue 将updateInfoObj这个结构体实例中的各项更新至originObj中各同名项中。
// 1. updateInfoObj 中的各项如果不是零值，且与originObj中的对应项值不同，才会更新至originObj
// 2. originObjPtr 是 originObj的指针；updateInfoObjPtr 是 updateInfoObj的指针
// updateInfoObj中的各项如果不是zero value，则将originObj中同名的项的值更新
// ! 注意：如果updateInfoObj中有些项刻意置为零值，在这里不会改变，需要额外单独赋值给originObj相应的项。例如，将state置为0，或将某字段置空“”
// changed: 表示原始值与info中有不同的值
// originMap: 如果changed == true，orginMap中列出变动前的项和原始值
// changedMap: 如果changed == true，changedMap中列出后的项和值
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
			originField, _ := originStruct.Elem().FieldByName(fieldName)
			// 如果info中的值与origin的值不同，则将该项的值设为info中的值 (注意：reflect.Value无法用==来判断是否相同，需要用 .Interface())
			if originValue.Elem().FieldByName(originField.Name).Interface() != infoValue.Elem().Field(i).Interface() {

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

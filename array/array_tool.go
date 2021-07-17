package array

import "strconv"

func StringsContains(array []string, val string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}

func IntContains(array []int, val int) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}

func IntIn(val int, array []int) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

func UintIn(val uint, array []uint) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

func IsEmptyStringsArray(array []string) bool {
	if array == nil {
		return true
	}

	if len(array) == 0 {
		return true
	}
	for i := 0; i < len(array); i++ {
		if array[i] != "" {
			return false
		}
	}
	return true
}

// 将字符串list转换为int list，如果字符串list中的某元素不可转换为int，则 int list中不会有此项。
func StringArrayToIntArray(strArray []string) (intArray []int) {
	intArray = make([]int, 0)
	for _, str := range strArray {
		converted, err := strconv.Atoi(str)
		if err == nil {
			intArray = append(intArray, converted)
		}
	}
	return intArray
}

// 给[]int 去重。
// 该函数总共初始化两个变量，一个长度为0的slice，一个空map。由于slice传参是按引用传递，没有创建额外的变量。
// 只是用了一个for循环，代码更简洁易懂。
// 利用了map的多返回值特性。
// 空struct不占内存空间，可谓巧妙。
func RemoveDuplicateElementInt(sourceArray []int) []int {
	result := make([]int, 0, len(sourceArray))
	temp := map[int]struct{}{}
	for _, item := range sourceArray {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 给[]string去重，原理同上
func RemoveDuplicateElementString(sourceArray []string) []string {
	result := make([]string, 0, len(sourceArray))
	temp := map[string]struct{}{}
	for _, item := range sourceArray {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 给[]uin去重，原理同上
func RemoveDuplicateElementUint(sourceArray []uint) []uint {
	result := make([]uint, 0, len(sourceArray))
	temp := map[uint]struct{}{}
	for _, item := range sourceArray {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 删除Array中的某个uint（如果该array中有多个相同的元素，每次只能删除1个）
func RemoveElementUint(sourceArray *[]uint, toBeRemoved uint) (removed bool) {
	// 如果列表为空，直接返回
	if len(*sourceArray) == 0 {
		return
	}

	var index int
	for i, v := range *sourceArray {
		if v == toBeRemoved {
			index = i
			removed = true
		}
	}
	*sourceArray = append((*sourceArray)[:index], (*sourceArray)[index+1:]...)
	return
}

// 删除Array中的某个int（如果该array中有多个相同的元素，每次只能删除1个）
func RemoveElementInt(sourceArray *[]int, toBeRemoved int) (removed bool) {
	// 如果列表为空，直接返回
	if len(*sourceArray) == 0 {
		return
	}

	var index int
	for i, v := range *sourceArray {
		if v == toBeRemoved {
			index = i
			removed = true
		}
	}
	*sourceArray = append((*sourceArray)[:index], (*sourceArray)[index+1:]...)
	return
}

// 如果sourceArray中有 toBeRemoved array中的元素，则从sourceArray中删除，最终返回新的array
// https://stackoverflow.com/questions/5020958/go-what-is-the-fastest-cleanest-way-to-remove-multiple-entries-from-a-slice/5022696
func RemoveElementsInt(sourceArray *[]int, toBeRemoved *[]int) (newArray []int) {
	//newArray = make([]int, 0)
	i := 0
	if len(*sourceArray) == 0 {
		return
	}
loop:
	for _, sourceElement := range *sourceArray {
		for _, toDeleteElement := range *toBeRemoved {
			if sourceElement == toDeleteElement {
				continue loop
			}
		}
		newArray = append(newArray, sourceElement)
		i++
	}
	return newArray
}

// 将一个string array 分成最大数量 length个的若干个string array，分成了groups组
// 当stringArray为空array时，groups == 0
func SplitStringArrayTo(stringArray *[]string, length int) (splitArray [][]string, groups int) {
	var newArray []string
	arrayLength := len(*stringArray)
	for i, v := range *stringArray {
		// 每length个元素，就新建一个子array，从第0个元素开始
		if i%length == 0 {
			newArray = *new([]string)
		}
		newArray = append(newArray, v)

		// 每length个元素，就append到splitArray里，从第length个元素开始
		// 最后一组不满length个，也append到splitArray里
		if (i-length+1)%length == 0 || arrayLength == i+1 {
			splitArray = append(splitArray, newArray)
			groups += 1
		}
	}
	return
}

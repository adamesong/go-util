package pagination

import (
	"strconv"
)

// GetOffset returns the number of records to skip before starting to return the records
// http://gorm.io/docs/query.html#Offset
// use GetOffset if parameter type is int64
func GetOffsetAndPageNum(pageNum, pageSize string, totalAmount int) (offSet, realPageNum, pageSizeInt int) {

	// convert string to int
	var page int
	if pageNum == "" {
		page = 1
	} else {
		page, _ = strconv.Atoi(pageNum)
	}
	pageSizeInt = 25    // 默认的每页数量
	if pageSize != "" { // 如果有自定义的每页数量，则用自定义的
		pageSizeInt, _ = strconv.Atoi(pageSize)
	}

	if page > 0 { // 如果页码大于0

		offSet = (page - 1) * pageSizeInt

		if offSet >= totalAmount { // 如果页码超出最大范围，则offset设为最后一页之前的数字
			lastPageSize := totalAmount % pageSizeInt
			if lastPageSize == 0 { // 如果最后一页是满页
				offSet = totalAmount - pageSizeInt
			} else {
				offSet = totalAmount - lastPageSize
			}

			realPageNum = offSet/pageSizeInt + 1
		} else { // 如果页码没超出最大范围
			realPageNum = page
		}
	} else { // 如果页码小于0，则返回第一页
		offSet = 0
		realPageNum = 1
	}
	return
}

func GetOffset(pageNum, pageSize, totalAmount int64) (offSet, realPageNum int64) {

	if pageNum > 0 { // 如果页码大于0

		offSet = (pageNum - 1) * pageSize

		if offSet >= totalAmount { // 如果页码超出最大范围，则offset设为最后一页之前的数字
			lastPageSize := totalAmount % pageSize
			if lastPageSize == 0 { // 如果最后一页是满页
				offSet = totalAmount - pageSize
			} else {
				offSet = totalAmount - lastPageSize
			}
			realPageNum = offSet/pageSize + 1
		} else { // 如果页码没超出最大范围
			realPageNum = pageNum
		}

	} else { // 如果页码小于0，则返回第一页
		offSet = 0
		realPageNum = 1
	}
	return
}

// 获得最后一页的最后一个是第几个
func GetTheLastNumber(pageNum, pageSize string, totalAmount int) (lastNum int) {
	offSet, _, pageSizeInt := GetOffsetAndPageNum(pageNum, pageSize, totalAmount)
	lastNum = offSet + pageSizeInt
	return
}

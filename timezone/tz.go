package timezone

import (
	"time"
)

// 将某个时间t 的字面时间设置为某时区的时间。适用于前端传入UTC时间后，在后端设置为用户自己时区的时间。
// ! 注意：此过程是强行设置，而非转换。字面上时间不变化，仅时区发生了变化。
// t: 时间，可能含任意时区信息。本过程将忽略t的时区，强制将t设置为时区tz。
// tz: 合法的时区字符串，如tz == ""，则设置为UTC时区
// 例如，t 为 2021-04-28 00:00:00 -0700 PDT  tz为：Asia/Shanghai，则 timeInTZ 为2021-04-28 00:00:00 +0800 PDT
func TimeSetTimezone(t time.Time, tz string) (timeInTZ time.Time, err error) {

	// fmt.Println("收到的时间：", t)
	// fmt.Println("收到的时区：", tz)

	// 判断tz是否为合法的timezone string
	loc, tzErr := time.LoadLocation(tz)
	if tzErr != nil {
		// fmt.Println("转换错误：", tzErr.Error())
		return time.Time{}, tzErr
	}

	timeInTZ = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	// fmt.Println("转化完的时间：", timeInTZ)
	return
}

// 将时间t转换为某时区的时间。
// ! 注意：此过程是强行设置，而非转换。如转换前后的时区不同，则时间字面上也发生变化。
func TimeConvertTimeZone(t time.Time, tz string) (timeInTZ time.Time, err error) {
	// 判断tz是否为合法的timezone string
	loc, tzErr := time.LoadLocation(tz)
	if tzErr != nil {
		return time.Time{}, tzErr
	}
	timeInTZ = t.In(loc)
	return
}

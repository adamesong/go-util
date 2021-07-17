package logging

// https://book.eddycjy.com/golang/gin/log.html
// https://www.flysnow.org/2017/05/06/go-in-action-go-log.html

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	DBLogSavePath = "runtime/db_logs/" // log文件所在的目录
	DBLogSaveName = "db_log"           // log文件名
	DBLogFileExt  = "log"              // log文件后缀
	DBTimeFormat  = "20060102"         // 时间格式（日期格式）
)

// log文件所在的目录
func getDBLogFilePath() string {
	// return fmt.Sprintf("%s", LogSavePath)
	return DBLogSavePath
}

// log文件的目录+文件名
func GetDBLogFileFullPath() string {
	prefixPath := getDBLogFilePath()
	suffixPath := fmt.Sprintf("%s%s.%s", DBLogSaveName, time.Now().Format(DBTimeFormat), DBLogFileExt)

	return fmt.Sprintf("%s%s", prefixPath, suffixPath)
}

func OpenDBLogFile(filePath string) *os.File {
	_, err := os.Stat(filePath) // 返回文件信息结构描述文件。如果出现错误，会返回*PathError
	switch {
	case os.IsNotExist(err):
		mkDBLogDir()
	case os.IsPermission(err):
		//log.Fatalf("Permission :%v", err)
		log.Printf("Permission :%v", err)
	}
	// 644 只有拥有者有读写权限；而属组用户和其他用户只有读权限。
	// 666 所有用户都有文件读、写权限。
	handle, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		//log.Fatalf("Fail to OpenFile :%v", err)
		log.Printf("Fail to OpenFile :%v", err)
	}

	return handle
}

func mkDBLogDir() {
	dir, _ := os.Getwd() // os.Getwd：返回与当前目录对应的根路径名
	// os.MkdirAll：创建对应的目录以及所需的子目录，若成功则返回nil，否则返回error
	err := os.MkdirAll(dir+"/"+getDBLogFilePath(), os.ModePerm) // os.ModePerm：const定义ModePerm FileMode = 0777
	if err != nil {
		panic(err)
	}
}

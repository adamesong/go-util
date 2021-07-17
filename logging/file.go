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
	LogSavePath = "runtime/logs/" // log文件所在的目录
	LogSaveName = "log"           // log文件名
	LogFileExt  = "log"           // log文件后缀
	TimeFormat  = "20060102"      // 时间格式（日期格式）
)

// log文件所在的目录
func getLogFilePath() string {
	// return fmt.Sprintf("%s", LogSavePath)
	return LogSavePath
}

// log文件的目录+文件名
func getLogFileFullPath() string {
	prefixPath := getLogFilePath()
	suffixPath := fmt.Sprintf("%s%s.%s", LogSaveName, time.Now().Format(TimeFormat), LogFileExt)

	return fmt.Sprintf("%s%s", prefixPath, suffixPath)
}

func openLogFile(filePath string) *os.File {
	_, err := os.Stat(filePath) // 返回文件信息结构描述文件。如果出现错误，会返回*PathError
	switch {
	case os.IsNotExist(err):
		mkDir()
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

func mkDir() {
	dir, _ := os.Getwd() // os.Getwd：返回与当前目录对应的根路径名
	// os.MkdirAll：创建对应的目录以及所需的子目录，若成功则返回nil，否则返回error
	err := os.MkdirAll(dir+"/"+getLogFilePath(), os.ModePerm) // os.ModePerm：const定义ModePerm FileMode = 0777
	if err != nil {
		panic(err)
	}
}

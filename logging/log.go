package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	F                  *os.File
	logger             *log.Logger
	logPrefix          = ""
	levelFlags         = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	DefaultCallerDepth = 2
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func init() {
	filePath := getLogFileFullPath()
	F = openLogFile(filePath)
	logger = log.New(F, "", log.LstdFlags)
}

// Go语言如何把panic信息重定向：https://zhuanlan.zhihu.com/p/36199704
// https://codeday.me/bug/20180830/231370.html
// 暂时未用上
// func initPanicFile() error {
// 	log.Println("init panic file in unix mode")
// 	if err := syscall.Dup2(int(F.Fd()), int(os.Stderr.Fd())); err != nil {
// 		return err
// 	}
// 	return nil
// }

func Debug(v ...interface{}) {
	setPrefix(DEBUG)
	logger.Println(v...)
}

func Info(v ...interface{}) {
	setPrefix(INFO)
	logger.Println(v...)
}

func Warn(v ...interface{}) {
	setPrefix(WARNING)
	logger.Println(v...)
}

func Error(v ...interface{}) {
	setPrefix(ERROR)
	logger.Println(v...)
}

func Fatal(v ...interface{}) {
	setPrefix(FATAL)
	logger.Fatalln(v...)
}

func setPrefix(level int) {
	// 关于runtime.Caller的说明 https://www.flysnow.org/2017/05/06/go-in-action-go-log.html
	// 用于找到哪个文件中的第几行代码
	_, file, line, ok := runtime.Caller(DefaultCallerDepth)
	if ok {
		//logPrefix = fmt.Sprintf("%s : %s:%d ", levelFlags[level], filepath.Base(file), line)
		logPrefix = fmt.Sprintf("%s : %s:%d ", levelFlags[level], filepath.Clean(file), line)

	} else {
		logPrefix = fmt.Sprintf("%s :", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

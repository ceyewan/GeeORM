package log

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	mu       sync.Mutex
)

var (
	// Error 打印错误日志
	Error = errorLog.Println
	// Errorf 打印格式化的错误日志
	Errorf = errorLog.Printf
	// Info 打印信息日志
	Info = infoLog.Println
	// Infof 打印格式化的信息日志
	Infof = infoLog.Printf
)

// log levels
const (
	// InfoLevel 输出info和error
	InfoLevel = iota
	// ErrorLevel 只输出error
	ErrorLevel
	// Disabled 不输出任何日志
	Disabled
)

// SetLevel 设置日志级别，只有大于等于level的日志才会被输出
// level: InfoLevel, ErrorLevel, Disabled
// InfoLevel: 输出info和error
// ErrorLevel: 只输出error
// Disabled: 不输出任何日志
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()
	switch level {
	case InfoLevel:
		errorLog.SetOutput(os.Stdout)
		infoLog.SetOutput(os.Stdout)
	case ErrorLevel:
		infoLog.SetOutput(io.Discard)
		errorLog.SetOutput(os.Stdout)
	case Disabled:
		infoLog.SetOutput(io.Discard)
		errorLog.SetOutput(io.Discard)
	}
}

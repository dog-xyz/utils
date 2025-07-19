package gcl

import (
    "encoding/json"
    "fmt"
    "os"
    "runtime"
    "time"
	"path"
	"github.com/dog-xyz/utils/bes"
)

// LogLevel 日志级别
type LogLevel string

const (
    DEBUG LogLevel = "debug"
    INFO  LogLevel = "info"
    ERROR LogLevel = "error"
)

// LogEntry 日志条目结构
type LogEntry struct {
    QuickId    string    `json:"quick_id"`
    CodeInfo   string    `json:"code_info"`
    Level      LogLevel  `json:"level"`
    Msg        string    `json:"msg"`
    ServerName string    `json:"server_name"`
    Timestamp  time.Time `json:"timestamp"`
}

// Logger 日志器
type Logger struct {
	logIndexPrefix string
    serverName string
	batES      *bes.BatES
}

// New 创建新的日志器
func NewGCLogger(logIndexPrefix string, serverName string, batES *bes.BatES) *Logger {
    return &Logger{
        logIndexPrefix: logIndexPrefix,
        serverName: serverName,
		batES:      batES,
    }
}

// getCodeInfo 获取代码信息（文件名:行号 函数名）
func (l *Logger)getCallerInfo(skip int) (info string) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		info = "runtime.Caller() failed"
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	fileName := path.Base(file)
	return fmt.Sprintf("[%s] [%s.%d] ", funcName, fileName, lineNo)
}

// shouldLog 判断是否应该输出日志
/*
func (l *Logger) shouldLog(level LogLevel) bool {
    levelMap := map[LogLevel]int{
        DEBUG: 0,
        INFO:  1,
        ERROR: 2,
    }
    
    return levelMap[level] >= levelMap[l.level]
}
*/

// Debug 调试日志
func (l *Logger) Debugf(quickId string, format string, args ...interface{}) {
    entry := LogEntry{
		QuickId:    quickId,
        CodeInfo:   l.getCallerInfo(2),
        Level:      DEBUG,
        Msg:        fmt.Sprintf(format, args...),
        ServerName: l.serverName,
        Timestamp:  time.Now(),
    }
    
    // 输出 JSON 格式
    jsonData, err := json.Marshal(entry)
    if err != nil {
        fmt.Fprintf(os.Stderr, "json marshal failed: %v\n", err)
        return
    }
    fmt.Println(string(jsonData))
}

// Info 信息日志
func (l *Logger) Infof(quickId string, format string, args ...interface{}) {
    entry := LogEntry{
		QuickId:    quickId,
        CodeInfo:   l.getCallerInfo(2),
        Level:      INFO,
		Msg:        fmt.Sprintf(format, args...),
		ServerName: l.serverName,
		Timestamp:  time.Now(),
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal failed: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
    date := time.Now().Format("2006-01-02")
    if l.batES != nil {
        l.batES.Input <- bes.EsData{
            Index: l.logIndexPrefix + "_" + date,
            Data: entry,
        }
    }
}

// Error 错误日志
func (l *Logger) Errorf(quickId string, format string, args ...interface{}) {
    entry := LogEntry{
		QuickId:    quickId,
        CodeInfo:   l.getCallerInfo(2),
        Level:      ERROR,
		Msg:        fmt.Sprintf(format, args...),
		ServerName: l.serverName,
		Timestamp:  time.Now(),
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal failed: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
    date := time.Now().Format("2006-01-02")
    if l.batES != nil {
        l.batES.Input <- bes.EsData{
            Index: l.logIndexPrefix + "_" + date,
            Data: entry,
        }
    }
}
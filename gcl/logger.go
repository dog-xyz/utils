package gcl

import (
    "encoding/json"
    "fmt"
    "os"
    "runtime"
    "time"
	"path"
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
    CodeInfo   string    `json:"code_info"`
    Level      LogLevel  `json:"level"`
    Msg        string    `json:"msg"`
    ServerName string    `json:"server_name"`
    Timestamp  time.Time `json:"timestamp"`
}

// Logger 日志器
type Logger struct {
    serverName string
    level      LogLevel
}

// New 创建新的日志器
func NewGCLogger(serverName string) *Logger {
    return &Logger{
        serverName: serverName,
        level:      INFO, // 默认级别
    }
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
    l.level = level
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
func (l *Logger) shouldLog(level LogLevel) bool {
    levelMap := map[LogLevel]int{
        DEBUG: 0,
        INFO:  1,
        ERROR: 2,
    }
    
    return levelMap[level] >= levelMap[l.level]
}


// Debug 调试日志
func (l *Logger) Debugf(format string, args ...interface{}) {
    if !l.shouldLog(DEBUG) {
        return
    }
    
    entry := LogEntry{
        CodeInfo:   l.getCallerInfo(2),
        Level:      DEBUG,
        Msg:        fmt.Sprintf(format, args...),
        ServerName: l.serverName,
        Timestamp:  time.Now(),
    }
    
    // 输出 JSON 格式
    jsonData, err := json.Marshal(entry)
    if err != nil {
        fmt.Fprintf(os.Stderr, "日志序列化失败: %v\n", err)
        return
    }
    
    fmt.Println(string(jsonData))
}

// Info 信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
    if !l.shouldLog(INFO) {
        return
    }
    
    entry := LogEntry{
        CodeInfo:   l.getCallerInfo(2),
        Level:      INFO,
		Msg:        fmt.Sprintf(format, args...),
		ServerName: l.serverName,
		Timestamp:  time.Now(),
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志序列化失败: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}

// Error 错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
    if !l.shouldLog(ERROR) {
        return
    }
    
    entry := LogEntry{
        CodeInfo:   l.getCallerInfo(2),
        Level:      ERROR,
		Msg:        fmt.Sprintf(format, args...),
		ServerName: l.serverName,
		Timestamp:  time.Now(),
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志序列化失败: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}
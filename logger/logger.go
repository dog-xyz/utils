package logger

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/cihub/seelog"
)

const (
	defaultLoggerConfig = `<seelog type="asynctimer" asyncinterval="5000000" minlevel="debug" maxlevel="error">
		<outputs formatid="rolllog">
			<rollingfile formatid="rolllog" type="size" filename="SERVER_NAME.log" maxsize="10485760" maxrolls="5" />
		</outputs>
		<formats>
			<format id="rolllog" format="%Date %Time [%LEVEL] [%File.%Line] %Msg%n"/>
		</formats>
	</seelog>`

	asyncInterval = 10 * time.Second
)

type Logger struct {
	SeeLogger      seelog.LoggerInterface
	ConfigPath string
	SrvName string
	SrvHost string
	done chan struct{}
}

type LogData struct {
	LogType    string `json:"log_type"`
	SrvName string `json:"service_name"`
	SrvHost string `json:"service_host"`
	LogLevel    string `json:"log_level"`
	CodeInfo    string `json:"code_info"`
	Msg         string `json:"msg"`
	LogTime     int64  `json:"hids_log_time"`
}

type InitArgs struct {
	ConfigPath  string
	SrvName string
	SrvHost string
}

func getCallerInfo(skip int) (info string) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		info = "runtime.Caller() failed"
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	fileName := path.Base(file)
	return fmt.Sprintf("[%s] [%s.%d] ", funcName, fileName, lineNo)
}

func NewLogger(args InitArgs) *Logger {
	if args.ConfigPath == "" {
		fmt.Printf("log config path is empty\n")
		return nil
	}
	seeLogger, err := seelog.LoggerFromConfigAsFile(args.ConfigPath)
	if err != nil {
		panic(fmt.Errorf("seelog init failed: %w", err))
	}

	logger := &Logger{
		SeeLogger:      seeLogger,
		ConfigPath: args.ConfigPath,
		SrvName: args.SrvName,
		SrvHost: args.SrvHost,
		done: make(chan struct{}),
	}
	return logger
}

func NewDefaultLogger(args InitArgs) *Logger {
	configContent := strings.Replace(defaultLoggerConfig, "SERVER_NAME", args.SrvName, -1)
	seeLogger, err := seelog.LoggerFromConfigAsBytes([]byte(configContent))
	if err != nil {
		panic(fmt.Errorf("seelog init failed: %w", err))
	}
	lg := &Logger{
		SeeLogger:      seeLogger,
		ConfigPath: args.ConfigPath,
		SrvName: args.SrvName,
		SrvHost: args.SrvHost,
		done: make(chan struct{}),
	}
	go lg.run()
	return lg
}

func (lg *Logger) run() {
	t := time.NewTicker(asyncInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			lg.SeeLogger.Flush()
		case <-lg.done:
			lg.SeeLogger.Infof("logger done")
			return
		}
	}
}

func (lg *Logger) Stop() {
	close(lg.done)
}

func (lg *Logger) Errorf(format string, params ...interface{}) {
	lg.SeeLogger.Errorf(getCallerInfo(2)+format, params...)
}

func (lg *Logger) Warnf(format string, params ...interface{}) {
	lg.SeeLogger.Warnf(getCallerInfo(2)+format, params...)
}

func (lg *Logger) Debugf(format string, params ...interface{}) {
	lg.SeeLogger.Debugf(getCallerInfo(2)+format, params...)
}

func (lg *Logger) Infof(format string, params ...interface{}) {
	lg.SeeLogger.Infof(getCallerInfo(2)+format, params...)
}

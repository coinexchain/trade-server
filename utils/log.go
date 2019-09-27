package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pelletier/go-toml"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const (
	PlainFormat = "plain"
	JSONFormat  = "json"

	FileName = "server.log"
)

func InitLog(svrConfig *toml.Tree) error {
	logDir := svrConfig.GetDefault("log-dir", "log").(string)
	level := svrConfig.GetDefault("log-level", "info").(string)
	format := svrConfig.GetDefault("log-format", "plain").(string)

	if _, err := os.Stat(logDir); err != nil && os.IsNotExist(err) {
		if err = os.Mkdir(logDir, 0755); err != nil {
			fmt.Print(err)
			return err
		}
	}

	var formatter logrus.Formatter
	if format == JSONFormat {
		formatter = &logrus.JSONFormatter{CallerPrettyfier: callerPrettyfier}
	} else {
		formatter = &PlainFormatter{}
	}
	logrus.SetFormatter(formatter)
	logrus.SetReportCaller(true)
	logrus.SetLevel(getLogLevel(level))
	logrus.AddHook(newRotateHook(logDir, FileName, 24*time.Hour, formatter))

	return nil
}

func newRotateHook(logPath string, logFileName string, rotationTime time.Duration, formatter logrus.Formatter) *lfshook.LfsHook {
	absPath, _ := filepath.Abs(logPath)
	baseLogPath := path.Join(absPath, logFileName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y-%m-%d",
		rotatelogs.WithLinkName(baseLogPath),
		rotatelogs.WithRotationTime(rotationTime),
	)
	if err != nil {
		fmt.Printf("config local file system logger error: %v", err)
	}

	return lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, formatter)
}

func getLogLevel(level string) logrus.Level {
	level = strings.ToLower(level)
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.WarnLevel
	}
}

func callerPrettyfier(frame *runtime.Frame) (string, string) {
	_, fileName := path.Split(frame.File)
	file := fmt.Sprintf("%s:%d", fileName, frame.Line)

	splitIdx := strings.LastIndex(frame.Function, "/")
	function := frame.Function[splitIdx+1:]

	return function, file
}

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	level := strings.ToUpper(entry.Level.String())
	logTime := entry.Time.Format("2006-01-02 15:04:05")
	message := strings.TrimRight(entry.Message, "\n")

	fileName := ""
	lineNum := 0
	if entry.HasCaller() {
		_, fileName = path.Split(entry.Caller.File)
		lineNum = entry.Caller.Line
	}

	fields := ""
	for k, v := range entry.Data {
		fields += fmt.Sprintf("%v:%v | ", k, v)
	}
	if len(fields) > 0 {
		fields = "| " + fields
	}

	output := fmt.Sprintf("[%v] %v %v:%v %v %v\n",
		level, logTime, fileName, lineNum, message, fields)

	return []byte(output), nil
}

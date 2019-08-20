package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml"
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
	file, err := os.OpenFile(logDir+"/"+FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Print(err)
		return err
	}

	if format == JSONFormat {
		logrus.SetFormatter(&logrus.JSONFormatter{CallerPrettyfier: callerPrettyfier})
	} else {
		logrus.SetFormatter(&PlainFormatter{})
	}
	logrus.SetReportCaller(true)
	logrus.SetLevel(getLogLevel(level))
	logrus.SetOutput(file)

	return nil
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
	dirs := strings.Split(frame.File, "/")
	fileName := dirs[len(dirs)-1]
	file := fmt.Sprintf("%s:%d", fileName, frame.Line)

	funcPaths := strings.Split(frame.Function, "/")
	function := funcPaths[len(funcPaths)-1]

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
		dirs := strings.Split(entry.Caller.File, "/")
		fileName = dirs[len(dirs)-1]
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

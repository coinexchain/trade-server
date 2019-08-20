package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

const (
	PlainFormat = "plain"
	JSONFormat  = "json"

	FileName     = "server.log"
	LineLogFmt   = "[%v] %v %v:%v %v"
	LineLogFmtRn = LineLogFmt + "\n"
)

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	dirs := strings.Split(entry.Caller.File, "/")
	fileName := dirs[len(dirs)-1]

	format := LineLogFmt
	if entry.Message[len(entry.Message)-1] != '\n' {
		format = LineLogFmtRn
	}
	output := fmt.Sprintf(format,
		strings.ToUpper(entry.Level.String()),
		entry.Time.Format("2006-01-02 15:04:05"),
		fileName,
		entry.Caller.Line,
		entry.Message)

	for k, val := range entry.Data {
		switch v := val.(type) {
		case string:
			output = strings.Replace(output, "%"+k+"%", v, 1)
		case int:
			s := strconv.Itoa(v)
			output = strings.Replace(output, "%"+k+"%", s, 1)
		case bool:
			s := strconv.FormatBool(v)
			output = strings.Replace(output, "%"+k+"%", s, 1)
		}
	}

	return []byte(output), nil
}

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
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetReportCaller(false)
	} else {
		logrus.SetFormatter(&PlainFormatter{})
		logrus.SetReportCaller(true)
	}
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

package zklog

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type stdFormatter struct{}
type color int

const (
	black color = iota
	red
	green
	yellow
	blue
	purple
	gray
)

func printWithColor(c color, msg string) {
	fmt.Printf("\033[4%dm%v\033[0m]", c, msg)
}
func (std *stdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var c color
	switch entry.Level {
	case logrus.InfoLevel:
		c = green
	case logrus.ErrorLevel:
		c = red
	case logrus.WarnLevel:
		c = yellow
	case logrus.DebugLevel:
		c = purple
	}
	var b *bytes.Buffer

	if entry.Buffer == nil {
		b = &bytes.Buffer{}
	} else {
		b = entry.Buffer
	}

	data := make([]string, 0)
	for k, v := range entry.Data {
		data = append(data, fmt.Sprintf("%s => %v", k, v))
	}
	// 设置格式
	fmt.Fprintf(b, "%v \033[3%dm[%v]-[%v | %v() line:%v]\033[0m \033[4%dm%v\033[0m\n",
		entry.Time.Format("2006-01-02 15:04:05"), c, entry.Level, entry.Caller.File, entry.Caller.Func.Name(), entry.Caller.Line,
		c, strings.Join(data, ";  "),
	)
	return b.Bytes(), nil
}

func init() {
	logFilePath := "logs"
	logFileName := "distributed"
	// 日志文件
	fileName := path.Join(logFilePath, logFileName)

	// 实例化
	Logger = logrus.New()
	// 设置输出
	// logger.Out = src
	// 设置日志级别
	Logger.SetLevel(logrus.TraceLevel)
	// Logger.SetFormatter(&logrus.TextFormatter{
	// 	TimestampFormat:           "2006-01-02 15:04:05",
	// 	ForceColors:               true,
	// 	EnvironmentOverrideColors: true,
	// 	FullTimestamp:             true,
	// })
	Logger.SetFormatter(&stdFormatter{})
	// 打印文件名和行号
	Logger.SetReportCaller(true)

	// 写入文件
	logWriter, _ := rotatelogs.New(
		// 文件名称
		fileName+"-%Y%m%d.log",
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}
	lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	// 新增 Hook
	Logger.AddHook(lfHook)
}

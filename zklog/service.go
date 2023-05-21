package zklog

import (
	"path"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var Logger *log.Logger

func init() {
	logFilePath := "/home/zzk/go/zkCacheDemo/zkCache/zklog"
	logFileName := "distributed"
	// 日志文件
	fileName := path.Join(logFilePath, logFileName)
	// 写入文件
	// src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	// if err != nil {
	// 	fmt.Println("err", err)
	// }
	// 实例化
	Logger = log.New()
	// 设置输出
	// logger.Out = src
	// 设置日志级别
	Logger.SetLevel(log.DebugLevel)
	Logger.SetFormatter(&log.TextFormatter{
		// TimestampFormat: "2006-01-02 15:04:05",
		TimestampFormat:           "2006-01-02 15:04:05",
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		FullTimestamp:             true,
	})
	logWriter, _ := rotatelogs.New(
		// 文件名称
		fileName+"-%Y%m%d.log",
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	writeMap := lfshook.WriterMap{
		log.InfoLevel:  logWriter,
		log.FatalLevel: logWriter,
		log.DebugLevel: logWriter,
		log.WarnLevel:  logWriter,
		log.ErrorLevel: logWriter,
		log.PanicLevel: logWriter,
	}
	lfHook := lfshook.NewHook(writeMap, &log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	// 新增 Hook
	Logger.AddHook(lfHook)
}

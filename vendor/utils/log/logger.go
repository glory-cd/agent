/*
@Time : 19-5-8 上午10:35
@Author : liupeng
@File : logger
*/

package log

import (
	"github.com/wantedly/gorm-zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var Slogger *zap.SugaredLogger
var Logger *zap.Logger

var logLevel = zap.NewAtomicLevel()

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func InitLog(filepath string, filemaxsize int, maxbackups int, maxage int, iscompress bool) {

	hook := lumberjack.Logger{
		Filename:   filepath,    // 日志文件路径
		MaxSize:    filemaxsize, // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: maxbackups,  // 日志文件最多保存多少个备份
		MaxAge:     maxage,      // 文件最多保存多少天
		Compress:   iscompress,  // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    //  大写编码器
		EncodeTime:     TimeEncoder,                    // 自定义时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	//logLevel.SetLevel(zapcore.DebugLevel)

	core := zapcore.NewCore(
		//zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		logLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	develoment := zap.Development()
	// 设置初始化字段
	//filed := zap.Fields(zap.String("appName", "agent"))
	// 构造日志
	Logger = zap.New(core, caller, develoment)

	defer Logger.Sync()

	Slogger = Logger.Sugar()
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

// 设置日志级别
func SetLevel(level string) {
	logLevel.SetLevel(getLoggerLevel(level))
}

func GetDBLogger() *gormzap.Logger {
	return gormzap.New(Logger)
}

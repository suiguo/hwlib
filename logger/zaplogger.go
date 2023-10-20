package logger

import (
	"fmt"
	"os"
	"sync"

	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// StdLogger is struct
type StdLogger struct {
	logger *zap.SugaredLogger
	Logger
}
type LoggerCfg struct {
	Name       string `json:"name"`
	Maxsize    int    `json:"maxsize"`
	Maxbackups int    `json:"maxbackups"`
	Maxage     int    `json:"maxage"`
	Compress   bool   `json:"compress"`
	Level      int    `json:"level"`
}

var once sync.Once
var instance *StdLogger
var levelInstance *StdLogger
var logger_map = make(map[string]*StdLogger)
var logger_lock sync.RWMutex

/*
   "name": "stdout",
        "maxsize": 500,
        "maxbackups": 5,
        "maxage": 20160,
        "compress": false,
        "level": 0
*/

// 获取一个默认log 对象 用于不是项目内部的打印 默认路径是./log
//
//	func GetDefaultInstance() (*StdLogger, error) {
//		return GetInstance("default_log_instance", defaultCfg)
//		// return log,err
//	}
//
// skipCall打印调用关系的堆栈
func GetInstance(tag string, cfg []*LoggerCfg, skipCall int) (*StdLogger, error) {
	if cfg == nil || len(cfg) <= 0 {
		return nil, fmt.Errorf("logger cfg is err")
	}
	logger_lock.RLock()
	instance := logger_map[tag]
	logger_lock.RUnlock()
	if instance == nil {
		cores := make([]zapcore.Core, 0)
		encoder := getEncoder()
		for _, v := range cfg {
			if v == nil {
				return nil, fmt.Errorf("cfg is nil")
			}
			switch v.Name {
			case "stdout":
				cores = append(cores, zapcore.NewCore(encoder, os.Stdout, zapcore.Level(v.Level)))
			case "stderr":
				cores = append(cores, zapcore.NewCore(encoder, os.Stderr, zapcore.Level(v.Level)))
			default:
				cores = append(cores, zapcore.NewCore(encoder, getCfgWriter(v), zapcore.Level(v.Level)))
			}
		}
		handler := zapcore.NewTee(cores...)
		// opt := []zap.Option{zap.AddCaller()}
		zaplogger := zap.New(handler, zap.AddCaller(), zap.AddCallerSkip(skipCall)) //不打印堆栈
		sugarLogger := zaplogger.Sugar()
		instance = &StdLogger{
			logger: sugarLogger,
		}
		logger_lock.Lock()
		logger_map[tag] = instance
		logger_lock.Unlock()
	}
	return instance, nil
}

// NewStdLogger is 输出到控制台. 多次调用只初始化一次
func NewStdLogger(callerSkip int) *StdLogger {
	if instance == nil {
		once.Do(func() {
			writeSyncer := getLogWriter()
			encoder := getEncoder()
			cores := make([]zapcore.Core, 0)
			cores = append(cores, zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel))
			handler := zapcore.NewTee(cores...)
			zaplogger := zap.New(handler, zap.AddCaller(), zap.AddCallerSkip(callerSkip)) //修改堆栈深度
			sugarLogger := zaplogger.Sugar()
			instance = &StdLogger{
				logger: sugarLogger,
			}
		})
	}
	return instance
}

// NewStdLogger is 输出到控制台. 多次调用只初始化一次。
//
// level
//
//	DebugLevel -1 InfoLevel 0 WarnLevel 1 ErrorLevel  2 DPanicLevel 3  PanicLevel 4  FatalLevel 5
func NewStdLoggerWithLevel(callerSkip int, level int) *StdLogger {
	if levelInstance == nil {
		once.Do(func() {
			writeSyncer := getLogWriter()
			encoder := getEncoder()
			cores := make([]zapcore.Core, 0)
			cores = append(cores, zapcore.NewCore(encoder, writeSyncer, zapcore.Level(level)))
			handler := zapcore.NewTee(cores...)
			zaplogger := zap.New(handler, zap.AddCaller(), zap.AddCallerSkip(callerSkip)) //修改堆栈深度
			sugarLogger := zaplogger.Sugar()
			levelInstance = &StdLogger{
				logger: sugarLogger,
			}
		})
	}
	return levelInstance
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	// cores = append(cores, zapcore.NewCore(encoder, os.Stdout, zapcore.Level(v.Level)))

	// lumberJackLogger := &lumberjack.Logger{
	// 	Filename:   "./logs/log.txt",
	// 	MaxSize:    1024,
	// 	MaxBackups: 10,
	// 	MaxAge:     60 * 24 * 14,
	// 	Compress:   false,
	// }
	return zapcore.AddSync(os.Stdout)
}

func getCfgWriter(cfg *LoggerCfg) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.Name,
		MaxSize:    cfg.Maxsize,
		MaxBackups: cfg.Maxbackups,
		MaxAge:     cfg.Maxage,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// Debug is for log warning level
func (l *StdLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debugw(msg, fields...)
}

// Info is for log warning level
func (l *StdLogger) Info(msg string, fields ...interface{}) {
	l.logger.Infow(msg, fields...)
}

// Error is for log warning level
func (l *StdLogger) Error(msg string, fields ...interface{}) {
	l.logger.Errorw(msg, fields...)
}

// Fatal is for log warning level
func (l *StdLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatalw(msg, fields...)
}

// Panic is for log warning level
func (l *StdLogger) Panic(msg string, fields ...interface{}) {
	l.logger.Panicw(msg, fields...)
}

func (l *StdLogger) Warning(msg string, fields ...interface{}) {
	l.logger.Warnw(msg, fields...)
}

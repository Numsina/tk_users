package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	l            *Logger
	ouWrite      zapcore.WriteSyncer       //IO输出
	debugConsole = zapcore.Lock(os.Stdout) // 控制台标准输出
	once         sync.Once
)

type Logger struct {
	*zap.Logger
	opts      *Option
	zapConfig zap.Config
}

func NewLogger(opts ...Options) *Logger {
	if l == nil {
		once.Do(func() {
			l = &Logger{
				opts: newOptions(opts...),
			}
			l.loadCfg()
			l.initLog()
			l.Info("[initLogger] zap plugin initializing completed")
		})
	}
	return l
}

func GetLogger() *Logger {
	if l == nil {
		fmt.Println("Please initialize the log service")
		return nil
	}

	return l
}

func (l *Logger) GetLoggerFromCtx(ctx context.Context) *zap.Logger {
	log, ok := ctx.Value(l.opts.CtxKey).(*zap.Logger)

	if ok {
		return log
	}

	return l.Logger
}

// 添加额外上下文信息
func (l *Logger) AddCtx(ctx context.Context, field ...zap.Field) (context.Context, *zap.Logger) {
	log := l.With(field...)
	ctx = context.WithValue(ctx, l.opts.CtxKey, log)
	return ctx, log
}

func (l *Logger) initLog() {
	l.setSyncers()
	var err error
	l.Logger, err = l.zapConfig.Build(l.cores())
	if err != nil {
		panic(err)
	}

	defer l.Logger.Sync()
}

func (l *Logger) GetLevel() (level zapcore.Level) {
	switch strings.ToLower(l.opts.Level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.DebugLevel // 默认debug级别
	}
}

func (l *Logger) loadCfg() {
	if l.opts.Development {
		l.zapConfig = zap.NewDevelopmentConfig()
	} else {
		l.zapConfig = zap.NewProductionConfig()
	}
}

func (l *Logger) setSyncers() {
	_ = zapcore.AddSync(&lumberjack.Logger{
		Filename:   l.opts.LogFileDir + "/" + l.opts.AppName + "./log",
		MaxSize:    l.opts.MaxSize,
		MaxBackups: l.opts.MaxBackups,
		MaxAge:     l.opts.MaxAge,
		Compress:   true,
		LocalTime:  true,
	})
	return
}

func (l *Logger) cores() zap.Option {
	encoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)
	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= l.GetLevel()
	})

	var cores []zapcore.Core

	if l.opts.WriteFile {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(encoder, ouWrite, priority),
		}...)
	}

	if l.opts.WriteConsole {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(encoder, debugConsole, priority),
		}...)
	}

	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

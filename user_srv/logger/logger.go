package logger

import "path/filepath"

type Option struct {
	Development  bool
	LogFileDir   string
	AppName      string
	MaxSize      int    //文件大小阈值，超过阈值进行文件切割
	MaxBackups   int    // 文件备份数量
	MaxAge       int    // 文件最大保存时间
	Level        string // 级别
	CtxKey       string // 通过ctx传递log对象
	WriteFile    bool
	WriteConsole bool
}

type Options func(o *Option)

func newOptions(opts ...Options) *Option {
	op := &Option{
		Development:  true,
		AppName:      "logApp",
		MaxSize:      100,
		MaxBackups:   60,
		MaxAge:       30,
		Level:        "debug",
		CtxKey:       "log-ctx",
		WriteFile:    false,
		WriteConsole: true,
	}

	op.LogFileDir, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
	op.LogFileDir += "/logs/"
	for _, o := range opts {
		o(op)
	}
	return op
}

func WithDevelopment(development bool) Options {
	return func(o *Option) {
		o.Development = development
	}
}

func WithtLogFileDir(logFileDir string) Options {
	return func(o *Option) {
		o.LogFileDir = logFileDir
	}
}

func WithAppName(appName string) Options {
	return func(o *Option) {
		o.AppName = appName
	}
}

func WithMaxSize(maxSize int) Options {
	return func(o *Option) {
		o.MaxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) Options {
	return func(o *Option) {
		o.MaxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) Options {
	return func(o *Option) {
		o.MaxAge = maxAge
	}
}

func WithLevel(level string) Options {
	return func(o *Option) {
		o.Level = level
	}
}

func WithCtxKey(ctxKey string) Options {
	return func(o *Option) {
		o.CtxKey = ctxKey
	}
}

func WithWriteFile(writeFile bool) Options {
	return func(o *Option) {
		o.WriteFile = writeFile
	}
}

func WithWriteConsole(writeConsole bool) Options {
	return func(o *Option) {
		o.WriteConsole = writeConsole
	}
}

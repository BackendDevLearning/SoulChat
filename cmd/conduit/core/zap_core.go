package core

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    conf "kratos-realworld/internal/conf"
    "os"
    "time"
)

type ZapCore struct {
	level zapcore.Level
	zapcore.Core
	logConf *conf.Log
}

func NewZapCore(level zapcore.Level, logConf *conf.Log) *ZapCore {
	entity := &ZapCore{level: level, logConf: logConf}
	syncer := entity.WriteSyncer()
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l == level
	})
	entity.Core = zapcore.NewCore(entity.DiyEncoder(), syncer, levelEnabler)
	return entity
}

// NewZapCoreForError 创建一个专门用于 error 级别日志的 core，写入 error.log
func NewZapCoreForError(logConf *conf.Log) *ZapCore {
	entity := &ZapCore{level: zapcore.ErrorLevel, logConf: logConf}
	syncer := entity.WriteSyncerForError()
	// 只接受 error 及以上级别（error, dpanic, panic, fatal）
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.ErrorLevel
	})
	entity.Core = zapcore.NewCore(entity.DiyEncoder(), syncer, levelEnabler)
	return entity
}

func (z *ZapCore) WriteSyncer(formats ...string) zapcore.WriteSyncer {
	cutter := NewCutter(
		z.logConf.Director,
		"all",
		CutterWithLayout(time.DateOnly),
		CutterWithFormats(formats...),
	)

	// 是否输出到终端
	if z.logConf.LogInConsole {
		return zapcore.AddSync(os.Stdout)
	}

	return zapcore.AddSync(cutter)
}

// WriteSyncerForError 创建专门用于 error.log 的 WriteSyncer
func (z *ZapCore) WriteSyncerForError(formats ...string) zapcore.WriteSyncer {
	cutter := NewCutter(
		z.logConf.Director,
		"error",
		CutterWithLayout(time.DateOnly),
		CutterWithFormats(formats...),
	)

	// 是否输出到终端
	if z.logConf.LogInConsole {
		return zapcore.AddSync(os.Stdout)
	}

	return zapcore.AddSync(cutter)
}

func (z *ZapCore) Enabled(level zapcore.Level) bool {
	return z.level == level
}

func (z *ZapCore) With(fields []zapcore.Field) zapcore.Core {
	return z.Core.With(fields)
}

func (z *ZapCore) Check(entry zapcore.Entry, check *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if z.Enabled(entry.Level) {
		return check.AddCore(entry, z)
	}
	return check
}

func (z *ZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	for i := 0; i < len(fields); i++ {
		if fields[i].Key == "service" || fields[i].Key == "biz" || fields[i].Key == "data" {
			syncer := z.WriteSyncer(fields[i].String)
			z.Core = zapcore.NewCore(z.DiyEncoder(), syncer, z.level)
		}
	}

	return z.Core.Write(entry, fields)
}

func (z *ZapCore) Sync() error {
	return z.Core.Sync()
}

func (z *ZapCore) DiyEncoder() zapcore.Encoder {
	config := zapcore.EncoderConfig{
		TimeKey:       "time",
		NameKey:       "name",
		LevelKey:      "level",
		CallerKey:     "caller",
		MessageKey:    "message",
		StacktraceKey: z.logConf.StacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeLevel:    z.LevelEncoder(),
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	if z.logConf.Format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)

}

// LevelEncoder 根据 EncodeLevel 返回 zapcore.LevelEncoder
// Author [SliverHorn](https://github.com/SliverHorn)
func (z *ZapCore) LevelEncoder() zapcore.LevelEncoder {
	switch {
	case z.logConf.EncodeLevel == "LowercaseLevelEncoder": // 小写编码器(默认)
		return zapcore.LowercaseLevelEncoder
	case z.logConf.EncodeLevel == "LowercaseColorLevelEncoder": // 小写编码器带颜色
		return zapcore.LowercaseColorLevelEncoder
	case z.logConf.EncodeLevel == "CapitalLevelEncoder": // 大写编码器
		return zapcore.CapitalLevelEncoder
	case z.logConf.EncodeLevel == "CapitalColorLevelEncoder": // 大写编码器带颜色
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

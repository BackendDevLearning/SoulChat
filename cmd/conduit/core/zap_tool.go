package core

import (
    "fmt"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    conf "kratos-realworld/internal/conf"
    "os"
)

// Levels 根据字符串转化为 zapcore.Levels
func Levels(levelStr string) []zapcore.Level {
	levels := make([]zapcore.Level, 0, 7)
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		level = zapcore.DebugLevel
	}
	for ; level <= zapcore.FatalLevel; level++ {
		levels = append(levels, level)
	}
	return levels
}

// Zap 获取 zap.Logger
// Author [SliverHorn](https://github.com/SliverHorn)
func Zap(conf *conf.Log) (logger *zap.Logger) {
    if _, err := os.Stat(conf.Director); os.IsNotExist(err) { // 判断是否有Director文件夹
        fmt.Printf("create %v directory\n", conf.Director)
        _ = os.Mkdir(conf.Director, os.ModePerm)
    }
    levels := Levels(conf.Level)
	length := len(levels)
	cores := make([]zapcore.Core, 0, length+1)
	
	// 创建所有级别的 core，都写入 all.log
	for i := 0; i < length; i++ {
		core := NewZapCore(levels[i], conf)
		cores = append(cores, core)
	}
	
	// 额外创建一个 error 级别的 core，写入 error.log
	errorCore := NewZapCoreForError(conf)
	cores = append(cores, errorCore)
	
	logger = zap.New(zapcore.NewTee(cores...))
	if conf.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	return logger
}

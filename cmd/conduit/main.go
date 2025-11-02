package main

import (
	"flag"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/data/migrate"
	"kratos-realworld/internal/pkg/env"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func initLogger(bc conf.Bootstrap) log.Logger {
	logConf := &conf.Log{
		Director:      bc.Log.Director,  // 日志目录
		LogInConsole:  true,              // 开发时输出到控制台
		Format:        bc.Log.Format,            // 生产环境用 json
		StacktraceKey: "stacktrace",
		EncodeLevel:   "LowercaseLevelEncoder",
	}
	
	// 创建不同级别的核心
	debugCore := core.NewZapCore(zapcore.DebugLevel, logConf)
	infoCore := core.NewZapCore(zapcore.InfoLevel, logConf)
	warnCore := core.NewZapCore(zapcore.WarnLevel, logConf)
	errorCore := core.NewZapCore(zapcore.ErrorLevel, logConf)
	
	// 组合核心
	teeCore := zapcore.NewTee(debugCore, infoCore, warnCore, errorCore)
	
	// 创建 Zap Logger
	zapLogger := zap.New(teeCore, 
		zap.AddCaller(),      // 添加调用信息
		zap.AddCallerSkip(1), // 跳过一层调用栈
	)
	
	// 适配成 Kratos Logger
	return core.NewZapLoggerAdapter(zapLogger)
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	logger := initLogger(bc)
	
	// 设置全局 logger
	log.SetLogger(logger)
	app, cleanup, err := initApp(bc.Server, bc.Data, bc.Jwt, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 仅在开发环境执行数据库中table的创建与迁移
	if env.IsDev() {
		if err := migrate.InitDBTable(app.DB); err != nil {
			log.Fatal("Failed to migrate database:", err)
		}
	}

	// start and wait for stop signal
	if err := app.App.Run(); err != nil {
		panic(err)
	}
}

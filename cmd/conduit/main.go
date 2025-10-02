package main

import (
	"flag"
	"kratos-realworld/internal/conf"
    "kratos-realworld/internal/kafka"
    wsrv "kratos-realworld/internal/websocket"
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

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	)
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

    // Kafka: 生产者/消费者初始化
    if bc.Data.Kafka != nil && bc.Data.Kafka.Enabled {
        kafka.InitProducer(bc.Data.Kafka.Topic, bc.Data.Kafka.Hosts)
        kafka.InitConsumer(bc.Data.Kafka.Hosts)
        go kafka.ConsumerMsg(wsrv.ConsumerKafkaMsg)
        defer kafka.Close()
        defer kafka.CloseConsumer()
    }

    // 配置静态目录（可选）
    if bc.Data != nil && bc.Data.Storage != nil {
        wsrv.SetStaticBaseDir(bc.Data.Storage.StaticDir)
    }

    // 启动websocket服务
    go wsrv.MyServer.Start()

	// start and wait for stop signal
	if err := app.App.Run(); err != nil {
		panic(err)
	}
}

# zap 日志库

## 高性能设计：

**无反射**：通过强类型的 API（如zap.Int("port", 8080)）避免**反射开销**。
零分配：核心组件使用**预分配缓冲区**，减少 GC 压力。
Benchmark：
zap的结构化日志比log快 10-100 倍。
sugar 非结构化日志（如logger.Info("message")）性能接近原生fmt。
zap 结构化日志：以键值对形式组织日志，支持多种数据类型：

sugar 非结构化：

```go
logger, _ := zap.NewProduction()
defer logger.Sync() // 在程序结束时将缓存同步到文件中
sugar := logger.Sugar()
sugar.Infow("failed to fetch URL",
  "url", url,
  "attempt", 3,
  "backoff", time.Second,
)
sugar.Infof("Failed to fetch URL: %s", url)
```



zap 结构化

```go
logger, _ := zap.NewProduction()
defer logger.Sync()
logger.Info("failed to fetch URL",
  // Structured context as strongly typed Field values.
  zap.String("url", url),
  zap.Int("attempt", 3),
  zap.Duration("backoff", time.Second),
)
```

> Zap 的使用非常简单，麻烦的点在于配置出一个适合自己项目的日志，官方例子很少，要多读源代码注释。



## zap 核心

```go
type ioCore struct {
   // 日志级别
   LevelEnabler
   // 日志编码
   enc Encoder
   // 日志书写
   out WriteSyncer
}

// core 三个参数之  编码
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
    return zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
}

// core 三个参数之  路径
func getLogWriter() zapcore.WriteSyncer {
	file,_ := os.Create("E:/test.log")
	return zapcore.AddSync(file)
}

通个核心参数初始化
var logger *zap.Logger
var sugarLogger *zap.SugaredLogger

func InitLogger() {
	encoder := getEncoder()
	writerSyncer := getLogWriter()
    // 集齐三个核心参数即可
	core := zapcore.NewCore(encoder,writerSyncer,zapcore.DebugLevel)
	logger = zap.New(core, zap.AddCaller())
	sugarLogger = logger.Sugar()
}


一般的调用过程
logger, _ := zap.NewProduction()
defer logger.Sync() // 在程序结束时将缓存同步到文件中
// sugar := logger.Sugar()


```



基本上有三个核心参数就可以正常调用，但是要想个性化设置还需要额外的配置



## 总的配置

```golang
type Config struct {
    // 最小日志级别
   Level AtomicLevel `json:"level" yaml:"level"`
    // 开发模式，主要影响堆栈跟踪
   Development bool `json:"development" yaml:"development"`
    // 调用者追踪
   DisableCaller bool `json:"disableCaller" yaml:"disableCaller"`
    // 堆栈跟踪
   DisableStacktrace bool `json:"disableStacktrace" yaml:"disableStacktrace"`
    // 采样，在限制日志对性能占用的情况下仅记录部分比较有代表性的日志，等于日志选择性记录
   Sampling *SamplingConfig `json:"sampling" yaml:"sampling"`
    // 编码，分为json和console两种模式
   Encoding string `json:"encoding" yaml:"encoding"`
    // 编码配置，主要是一些输出格式化的配置
   EncoderConfig zapcore.EncoderConfig `json:"encoderConfig" yaml:"encoderConfig"`
    // 日志文件输出路径
   OutputPaths []string `json:"outputPaths" yaml:"outputPaths"`
    // 错误文件输出路径
   ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`
    // 给日志添加一些默认输出的内容
   InitialFields map[string]interface{} `json:"initialFields" yaml:"initialFields"`
}


```



### 三个核心中的日志编码

```go

编码配置，放到三个核心中的日志编码中
type EncoderConfig struct {
   // 键值，如果key为空，那么对于的属性将不会输出
   MessageKey     string `json:"messageKey" yaml:"messageKey"`
   LevelKey       string `json:"levelKey" yaml:"levelKey"`
   TimeKey        string `json:"timeKey" yaml:"timeKey"`
   NameKey        string `json:"nameKey" yaml:"nameKey"`
   CallerKey      string `json:"callerKey" yaml:"callerKey"`
   FunctionKey    string `json:"functionKey" yaml:"functionKey"`
   StacktraceKey  string `json:"stacktraceKey" yaml:"stacktraceKey"`
   SkipLineEnding bool   `json:"skipLineEnding" yaml:"skipLineEnding"`
   LineEnding     string `json:"lineEnding" yaml:"lineEnding"`
   // 一些自定义的编码器
   EncodeLevel    LevelEncoder    `json:"levelEncoder" yaml:"levelEncoder"`
   EncodeTime     TimeEncoder     `json:"timeEncoder" yaml:"timeEncoder"`
   EncodeDuration DurationEncoder `json:"durationEncoder" yaml:"durationEncoder"`
   EncodeCaller   CallerEncoder   `json:"callerEncoder" yaml:"callerEncoder"`
   // 日志器名称编码器
   EncodeName NameEncoder `json:"nameEncoder" yaml:"nameEncoder"`
   // 反射编码器，主要是对于interface{}类型，如果没有默认jsonencoder
   NewReflectedEncoder func(io.Writer) ReflectedEncoder `json:"-" yaml:"-"`
   // 控制台输出间隔字符串
   ConsoleSeparator string `json:"consoleSeparator" yaml:"consoleSeparator"`
}

举个例子：
func zapEncoder(config *ZapConfig) zapcore.Encoder {
   // 新建一个配置
   encoderConfig := zapcore.EncoderConfig{
      TimeKey:       "Time",
      LevelKey:      "Level",
      NameKey:       "Logger",
      CallerKey:     "Caller",
      MessageKey:    "Message",
      StacktraceKey: "StackTrace",
      LineEnding:    zapcore.DefaultLineEnding,
      FunctionKey:   zapcore.OmitKey,
   }
   // 自定义时间格式
   encoderConfig.EncodeTime = CustomTimeFormatEncoder
   // 日志级别大写
   encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
   // 秒级时间间隔
   encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
   // 简短的调用者输出
   encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
   // 完整的序列化logger名称
   encoderConfig.EncodeName = zapcore.FullNameEncoder
   // 最终的日志编码 json或者console
   switch config.Encode {
   case "json":
      {
         return zapcore.NewJSONEncoder(encoderConfig)
      }
   case "console":
      {
         return zapcore.NewConsoleEncoder(encoderConfig)
      }
   }
   // 默认console
   return zapcore.NewConsoleEncoder(encoderConfig)
}
```



### 三个核心中的日志输出

zapcore.AddSync 是添加的输出配置

```go
 func zapWriteSyncer(cfg *ZapConfig) zapcore.WriteSyncer {
   syncers := make([]zapcore.WriteSyncer, 0, 2)
   // 如果开启了日志控制台输出，就加入控制台书写器
   if cfg.Writer == config.WriteBoth || cfg.Writer == config.WriteConsole {
      syncers = append(syncers, zapcore.AddSync(os.Stdout))
   }

   // 如果开启了日志文件存储，就根据文件路径切片加入书写器
   if cfg.Writer == config.WriteBoth || cfg.Writer == config.WriteFile {
      // 添加日志输出器
      for _, path := range cfg.LogFile.Output {
         logger := &lumberjack.Logger{
            Filename:   path, //文件路径
            MaxSize:    cfg.LogFile.MaxSize, //分割文件的大小
            MaxBackups: cfg.LogFile.BackUps, //备份次数
            Compress:   cfg.LogFile.Compress, // 是否压缩
            LocalTime:  true, //使用本地时间
         }
         syncers = append(syncers, zapcore.Lock(zapcore.AddSync(logger)))
      }
   }
   return zap.CombineWriteSyncers(syncers...)
}
```



### 三个核心中的等级

```go
func zapLevelEnabler(cfg *ZapConfig) zapcore.LevelEnabler {
   switch cfg.Level {
   case config.DebugLevel:
      return zap.DebugLevel
   case config.InfoLevel:
      return zap.InfoLevel
   case config.ErrorLevel:
      return zap.ErrorLevel
   case config.PanicLevel:
      return zap.PanicLevel
   case config.FatalLevel:
      return zap.FatalLevel
   }
   // 默认Debug级别
   return zap.DebugLevel
}
```





### 最后融合

```go
func InitZap(config *ZapConfig) *zap.Logger {
   // 构建编码器
   encoder := zapEncoder(config)
   // 构建日志级别
   levelEnabler := zapLevelEnabler(config)
   // 最后获得Core和Options
   subCore, options := tee(config, encoder, levelEnabler)
    // 创建Logger
   return zap.New(subCore, options...)
}

// 将所有合并
func tee(cfg *ZapConfig, encoder zapcore.Encoder, levelEnabler zapcore.LevelEnabler) (core zapcore.Core, options []zap.Option) {
   sink := zapWriteSyncer(cfg)
   return zapcore.NewCore(encoder, sink, levelEnabler), buildOptions(cfg, levelEnabler)
}

// 构建Option
func buildOptions(cfg *ZapConfig, levelEnabler zapcore.LevelEnabler) (options []zap.Option) {
   if cfg.Caller {
      options = append(options, zap.AddCaller())
   }

   if cfg.StackTrace {
      options = append(options, zap.AddStacktrace(levelEnabler))
   }
   return
}
```



## 项目

三大核心实现主要在 zap.core 文件中，封装的有构造，特殊实现通个特殊构造完成。

因为ZapCore实现了zapcore.core所有方法，所以他是他的子类



### 1) 配置

- 文件：`configs/config.yaml`
- 类型：`internal/conf/conf.proto` → `message Log`

示例：

```yaml
log:
  director: "./logs"
  level: "debug"                 # debug|info|warn|error|dpanic|panic|fatal
  format: "json"                 # json|console
  stacktrace_key: "stacktrace"
  encode_level: "LowercaseLevelEncoder"
  log_in_console: false           # 仅写文件
  show_line: true                 # 输出调用位置
```

说明：

- `director`：日志根目录
- `level`：最低生效级别（包含其以上级别）
- `format`：输出格式
- `encode_level`：等级编码方式（console 下可配颜色）
- `log_in_console=false`：不输出控制台，仅写文件
- `show_line=true`：输出调用位置（文件:行号）

### 2) 初始化（已在 main.go 完成）

`cmd/conduit/main.go` 中已将 zap 适配为 Kratos 的 `log.Logger`，并注入到应用：

```go
kratosLogger := core.NewZapLoggerAdapter(zapLogger)
logger := log.With(kratosLogger,
    "service.id", id,
    "service.name", Name,
    "service.version", Version,
    "trace.id", tracing.TraceID(),
    "span.id", tracing.SpanID(),
)
```

### 3) 业务中如何使用

使用 `*log.Helper`（推荐）：

```go
// Info
s.log.Infow("msg",
    "event", "login_ok",
    "user_id", uid,
)

// Error
s.log.Errorw("msg",
    "err", err,
    "op", "CreateArticle",
)
```

使用 `log.Logger`：

```go
_ = logger.Log(log.LevelInfo, "msg", "boot", "version", Version)
_ = logger.Log(log.LevelError, "msg", "db connect failed", "err", err)
// 结构化键值对（推荐在全局 logger 场景，如 main/kafka 初始化）
_ = logger.Log(log.LevelInfo,
    "msg", "initializing kafka",
    "hosts", bc.Data.Kafka.Hosts,
    "topic", bc.Data.Kafka.Topic,
)
```

### 4) 按业务分目录写日志

携带以下任一键即可写入对应子目录（优先级：全部等价）：`service`、`biz`、`data`。

```go
// 写入 logs/YYYY-MM-DD/auth/all.log
s.log.Infow("msg", "service", "auth", "event", "login_ok", "user_id", uid)

// 写入 logs/YYYY-MM-DD/im/all.log
_ = logger.Log(log.LevelError, "msg", "delivery failed", "biz", "im", "err", err)

// 写入 logs/YYYY-MM-DD/user/all.log
_ = logger.Log(log.LevelDebug, "msg", "profile loaded", "data", "user", "user_id", uid)
```

支持多级目录：值可写为路径，如 `"service":"im/chat"` → `logs/YYYY-MM-DD/im/chat/all.log`。

### 5) 使用风格对比与建议

- 结构化（键值对）写法：`logger.Log(level, "key1", v1, "key2", v2, ...)`
  - 适合与全局中间件、管道日志统一；便于查询与统计
- 格式化写法：`helper.Infof("a=%s b=%s", a, b)`
  - 适合快速输出描述性文本；内部仍会走结构化适配

两者最终都会进入 zap，遵循同样的分流与落盘规则。

你用的 s.log.Infof/Infow/... 是 log.Helper，内部会调用注入的 log.Logger 的 Log(...)。
你注入的是 core.NewZapLoggerAdapter(...)，它实现了 log.Logger，内部用 zap.SugaredLogger 的 Infow/Errorw/... 实现。
最终就会走到 zap，并通过你自定义的 ZapCore 和 Cutter 落盘。
所以无论 Infof 还是 Infow，都会到底层 zap，并按你现在的文件分流规则写日志。

### 6) 输出格式与调用位置

- 切换 JSON/Console：`format`
- 颜色/大小写：`encode_level`（彩色仅 console 格式可见）
- 调用位置：`show_line=true` 时启用 `zap.AddCaller()`，并以完整路径输出（`FullCallerEncoder`）。



### 7) 常见问题

- 看不到文件：
  - `log_in_console` 是否为 `false`
  - 进程有写权限；工作目录是否正确
  - 至少打出一条日志；`level` 不要设太高
- 同时输出控制台+文件：需改 `WriteSyncer` 为 `zapcore.NewMultiWriteSyncer(os.Stdout, cutter)`（当前为二选一）。

### 8) 原理说明（调用链）

- 你在 `main.go` 中构建了 zap 并适配为 Kratos 日志：
  1. `zapLogger := core.Zap(bc.Log)` → 生成带自定义 `ZapCore` 的 zap.Logger
  2. `kratosLogger := core.NewZapLoggerAdapter(zapLogger)` → 适配成 Kratos 的 `log.Logger`
  3. `logger := log.With(kratosLogger, ...)` → 追加统一字段（service.id/name/version 等）
- 业务侧调用：
  - `log.Helper` 路径：`helper.Infof/Infow` → 调用底层 `log.Logger.Log(...)`
  - `log.Logger` 路径：直接 `logger.Log(level, keyvals...)`
- 适配器 `ZapLogger`：把 `log.Level` 映射为 zap 的级别，并调用 `zap.SugaredLogger` 的 `Debugw/Infow/Errorw...`
- zap 流水线：
  - `Check`/`Enabled` 决定级别是否通过
  - `Write` 中根据字段键 `service|biz|data` 决定写入哪个子目录（通过 `WriteSyncer` 返回 `Cutter`）
  - `Cutter.Write` 负责路径拼接、目录创建、文件打开并追加写入



s.log.Info("service started")
s.log.Infof("a is %s, b is %s", a, b)
s.log.Infow("msg", "service", "auth", "event", "login_ok", "user_id", uid)

s.log.Warn("slow query")
s.log.Warnf("slow query: %d ms", ms)
s.log.Warnw("msg", "sql", query, "elapsed_ms", ms)

s.log.Error("db connect failed")
s.log.Errorf("db connect failed: %v", err)
s.log.Errorw("msg", "err", err, "dsn", dsn)

// 固定字段
s2 := s.log.With("component", "profile")
s2.Infow("msg", "user_id", uid)


s.log 是 Kratos 的 log.Helper。Infof 只是一个便捷方法，内部会把格式化后的字符串转成一次 Logger.Log(...) 调用（级别=Info）。
你在 main.go 注入的底层 Logger 实际是 core.NewZapLoggerAdapter(zapLogger)，它实现了 Kratos 的 log.Logger 接口。
这个适配器在 Log(...) 里会把 Kratos 的级别映射到 zap，并调用 zap.SugaredLogger 的对应方法（如 Infow）。
zap 接着走你自定义的 ZapCore，由 WriteSyncer 返回的 Cutter 将日志写入文件。
所以链路是：Helper.Infof → Logger.Log(LevelInfo, ...) → ZapLoggerAdapter → zap → ZapCore → Cutter 落盘。
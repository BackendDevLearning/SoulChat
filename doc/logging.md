## 项目日志使用说明（zap + Kratos，文件输出）

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

### 3) 落盘路径规则

- 最终路径：`<director>/<YYYY-MM-DD>/<可选子目录>/all.log`
- 日期目录格式：`2006-01-02`（由 `CutterWithLayout(time.DateOnly)` 决定）
- 带上 `service`/`biz`/`data` 字段可切换到对应子目录。

相关代码：

```58:66:cmd/conduit/core/zap_core.go
func (z *ZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
    for i := 0; i < len(fields); i++ {
        if fields[i].Key == "service" || fields[i].Key == "biz" || fields[i].Key == "data" {
            syncer := z.WriteSyncer(fields[i].String)
            z.Core = zapcore.NewCore(z.DiyEncoder(), syncer, z.level)
        }
    }
    return z.Core.Write(entry, fields)
}
```

```64:74:cmd/conduit/core/cutter.go
values = append(values, c.director)
if c.layout != "" {
    values = append(values, time.Now().Format(c.layout))
}
for i := 0; i < length; i++ {
    values = append(values, c.formats[i])
}
values = append(values, c.level+".log")
filename := filepath.Join(values...)
```

### 4) 业务中如何使用

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

### 5) 按业务分目录写日志

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

### 5.1) 使用风格对比与建议

- 结构化（键值对）写法：`logger.Log(level, "key1", v1, "key2", v2, ...)`
  - 适合与全局中间件、管道日志统一；便于查询与统计
- 格式化写法：`helper.Infof("a=%s b=%s", a, b)`
  - 适合快速输出描述性文本；内部仍会走结构化适配

两者最终都会进入 zap，遵循同样的分流与落盘规则。

你用的 s.log.Infof/Infow/... 是 log.Helper，内部会调用注入的 log.Logger 的 Log(...)。
你注入的是 core.NewZapLoggerAdapter(...)，它实现了 log.Logger，内部用 zap.SugaredLogger 的 Infow/Errorw/... 实现。
最终就会走到 zap，并通过你自定义的 ZapCore 和 Cutter 落盘。
所以无论 Infof 还是 Infow，都会到底层 zap，并按你现在的文件分流规则写日志。

### 6) 级别选择策略

当前 `level` 为“阈值”语义：如 `info` 表示启用 `info, warn, error, dpanic, panic, fatal`。

如需“离散多级别”（例如仅 `info,error`），需扩展 `Levels` 支持逗号分隔（暂未启用，可按需修改 `cmd/conduit/core/zap_tool.go` 的 `Levels` 实现）。

### 7) 输出格式与调用位置

- 切换 JSON/Console：`format`
- 颜色/大小写：`encode_level`（彩色仅 console 格式可见）
- 调用位置：`show_line=true` 时启用 `zap.AddCaller()`，并以完整路径输出（`FullCallerEncoder`）。

### 8) 启动阶段验证落盘

```go
_ = logger.Log(log.LevelInfo, "msg", "service boot", "service", "system")
```

结果示例：`./logs/<YYYY-MM-DD>/system/all.log`

### 9) 常见问题

- 看不到文件：
  - `log_in_console` 是否为 `false`
  - 进程有写权限；工作目录是否正确
  - 至少打出一条日志；`level` 不要设太高
- 同时输出控制台+文件：需改 `WriteSyncer` 为 `zapcore.NewMultiWriteSyncer(os.Stdout, cutter)`（当前为二选一）。

### 10) 原理说明（调用链）

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
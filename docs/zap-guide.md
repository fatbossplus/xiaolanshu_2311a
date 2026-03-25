# zap 入门指南

## zap 是什么

zap 是 Uber 开源的 Go 高性能日志库，核心特点是**零分配**、**结构化字段**、**速度快**。
与标准库 `log` 不同，zap 输出的是结构化 JSON，每个字段都是独立的 key-value，方便日志平台检索。

---

## 从零开始：zap 的三层结构

理解 zap 只需要搞清楚三个东西是什么、怎么组合：

```
zapcore.Encoder   — 决定日志格式（JSON 还是 console 文本）
zapcore.WriteSyncer — 决定日志写到哪里（控制台、文件）
zapcore.Level     — 决定哪些级别的日志会被输出
        ↓
zapcore.Core      — 把上面三个组合在一起
        ↓
zap.Logger        — 最终使用的 logger，在 Core 基础上加选项
```

---

## 最简单的写法

```go
// zap 提供了两个开箱即用的预设，适合快速上手

// 1. 开发模式：console 格式，彩色输出，有 DEBUG
logger, _ := zap.NewDevelopment()

// 2. 生产模式：JSON 格式，INFO 及以上
logger, _ := zap.NewProduction()

defer logger.Sync() // 程序退出前刷新缓冲区

logger.Info("hello zap")
```

输出（production）：
```json
{"level":"info","ts":1711234567.89,"caller":"main.go:10","msg":"hello zap"}
```

---

## 手动组装（项目中的写法）

预设不够灵活，实际项目一般手动组装三层：

### 第一层：Encoder 格式

```go
// EncoderConfig 控制每个字段的格式
encCfg := zap.NewProductionEncoderConfig()

// 时间格式，默认是 Unix 时间戳，改成可读格式
encCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

// 日志级别格式：CapitalLevelEncoder → "INFO"，默认是 "info"
encCfg.EncodeLevel = zapcore.CapitalLevelEncoder

// 两种 Encoder：
enc := zapcore.NewJSONEncoder(encCfg)    // JSON 格式，生产环境用
enc := zapcore.NewConsoleEncoder(encCfg) // 文本格式，开发时用
```

### 第二层：WriteSyncer 输出目标

```go
// 输出到控制台
ws := zapcore.AddSync(os.Stdout)

// 输出到文件
f, _ := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
ws := zapcore.AddSync(f)

// 同时输出到控制台和文件
ws := zapcore.NewMultiWriteSyncer(
    zapcore.AddSync(os.Stdout),
    zapcore.AddSync(f),
)
```

### 第三层：Level 日志级别

```go
// 方式一：直接用常量
level := zapcore.InfoLevel  // DEBUG < INFO < WARN < ERROR < FATAL

// 方式二：从字符串解析（适合读配置文件）
var level zapcore.Level
_ = level.UnmarshalText([]byte("debug")) // 解析 "debug"/"info"/"warn"/"error"
```

### 组合成 Core，再建 Logger

```go
// 把三层组合成 Core
core := zapcore.NewCore(enc, ws, level)

// 用 Core 创建 Logger，后面是可选项
logger := zap.New(core,
    zap.AddCaller(),       // 记录调用文件和行号
    zap.AddCallerSkip(1),  // 调用层级跳过几层（封装时用）
    zap.Fields(           // 全局固定字段，每条日志都带
        zap.String("service", "gateway"),
    ),
)
```

---

## 写日志：两种风格

### 1. 结构化字段（推荐）

```go
logger.Info("用户登录",
    zap.String("username", "zhangsan"),  // 字符串
    zap.Int64("userID", 10086),          // 整数
    zap.Bool("isVip", true),             // 布尔
    zap.Duration("latency", time.Since(start)), // 时间段
    zap.Error(err),                      // error 类型
)
```

输出：
```json
{"level":"INFO","time":"2026-03-24 10:00:00","msg":"用户登录","service":"gateway","username":"zhangsan","userID":10086,"isVip":true,"latency":"3ms"}
```

### 2. Sugar 风格（类似 fmt.Printf）

```go
sugar := logger.Sugar()

sugar.Infof("用户 %s 登录", username)       // Printf 风格
sugar.Infow("用户登录", "username", "zhangsan") // key-value 风格
```

> 推荐使用结构化字段，Sugar 性能稍差，且字段类型不安全。

---

## AddCallerSkip 是什么

`zap.AddCaller()` 会记录**调用 logger 的那一行代码**的位置。
但当你把 logger 封装成全局函数时，caller 会指向封装层，而不是真正的业务代码。

```go
// 封装层：logger/logger.go 第 42 行
func Info(msg string, fields ...zap.Field) {
    log.Info(msg, fields...) // ← caller 会指向这里，不是业务代码
}

// 业务层：handler/auth.go 第 20 行
logger.Info("用户登录") // ← 我们希望 caller 指向这里
```

`zap.AddCallerSkip(1)` 的作用就是告诉 zap：**往上跳一层**，跳过封装层，指向真正的调用方。

```
调用栈：
  handler/auth.go:20   ← AddCallerSkip(1) 后指向这里 ✅
  logger/logger.go:42  ← AddCallerSkip(0) 默认指向这里 ❌
  zap 内部
```

如果封装了两层，就用 `AddCallerSkip(2)`，以此类推。

---

## 全局封装模式（本项目的写法）

```go
var log *zap.Logger // 包级全局变量

func Init(cfg LoggerConfig) {
    // ... 组装 logger
    log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), ...)
}

// 暴露全局函数，业务代码直接调用，不需要传 logger 对象
func Info(msg string, fields ...zap.Field)  { log.Info(msg, fields...) }
func Error(msg string, fields ...zap.Field) { log.Error(msg, fields...) }
```

业务代码使用：
```go
import "backend/pkg/logger"

logger.Info("订单创建", zap.Int64("orderID", 1001))
logger.Error("数据库查询失败", zap.Error(err))
```

---

## 派生 logger（With）

`With` 返回一个携带固定字段的新 logger，不影响原 logger：

```go
// 基础 logger
logger.Info("请求开始") // {"msg":"请求开始","service":"gateway"}

// 派生一个带 traceID 的 logger
reqLogger := log.With(zap.String("traceID", "abc123"))
reqLogger.Info("请求开始") // {"msg":"请求开始","service":"gateway","traceID":"abc123"}
reqLogger.Info("查询用户") // {"msg":"查询用户","service":"gateway","traceID":"abc123"}

// 原 logger 不受影响
logger.Info("其他日志") // {"msg":"其他日志","service":"gateway"}
```

本项目中 `InjectCtx` / `FromCtx` 就是用这个机制，把带 traceID 的派生 logger 存入 gin.Context，让同一请求的所有日志都自动带上 traceID。

---

## 常用 Field 速查

```go
zap.String("key", "value")
zap.Int("key", 1)
zap.Int64("key", int64(1))
zap.Float64("key", 3.14)
zap.Bool("key", true)
zap.Error(err)                        // 等同于 zap.NamedError("error", err)
zap.Duration("key", time.Second)
zap.Time("key", time.Now())
zap.Any("key", anyValue)              // 任意类型，性能较差，少用
```

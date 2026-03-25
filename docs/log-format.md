# 日志格式规范

## 企业级日志标准

企业级日志需要满足三个核心要求：**可读**（人工排查）、**可搜索**（日志平台检索）、**可关联**（链路追踪）。
统一格式是实现这三点的基础。

---

## 日志结构

每一条日志必须包含以下字段：

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `time` | string | 时间戳，精确到毫秒 | `2026-03-24 10:00:00.123` |
| `level` | string | 日志级别（大写） | `INFO` |
| `caller` | string | 调用位置，文件名:行号 | `handler/user.go:42` |
| `msg` | string | 日志主题，简短描述事件 | `user login` |
| `traceID` | string | 链路追踪 ID，全局唯一 | `1711234567890123456` |
| `service` | string | 服务名 | `gateway` |

业务字段按需附加在以上基础字段之后。

---

## 日志级别

| 级别 | 使用场景 |
|------|---------|
| `DEBUG` | 开发调试，详细的变量值、执行路径。**生产环境关闭** |
| `INFO` | 正常业务流程节点，如请求进出、用户登录、订单创建 |
| `WARN` | 未达到错误但需要关注，如接口慢响应、重试、降级 |
| `ERROR` | 业务异常、第三方调用失败、数据库错误，需要告警 |
| `FATAL` | 系统无法继续运行，记录后立即退出进程 |

> 原则：INFO 记录"发生了什么"，ERROR 记录"哪里出错了"，不要用 ERROR 记录业务上的"失败"（如密码错误），那应该是 INFO/WARN。

---

## 日志格式：JSON

生产环境统一使用 JSON 格式，便于 ELK、Loki 等日志平台解析和检索。

### 请求日志

```json
{
  "time": "2026-03-24 10:00:00.123",
  "level": "INFO",
  "caller": "middleware/logger.go:28",
  "msg": "request",
  "traceID": "1711234567890123456",
  "service": "gateway",
  "method": "POST",
  "path": "/v1/auth/login",
  "status": 200,
  "latency": "12ms",
  "ip": "112.64.20.1",
  "userAgent": "Mozilla/5.0"
}
```

### 业务日志

```json
{
  "time": "2026-03-24 10:00:00.118",
  "level": "INFO",
  "caller": "handler/auth.go:56",
  "msg": "user login",
  "traceID": "1711234567890123456",
  "service": "gateway",
  "userID": 10086,
  "username": "zhangsan"
}
```

### 错误日志

```json
{
  "time": "2026-03-24 10:00:00.200",
  "level": "ERROR",
  "caller": "handler/auth.go:61",
  "msg": "query user failed",
  "traceID": "1711234567890123456",
  "service": "gateway",
  "userID": 10086,
  "error": "dial tcp 127.0.0.1:3306: connect: connection refused",
  "stack": "..."
}
```

---

## 字段命名规范

- 统一使用 **camelCase**（小驼峰），如 `userID`、`traceID`、`requestBody`
- 布尔值加 `is` 前缀，如 `isLogin`、`isVip`
- 时间类字段加 `At` 后缀，如 `createdAt`、`expiredAt`
- ID 类字段统一大写 `ID`，如 `userID`、`orderID`，不写成 `userId`

---

## 禁止记录的内容

以下内容**严禁**出现在日志中，防止敏感信息泄露：

| 禁止字段 | 原因 |
|---------|------|
| 密码、密钥 | 安全合规 |
| 完整手机号、身份证号 | 用户隐私，需脱敏后记录 |
| 银行卡号、支付信息 | PCI-DSS 合规要求 |
| Token / Cookie 完整值 | 防止凭证泄露 |
| 完整请求体 | 可能包含以上任意内容 |

脱敏示例：`138****8888`、`id_card: 310***********1234`

---

## 代码规范

### 初始化（含 service 字段）

```go
// 建议在 logger.Init 时注入固定字段，之后每条日志自动携带
func Init(cfg config.LoggerConfig) {
    core := buildCore(cfg)
    log = zap.New(core,
        zap.AddCaller(),
        zap.AddCallerSkip(1),
        zap.Fields(zap.String("service", "gateway")), // 全局固定字段
    )
}
```

### 正确写法

```go
// ✅ 结构化字段，key-value 清晰，日志平台可直接检索
logger.Info("user login",
    zap.String("traceID", traceID),
    zap.Int64("userID", userID),
    zap.String("username", username),
)

// ✅ 错误日志带上 error 字段
logger.Error("query user failed",
    zap.String("traceID", traceID),
    zap.Int64("userID", userID),
    zap.Error(err),
)
```

### 错误写法

```go
// ❌ 字符串拼接，日志平台无法解析字段
logger.Info(fmt.Sprintf("user %d login", userID))

// ❌ 日志内容含敏感信息
logger.Info("user login", zap.String("password", password))

// ❌ 级别错误：密码错误是业务流程，不是系统错误
logger.Error("password incorrect")   // 应改为 Info 或 Warn
```

---

## 与 traceID 联动

同一请求的所有日志都带上相同的 `traceID`，在日志平台中按 traceID 过滤，
即可还原完整的请求链路：

```
traceID=1711234567890123456

10:00:00.100  INFO  --> POST /v1/auth/login          # 请求进入 Logger 中间件
10:00:00.101  INFO  trace start                       # 进入 Trace 中间件，生成 traceID
10:00:00.102  INFO  user login   userID=10086         # Handler 业务日志
10:00:00.115  INFO  trace end    status=200 12ms      # Trace 中间件出站
10:00:00.116  INFO  <-- POST /v1/auth/login 200 16ms  # Logger 中间件出站
```

# 洋葱模型

## 什么是洋葱模型

洋葱模型（Onion Model）是一种中间件执行模型。请求从最外层中间件进入，逐层向内传递，到达核心 Handler 执行后，再逐层向外返回。整个过程像剥洋葱一样，每一层都能在请求进来和响应出去两个时机执行逻辑。

```
请求 ──────────────────────────────── 响应
     │ Logger 入站                    │ Logger 出站
     │   │ Trace 入站                 │ Trace 出站 │
     │   │   │ Auth 入站              │ Auth 出站 │ │
     │   │   │   │  Handler          │          │ │ │
     └───┴───┴───┴───────────────────┴──────────┴─┴─┘
```

**执行顺序：**

1. Logger 入站
2. Trace 入站
3. Auth 入站
4. Handler 执行
5. Auth 出站
6. Trace 出站
7. Logger 出站

> 每一层都"包裹"着内层，入站从外到内，出站从内到外，形成对称结构。

---

## 为什么使用洋葱模型

| 优势 | 说明 |
|------|------|
| 精准计时 | 在入站记录开始时间，出站计算总耗时，覆盖完整的请求生命周期 |
| 链路追踪 | 入站生成 traceID，出站记录最终状态，全程贯穿 |
| 异常捕获 | `c.Next()` 后可拿到 handler 的执行结果，集中处理错误和状态码 |
| 职责清晰 | 每个中间件只关心自己的入站和出站逻辑，互不干扰 |
| 执行可控 | 可在入站阶段拦截请求（如鉴权失败直接返回，不调用 `c.Next()`） |

**与线性模型的区别：**

```
线性模型：A → B → C → Handler        只有一个方向，无法在响应阶段介入
洋葱模型：A → B → C → Handler → C → B → A    来回两个方向都能介入
```

---

## 代码实现

### 核心结构

`c.Next()` 是洋葱模型的分界线，调用它会立即跳转去执行后续所有中间件和 Handler，
等它们全部执行完毕后，才回到当前位置继续向下执行出站逻辑：

```go
func XxxMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        // ========== 入站逻辑 ==========
        // 此时请求刚进来，后续 Handler 还未执行
        // 适合做：记录开始时间、生成 traceID、鉴权拦截等

        c.Next() // 执行后续所有中间件 + Handler，执行完后回到这里

        // ========== 出站逻辑 ==========
        // 此时 Handler 已经执行完，响应已经写入
        // 可以拿到：c.Writer.Status()（响应状态码）、耗时等
        // 适合做：记录响应日志、统计耗时、错误收集等
    }
}
```

> **注意**：如果入站阶段调用了 `c.Abort()`，会阻止后续中间件和 Handler 执行，
> 但当前中间件 `c.Next()` 之后的出站逻辑仍然会执行。

---

### 日志中间件

```go
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        // 【入站】在请求进来时记录开始时间
        // 此时 Handler 还未执行，无法拿到状态码，只记录请求基本信息
        start := time.Now()
        logger.Info("-->",
            zap.String("method", c.Request.Method), // HTTP 方法：GET/POST 等
            zap.String("path", c.Request.URL.Path), // 请求路径
            zap.String("ip", c.ClientIP()),          // 客户端 IP
        )

        c.Next() // 等待 Handler 执行完毕

        // 【出站】Handler 已执行完，此时可以拿到响应状态码
        // time.Since(start) 计算的是从入站到出站的完整耗时，包含 Handler 执行时间
        logger.Info("<--",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),       // 响应状态码，Handler 写入后才能读取
            zap.Duration("latency", time.Since(start)), // 完整耗时
        )
    }
}
```

---

### 链路追踪中间件

链路追踪的核心是给每个请求分配一个唯一 ID（traceID），让这个 ID 贯穿整个请求生命周期，
方便在日志中将同一请求的所有日志关联起来排查问题。

```go
func TraceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        // 【入站】生成或继承 traceID
        // 优先读请求头中的 X-Trace-Id：
        //   - 如果客户端或上游服务已经设置了，就复用它（保证跨服务链路一致）
        //   - 如果没有，则在本服务生成一个新的
        traceID := c.GetHeader("X-Trace-Id")
        if traceID == "" {
            traceID = genTraceID() // 生成唯一 ID
        }

        // 将 traceID 写入响应头，客户端可以凭此 ID 到日志系统查询完整链路
        c.Header("X-Trace-Id", traceID)

        // 将 traceID 存入 gin.Context，后续 Handler 可通过 c.GetString("X-Trace-Id") 取用
        // 例如：在数据库查询、RPC 调用时携带 traceID，实现全链路追踪
        c.Set("X-Trace-Id", traceID)

        start := time.Now()
        logger.Info("trace start",
            zap.String("traceID", traceID),
            zap.String("path", c.Request.URL.Path),
        )

        c.Next() // 等待 Handler 执行完毕

        // 【出站】记录本次链路的最终结果
        // 此时可以看到请求最终的状态码，结合 traceID 可以在日志中过滤出完整的请求链路
        logger.Info("trace end",
            zap.String("traceID", traceID),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("latency", time.Since(start)),
        )
    }
}
```

---

### 注册顺序

```go
v1 := r.Group("/v1")

// 越先注册越靠外层，入站最先执行，出站最后执行
// Logger 放最外层：保证统计的耗时包含所有中间件的执行时间
v1.Use(middleware.LoggerMiddleware())

// Trace 紧跟 Logger：尽早生成 traceID，让后续所有中间件的日志都能带上 traceID
v1.Use(middleware.TraceMiddleware())

// 限流在鉴权之前：被限流的请求不需要再做鉴权，节省资源
v1.Use(middleware.RateLimitMiddleware())

// 权限和鉴权放内层：越靠近 Handler 越合理，鉴权失败直接 Abort 不再继续
v1.Use(middleware.PermissionMiddleware())
v1.Use(middleware.AuthMiddleware()) // 最内层，最后入站，最先出站
```

> **原则**：需要包裹完整生命周期的（计时、追踪）放外层；越早能拦截越好的（限流、鉴权）放内层。

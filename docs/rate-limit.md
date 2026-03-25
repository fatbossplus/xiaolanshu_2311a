# 限流

## 什么是限流

限流是保护服务的一道防线，目的是控制单位时间内允许通过的请求数量，防止流量突增压垮服务。

常见场景：
- 用户高频刷接口（恶意/误操作）
- 爬虫批量请求
- 活动期间流量突增

---

## 令牌桶算法

本项目使用**令牌桶算法**，是最常用的限流算法之一。

### 原理

```
                  ┌─────────────────┐
  每秒生成 100 个  │  令牌桶（最多   │
  令牌 ────────►  │   存 200 个）   │
                  └────────┬────────┘
                           │ 每个请求消耗 1 个令牌
                           ▼
                  有令牌 → 请求通过 ✅
                  无令牌 → 请求拒绝 ❌ 429
```

- **rps（rate）**：令牌生成速率，每秒补充 100 个令牌，决定**平均速率上限**
- **burst（桶容量）**：最多积累 200 个令牌，决定**突发流量上限**
- 令牌持续补充，桶满后不再积累
- 每个请求消耗一个令牌，桶空则拒绝

### 与漏桶算法的区别

| | 令牌桶 | 漏桶 |
|--|--------|------|
| 突发流量 | 允许（桶内有积累的令牌） | 不允许（固定速率流出） |
| 适合场景 | 接口限流，允许合理突发 | 流量整形，严格匀速 |

令牌桶更适合接口限流，因为真实用户访问本身就存在突发性。

---

## 代码实现

### 核心依赖

```
golang.org/x/time/rate  — Go 官方令牌桶实现
sync.Map                — 并发安全的 map，存储每个 IP 的限流器
```

### 关键参数

```go
rps   = rate.Limit(100) // 每秒生成 100 个令牌（平均速率）
burst = 200             // 桶容量 200（允许最大突发）
```

### 按 IP 限流

每个 IP 拥有独立的令牌桶，互不影响：

```go
var limiters sync.Map // key=IP, value=*rate.Limiter

func getLimiter(ip string) *rate.Limiter {
    // LoadOrStore：存在则返回已有的，不存在则创建新的
    // 保证同一 IP 始终使用同一个限流器，并发安全
    v, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rps, burst))
    return v.(*rate.Limiter)
}
```

### Allow / Wait / Reserve 的区别

`rate.Limiter` 提供三种消费令牌的方式：

| 方法 | 行为 | 适合场景 |
|------|------|---------|
| `Allow()` | 立即返回 true/false，不等待 | HTTP 接口限流（本项目） |
| `Wait(ctx)` | 阻塞等待直到有令牌或 ctx 超时 | 后台任务、队列消费 |
| `Reserve()` | 返回预约信息，可获知需等待多久 | 需要告知客户端重试时间 |

本项目使用 `Allow()`，请求到来时立即判断，超限直接返回 `429`，不阻塞 goroutine。

### 中间件实现

```go
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        if !getLimiter(ip).Allow() {
            // 超限：立即拒绝，不进入后续中间件和 handler
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "请求过于频繁"})
            return
        }
        c.Next()
    }
}
```

---

## 注册位置与顺序

```go
v1.Use(middleware.CORSMiddleware())
v1.Use(middleware.LoggerMiddleware())
v1.Use(middleware.TraceMiddleware())
v1.Use(middleware.RateLimitMiddleware()) // ← 限流在鉴权之前
v1.Use(middleware.PermissionMiddleware())
v1.Use(middleware.AuthMiddleware())
```

**为什么限流要放在鉴权之前？**

被限流的请求直接拒绝，不需要再查数据库验 token，节省资源。
鉴权通常涉及 DB 或 Redis 查询，让超限请求进入鉴权是一种浪费。

---

## 参数调优参考

| 场景 | rps | burst |
|------|-----|-------|
| 普通接口 | 100 | 200 |
| 登录/注册（防爆破） | 5 | 10 |
| 开放 API | 1000 | 2000 |
| 文件上传 | 10 | 20 |

> 敏感接口（登录、发短信）应单独设置更严格的限流，防止暴力破解。

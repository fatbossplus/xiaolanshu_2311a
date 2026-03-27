# 小蓝书后端 - 项目结构说明

## 目录结构

```
backend/
├── config.yaml                # 全局配置文件
├── go.mod / go.sum
│
├── pkg/                       # 公共工具包
│   ├── logger/                # 日志（zap）
│   ├── config/                # 配置加载
│   ├── database/              # 数据库
│   ├── redis/                 # Redis
│   ├── cache/                 # 缓存
│   ├── kafka/                 # 消息队列
│   ├── es/                    # Elasticsearch
│   ├── jwt/                   # JWT 鉴权
│   ├── middleware/            # 中间件
│   ├── response/              # 响应封装
│   ├── snowflake/             # 雪花ID
│   ├── bloom/                 # 布隆过滤器
│   └── singleflight/          # 防缓存击穿
│
├── proto/                     # gRPC 协议定义
│
├── gateway/                   # API 网关层（HTTP 入口）
│   ├── main.go
│   ├── router/                # 路由注册
│   │   ├── router.go          # 总路由
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── note.go
│   │   ├── comment.go
│   │   ├── social.go
│   │   ├── recommend.go
│   │   ├── search.go
│   │   ├── notify.go
│   │   ├── message.go
│   │   ├── upload.go
│   │   └── admin.go
│   ├── internal/handler/      # HTTP Handler（调用 RPC）
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── note_handler.go
│   │   ├── social_handler.go
│   │   ├── recommend_handler.go
│   │   ├── search_handler.go
│   │   ├── notify_handler.go
│   │   ├── upload_handler.go
│   │   └── admin_handler.go
│   ├── rpc/
│   │   └── clients.go         # gRPC 客户端连接池
│   └── static/uploads/        # 上传的图片文件
│
└── services/                  # 微服务层（gRPC 服务端）
    ├── user/                  # 用户服务
    │   ├── main.go
    │   ├── server/user_server.go
    │   └── internal/
    │       ├── handler/user_handler.go
    │       ├── service/user_service.go
    │       ├── repository/user_repo.go
    │       └── model/user.go
    ├── content/               # 内容（笔记）服务
    │   ├── main.go
    │   ├── server/content_server.go
    │   └── internal/
    │       ├── handler/note_handler.go
    │       ├── service/note_service.go
    │       ├── repository/note_repo.go
    │       └── model/note.go
    ├── social/                # 社交服务（关注/点赞/收藏）
    │   ├── main.go
    │   ├── server/social_server.go
    │   └── internal/
    │       ├── handler/social_handler.go
    │       ├── service/social_service.go
    │       ├── repository/social_repo.go
    │       └── model/social.go
    ├── recommend/             # 推荐服务
    │   ├── main.go
    │   ├── server/recommend_server.go
    │   └── internal/
    │       ├── handler/recommend_handler.go
    │       ├── service/recommend_service.go
    │       ├── repository/recommend_repo.go
    │       └── model/recommend.go
    ├── search/                # 搜索服务
    │   ├── main.go
    │   ├── server/search_server.go
    │   └── internal/
    │       ├── handler/search_handler.go
    │       ├── service/search_service.go
    │       ├── repository/search_repo.go
    │       └── model/search.go
    ├── notify/                # 通知服务（含 WebSocket）
    │   ├── main.go
    │   ├── server/notify_server.go
    │   └── internal/
    │       ├── handler/notify_handler.go
    │       ├── service/notify_service.go
    │       ├── repository/notify_repo.go
    │       ├── model/notification.go
    │       └── websocket/hub.go
    └── audit/                 # 审核服务
        ├── main.go
        ├── server/audit_server.go
        └── internal/
            ├── handler/audit_handler.go
            ├── service/audit_service.go
            ├── repository/audit_repo.go
            └── model/audit.go
```

## 架构分层说明

整个后端分为两大部分：**网关层（gateway）** 和 **服务层（services）**，
它们之间通过 **gRPC** 通信，而不是普通的 HTTP 调用。

> **gRPC 是什么？** 可以简单理解为：服务之间互相调用函数的一种方式，
> 比 HTTP 更快、更规范，调用远程服务就像调用本地函数一样。

```
前端发来的 HTTP 请求
        ↓
  gateway/router         → 第一步：根据 URL 路径找到对应的处理函数（路由分发）
        ↓
  gateway/handler        → 第二步：解析请求参数，决定调用哪个微服务
        ↓
  gateway/rpc/clients    → 第三步：通过 gRPC 向对应微服务发起调用
        ↓
  services/xxx/server    → 第四步：微服务收到 gRPC 请求，进入服务内部处理
        ↓
  internal/handler       → 第五步：流程编排，决定先做什么、再做什么
        ↓
  internal/service       → 第六步：执行核心业务逻辑（最重要的一层）
        ↓
  internal/repository    → 第七步：读写数据库 / Redis / ES
        ↓
  MySQL / Redis / ES     → 最终数据存储
```

### 微服务内部四层结构

每个微服务（user、content、social 等）内部都采用相同的四层结构。
统一结构的好处是：**任何人打开任何一个服务，都知道去哪里找对应的代码**。

> 详细说明请参阅：[微服务四层架构设计文档](./docs/microservice-layer-design.md)

#### 快速理解四层（用"餐厅"来类比）

| 层级         | 文件位置                        | 职责说明                                           | 餐厅类比         |
|--------------|---------------------------------|----------------------------------------------------|------------------|
| `server`     | `server/xxx_server.go`          | gRPC 协议入口，接收请求并转交，不处理业务          | 前台服务员       |
| `handler`    | `internal/handler/xxx_handler.go` | 编排业务流程，决定调用哪些 service、顺序是什么   | 店长（指挥）     |
| `service`    | `internal/service/xxx_service.go` | 核心业务规则，例如"点赞不能重复"、"发帖需审核"  | 厨师（做事的）   |
| `repository` | `internal/repository/xxx_repo.go` | 数据库读写封装，只管存取数据，不管业务逻辑       | 仓库管理员       |

#### 一个请求的完整流程示例

以"用户点赞一篇笔记"为例：

```
① server      收到 gRPC 请求 LikeNote(userID=1, noteID=100)
      ↓
② handler     检查参数，调用 service.LikeNote()
      ↓
③ service     业务判断：
              - 查询该用户是否已点赞过（不能重复）
              - 写入点赞记录
              - 更新笔记点赞数 +1
              - 发送 Kafka 消息，通知推荐系统
      ↓
④ repository  执行 SQL：INSERT INTO likes ... / UPDATE notes SET like_count=...
      ↓
      返回成功
```

**为什么不把所有逻辑写在一个文件里？**
因为一旦混在一起，改一个地方可能导致其他地方出错，
而且代码越来越长，根本找不到问题在哪。
四层结构让每一层各司其职，**改数据库只改 repository，改业务规则只改 service**，
互不干扰。

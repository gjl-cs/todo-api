# ShortLink

一个基于 Go 构建的高并发短链接服务，支持短链接创建、302 重定向、Redis 缓存、缓存击穿防护、访问统计、限流以及 Prometheus + Grafana 可观测性。

项目重点围绕高并发读场景进行设计，并通过 k6 进行压力测试。

## 项目简介

ShortLink 是一个面向高并发访问场景设计的短链接服务。

用户提交一个长 URL 后，系统生成唯一短码，例如：

```text
https://example.com/very-long-url
        ↓
http://localhost:8080/uDJkH7
```

访问短链接时：

```text
Client
  ↓
Gin
  ↓
Redis Cache
  ↓
Cache Hit → 302 Redirect
  ↓
Cache Miss
  ↓
MySQL
  ↓
写入 Redis
  ↓
302 Redirect
```

项目重点解决了高并发短链接访问中的缓存、缓存击穿、数据库连接管理、限流和可观测性问题。

---

## 技术栈

- Go
- Gin
- MySQL 8
- Redis 7
- go-redis/v9
- Prometheus
- Grafana
- k6
- Docker
- MySQL Connection Pool
- Singleflight

---

## 核心功能

### 1. 短链接创建

```http
POST /api/v1/short-links
```

系统生成 6 位短码，并将短码与长 URL 持久化到 MySQL。

短码具有唯一索引：

```sql
UNIQUE INDEX uk_code(code)
```

当生成短码发生唯一键冲突时，服务会自动重新生成，最多重试 5 次。

---

### 2. 短链接重定向

```http
GET /:code
```

访问短码后返回：

```http
HTTP/1.1 302 Found
Location: https://example.com
```

核心访问路径采用 Redis Cache-Aside：

```text
Redis
  │
  ├── Hit ───────→ 直接 Redirect
  │
  └── Miss
        ↓
      MySQL
        ↓
    写入 Redis
        ↓
      Redirect
```

高频短链接访问不会持续查询 MySQL。

---

### 3. Redis 缓存

使用 Redis 缓存：

```text
shortlink:{code}
```

缓存默认设置 TTL，避免缓存永久存在。

同时维护访问量：

```text
shortlink:pv:{code}
```

通过 Redis 原子操作统计 PV。

---

### 4. Negative Cache

对于不存在的短码，也进行短时间缓存。

例如：

```text
shortlink:notfound:{code}
```

这样大量请求不存在的短码时，不会每次都访问 MySQL。

访问路径：

```text
请求不存在短码
      ↓
Redis Negative Cache
      ↓
命中
      ↓
直接返回 Not Found
```

降低恶意请求或热点不存在 Key 对数据库造成的压力。

---

### 5. Singleflight 防缓存击穿

当一个热门短链接缓存同时失效时，大量请求可能同时访问 MySQL。

项目使用 `singleflight` 合并并发请求：

```text
             ┌─ Request 1 ─┐
             ├─ Request 2 ─┤
Cache Miss → ├─ Request 3 ─┤
             ├─ Request 4 ─┤
             └─ Request N ─┘
                    ↓
              Singleflight
                    ↓
              一次 MySQL 查询
                    ↓
               写入 Redis
                    ↓
             所有请求共享结果
```

避免缓存击穿情况下的大量重复数据库查询。

---

### 6. MySQL 连接池

数据库连接池进行了显式配置：

```text
MaxOpenConns: 100
MaxIdleConns: 20
ConnMaxLifetime: 1h
```

避免高并发请求下数据库连接管理失控。

---

### 7. 限流

服务提供请求限流能力。

当系统达到保护阈值时返回：

```http
429 Too Many Requests
```

避免突发流量持续压垮后端服务。

---

### 8. 统一错误处理

服务包含统一错误响应格式，并通过 Gin Middleware 处理异常。

正常响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

异常请求返回对应 HTTP 状态码和统一错误结构。

同时提供 Panic Recovery，避免单个请求异常导致整个服务退出。

---

### 9. Prometheus + Grafana

服务集成 Prometheus Metrics。

可以监控：

- HTTP 请求数量
- 请求耗时
- 请求状态
- Redirect 请求
- 缓存命中/未命中
- PV
- 限流请求
- 服务运行状态

Grafana 用于可视化展示指标。

---

## 系统架构

```text
                    ┌──────────────┐
                    │    Client    │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │     Gin      │
                    │   Router     │
                    └──────┬───────┘
                           │
              ┌────────────┴────────────┐
              │                         │
              ▼                         ▼
       ┌──────────────┐         ┌──────────────┐
       │   Middleware │         │    Handler   │
       │ Rate Limit   │         │ Redirect/API │
       │ Metrics      │         └──────┬───────┘
       └──────────────┘                │
                                      ▼
                              ┌──────────────┐
                              │   Service    │
                              │ Singleflight │
                              └──────┬───────┘
                                     │
                         ┌───────────┴───────────┐
                         │                       │
                         ▼                       ▼
                  ┌──────────────┐       ┌──────────────┐
                  │    Redis     │       │    MySQL     │
                  │    Cache     │       │ Persistence  │
                  └──────────────┘       └──────────────┘


                  ┌──────────────┐
                  │  Prometheus  │
                  └──────┬───────┘
                         ▼
                  ┌──────────────┐
                  │   Grafana    │
                  └──────────────┘
```

---

## 项目结构

```text
shortlink/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── cache/
│   │   └── redis.go
│   │
│   ├── handler/
│   │   ├── health.go
│   │   ├── short_link.go
│   │   └── *_test.go
│   │
│   ├── metrics/
│   │   └── metrics.go
│   │
│   ├── middleware/
│   │   ├── error.go
│   │   ├── logger.go
│   │   ├── metrics.go
│   │   ├── rate_limit.go
│   │   └── *_test.go
│   │
│   ├── repository/
│   │   ├── db.go
│   │   └── short_link.go
│   │
│   ├── response/
│   │   └── response.go
│   │
│   ├── router/
│   │   └── router.go
│   │
│   ├── service/
│   │   ├── short_link.go
│   │   ├── singleflight_test.go
│   │   └── *_test.go
│   │
│   └── utils/
│       ├── short_code.go
│       ├── url.go
│       └── *_test.go
│
├── deploy/
│   └── prometheus/
│       └── prometheus.yml
│
├── reports/
│   └── k6 test reports
│
├── scripts/
│   ├── cache-hit-test.js
│   ├── load-test.js
│   ├── redirect-load-test.js
│   ├── redirect-stress-test.js
│   └── run-tests.ps1
│
├── go.mod
├── go.sum
└── README.md
```

---

## API

### 创建短链接

```http
POST /api/v1/short-links
Content-Type: application/json
```

请求：

```json
{
  "long_url": "https://example.com"
}
```

---

### 短链接跳转

```http
GET /:code
```

例如：

```http
GET /uDJkH7
```

成功返回：

```http
302 Found
Location: https://example.com
```

---

### 健康检查

```http
GET /health
```

---

## 本地运行

### 1. 启动 Redis

Redis 使用 Docker：

```powershell
docker run -d `
  --name shortlink-redis `
  -p 6379:6379 `
  redis:7
```

如果容器已经存在：

```powershell
docker start shortlink-redis
```

### 2. 配置环境变量

项目通过 `.env` 配置数据库和 Redis。

示例：

```text
DB_USER=...
DB_PASSWORD=...
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=shortlink

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
```

> `.env` 不应提交到 Git。

### 3. 启动服务

```powershell
go run ./cmd/server
```

服务默认监听：

```text
http://localhost:8080
```

---

## 测试

执行全部 Go 测试：

```powershell
go test ./...
```

当前项目测试全部通过。

---

## 高并发压测

使用 k6 对 Redirect 接口进行压力测试。

测试短码：

```text
uDJkH7
```

正式压测：

```powershell
$env:VUS="400"
$env:DURATION="30s"
$env:SHORT_CODE="uDJkH7"

k6 run .\scripts\redirect-stress-test.js
```

### 400 VU / 30s 测试结果

| 指标 | 结果 |
|---|---:|
| Virtual Users | 400 |
| Duration | 30s |
| Total Requests | 75,179 |
| Throughput | 2,502.76 req/s |
| Redirect Successes | 74,926 |
| Unexpected Status | 253 |
| HTTP Failed | 0.33% |
| P95 | 217.88 ms |
| P90 | 193.41 ms |
| Max | 593.97 ms |

其中异常请求主要表现为本机压测环境下的：

```text
dial: connection refused
```

并非业务接口返回 4xx/5xx。

在 50 VU 测试中，服务能够稳定完成约 2,300 req/s 的 Redirect 请求，且测试结果为 100% 成功。

---

## 性能优化思路

项目针对短链接典型的“高读低写”场景进行了优化：

### Redis Cache-Aside

减少热点短链接对 MySQL 的访问。

### Negative Cache

避免不存在 Key 被反复查询数据库。

### Singleflight

避免热点 Key 缓存失效时产生大量重复数据库查询。

### MySQL Connection Pool

通过连接池限制数据库连接数量，提高高并发场景下的稳定性。

### Redis PV Counter

使用 Redis 原子操作统计访问量，避免每次 Redirect 都更新 MySQL。

### Rate Limit

在入口处限制突发流量，保护后端服务。

### Metrics

通过 Prometheus + Grafana 对服务运行状态进行持续观测。

---

## 项目亮点

- 基于 Go + Gin 构建高并发短链接服务
- 使用 Redis Cache-Aside 降低 MySQL 查询压力
- 使用 Negative Cache 防止不存在 Key 穿透数据库
- 使用 Singleflight 防止热点 Key 缓存击穿
- 使用 Redis 原子操作实现高并发 PV 统计
- 配置 MySQL Connection Pool，提高数据库并发处理能力
- 增加 Rate Limit，保护服务稳定性
- 集成 Prometheus + Grafana，实现服务指标监控
- 使用 k6 完成 400 VU 高并发压力测试
- 完善单元测试、并发测试和 Benchmark
- 使用统一 Response 与 Middleware 进行工程化封装

---

## 后续优化方向

如果继续向生产级系统演进，可以进一步增加：

- 分布式限流
- Redis Cluster
- MySQL 读写分离
- 分布式 ID / 更高效短码生成策略
- 异步 PV 持久化
- Kafka 消息队列
- 服务水平扩展
- Docker Compose / Kubernetes 部署
- CI/CD 自动化测试与部署

---

## 项目定位

本项目主要用于学习和实践 Go 后端工程、高并发服务设计、缓存架构、数据库优化以及服务可观测性。

核心目标不是单纯实现“短链接生成”，而是围绕高并发 Redirect 场景，完整实践：

```text
API
 ↓
Middleware
 ↓
Service
 ↓
Cache
 ↓
Database
 ↓
Metrics
 ↓
Load Test
```

形成一个完整的 Go 后端项目工程闭环。
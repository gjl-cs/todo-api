# Todo API

一个使用 Go + Gin + SQLite 开发的 Todo REST API 项目。

## 技术栈

- Go
- Gin
- SQLite
- database/sql
- Swagger
- RESTful API

## 项目功能

- Todo 创建
- Todo 查询
- Todo 修改
- Todo 删除
- Todo 完成状态修改
- Todo 搜索
- 分页
- 优先级排序
- 截止日期排序
- 完成状态筛选
- 参数校验
- 统一 API 响应
- 错误处理
- Logger Middleware
- SQLite 数据库
- Swagger API 文档

## 项目结构

```text
todo/
├── main.go
├── go.mod
├── go.sum
├── model/
├── service/
├── storage/
├── web/
├── middleware/
├── utils/
└── docs/
```

## 运行项目

确保已经安装 Go。

进入项目目录：

```bash
cd todo
```

安装依赖：

```bash
go mod tidy
```

启动项目：

```bash
go run .
```

服务器启动后：

```text
http://localhost:8080
```

## Swagger

启动项目后访问：

```text
http://localhost:8080/swagger/index.html
```

可以通过 Swagger 查看和测试 API。

## 查询 Todo

### 查询全部

```http
GET /todos
```

### 搜索

```http
GET /todos?keyword=Go
```

### 分页

```http
GET /todos?page=1&pageSize=10
```

### 优先级排序

```http
GET /todos?sort=priority_desc
```

支持：

```text
priority_asc
priority_desc
```

### 截止日期排序

```http
GET /todos?sort=due_date_asc
```

支持：

```text
due_date_asc
due_date_desc
```

### 完成状态筛选

```http
GET /todos?done=true
```

或者：

```http
GET /todos?done=false
```

### 组合查询

```http
GET /todos?keyword=Go&done=false&sort=priority_desc&page=1&pageSize=10
```

## 创建 Todo

```http
POST /todos
```

请求示例：

```json
{
  "title": "学习 Go",
  "done": false,
  "priority": 3,
  "due_date": "2026-10-01"
}
```

其中：

- `title`：1～100 个字符
- `priority`：1～5
- `due_date`：`YYYY-MM-DD` 格式

## 项目学习目标

通过这个项目学习 Go 后端开发中的：

- HTTP
- REST API
- Gin
- JSON
- Handler / Service / Storage 分层
- SQL
- SQLite
- 分页
- 排序
- 动态查询
- 参数校验
- Middleware
- 错误处理
- Swagger API 文档

## 后续计划

- Docker
- 项目部署
- 单元测试
- API 测试
- 项目进一步工程化
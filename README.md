# Fool Blog

**一个简洁优雅的个人博客系统**

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-1.12-00ADD8?logo=go&logoColor=white)
![GORM](https://img.shields.io/badge/GORM-1.31-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

一个轻量级、功能完整的博客系统，基于 Go 构建，支持 SQLite/MySQL 双数据库，开箱即用。

---

## 功能特性

### 前台功能
- **文章展示** - 置顶文章、分页列表、Markdown 渲染、目录导航
- **互动功能** - 浏览量统计、点赞、评论（登录后）
- **内容发现** - 全文搜索（标题/摘要/正文/标签）、标签分类、按年月归档
- **AI 点评** - 接入 LLM 自动生成文章点评
- **个性化** - 随机壁纸、站点配置

### 后台管理
- **数据仪表盘** - 文章/用户/评论统计概览
- **文章管理** - 新建/编辑/删除/置顶/发布/草稿、Markdown 导入导出（单篇/批量 ZIP）
- **内容管理** - 标签管理、评论管理
- **用户管理** - 角色分配、状态控制、禁用/删除
- **站点设置** - 名称/副标题/公告/壁纸/备案号/建站时间
- **操作日志** - 全操作审计记录

### 用户系统
- 注册（邮箱验证码验证）
- 登录/登出
- 个人中心（修改资料/随机头像）

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://github.com/go-gorm/gorm) |
| 数据库 | [SQLite](https://github.com/glebarez/sqlite) / [MySQL](https://github.com/go-sql-driver/mysql) |
| 模板引擎 | Go `html/template` |
| Markdown | [Goldmark](https://github.com/yuin/goldmark) |
| 代码高亮 | [Chroma](https://github.com/alecthomas/chroma) |
| AI 集成 | 兼容 OpenAI API 的 LLM 服务 |

## 快速开始

### 环境要求

- Go 1.25+

### 安装

```bash
git clone https://github.com/your-username/fool-blog.git
cd fool-blog
go mod tidy
```

### 配置

复制并编辑配置文件：

```bash
cp config.yaml config.yaml
```

编辑 `config.yaml`：

```yaml
app:
  addr: 0.0.0.0:9200
  name: My Blog

database:
  driver: sqlite          # sqlite 或 mysql
  path: data/blog.db      # SQLite 路径
  # MySQL 配置（driver 为 mysql 时启用）
  # host: 127.0.0.1
  # port: "3306"
  # name: blog
  # user: root
  # password: ""

security:
  session_key: your-random-session-key-here
  password_salt: your-password-salt

admin:
  default_username: admin
  default_password: admin123

# 可选：AI 文章点评
ai:
  base_url: https://api.openai.com/v1
  api_key: sk-your-api-key
  model: gpt-4o-mini
```

### 运行

```bash
go run main.go
```

访问 `http://localhost:9200` 查看博客，`http://localhost:9200/admin` 进入后台。

默认管理员账号：`admin` / `admin123`

### 编译

```bash
# Linux/macOS
go build -o fool_blog

# Windows
go build -o fool_blog.exe
```

## 目录结构

```
fool-blog/
├── main.go                 # 应用入口
├── config.yaml             # 配置文件
├── go.mod                  # Go 依赖
├── internal/
│   ├── ai/                 # AI 点评集成
│   ├── config/             # 配置加载
│   ├── database/           # 数据库初始化
│   ├── email/              # 邮件服务
│   ├── handler/            # HTTP 路由处理
│   │   ├── admin_handler.go
│   │   ├── article_handler.go
│   │   ├── auth_handler.go
│   │   └── front_handler.go
│   ├── middleware/          # 中间件（认证/日志/访客追踪）
│   ├── model/              # 数据模型
│   ├── service/            # 业务逻辑
│   └── util/               # 工具函数
├── templates/              # HTML 模板
│   ├── admin/              # 后台模板
│   └── partials/           # 公共组件
├── static/                 # 静态资源（CSS/JS）
├── uploads/                # 上传文件
│   ├── avatar/
│   ├── cover/
│   └── article/
├── data/                   # SQLite 数据库文件
└── exports/                # Markdown 导出文件
```

## API 路由

### 前台

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 首页 |
| GET | `/article/:slug` | 文章详情 |
| GET | `/search?q=` | 搜索 |
| GET | `/tags` | 标签列表 |
| GET | `/tag/:id` | 标签详情 |
| GET | `/archive` | 归档 |
| POST | `/article/:id/like` | 点赞 |
| POST | `/article/:id/comment` | 评论 |
| GET | `/api/ai-review/:id` | AI 点评 |

### 用户

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/login` | 登录页 |
| POST | `/login` | 登录 |
| GET | `/register` | 注册页 |
| POST | `/register` | 注册 |
| POST | `/register/send-code` | 发送验证码 |
| POST | `/logout` | 登出 |
| GET | `/me` | 个人中心 |
| POST | `/me` | 更新资料 |
| POST | `/me/avatar` | 上传头像 |

### 后台

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin` | 仪表盘 |
| GET | `/admin/articles` | 文章列表 |
| POST | `/admin/articles` | 创建文章 |
| POST | `/admin/articles/:id` | 更新文章 |
| POST | `/admin/articles/:id/delete` | 删除文章 |
| POST | `/admin/articles/import` | 导入 Markdown |
| GET | `/admin/articles/:id/export` | 导出 Markdown |
| GET | `/admin/tags` | 标签管理 |
| GET | `/admin/comments` | 评论管理 |
| GET | `/admin/users` | 用户管理 |
| GET | `/admin/site` | 站点设置 |
| GET | `/admin/logs` | 操作日志 |

## 配置说明

| 配置项 | 说明 |
|--------|------|
| `app.addr` | 监听地址 |
| `database.driver` | 数据库类型：`sqlite` 或 `mysql` |
| `security.session_key` | Session 加密密钥（请修改） |
| `security.password_salt` | 密码盐值（请修改） |
| `upload.max_size_mb` | 上传文件大小限制（MB） |
| `wallpaper.enabled` | 是否启用随机壁纸 |
| `ai.*` | AI 点评配置（可选） |

## 部署

### Docker（推荐）

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o fool_blog .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/fool_blog .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 9200
CMD ["./fool_blog"]
```

```bash
docker build -t fool-blog .
docker run -p 9200:9200 -v ./data:/app/data fool-blog
```

### Systemd

```ini
[Unit]
Description=Fool Blog
After=network.target

[Service]
Type=simple
User=www
WorkingDirectory=/opt/fool-blog
ExecStart=/opt/fool-blog/fool_blog
Restart=always

[Install]
WantedBy=multi-user.target
```

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

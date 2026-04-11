# bibibibi

一个开源、自托管的笔记工具。

## 功能特性

- **笔记管理**: 创建、编辑、删除、查看笔记
- **Markdown 支持**: 实时预览 Markdown 内容
- **标签系统**: 为笔记添加标签，按标签筛选
- **评论系统**: 匿名评论，支持 Gravatar 头像
- **搜索功能**: 全文搜索笔记内容
- **用户系统**: 注册、登录、JWT 认证
- **RESTful API**: 完整的 CRUD API

## 快速开始

### Docker 部署

```bash
# 构建并启动
docker-compose up -d

# 访问 http://localhost:8080
```

### 常用命令

```bash
# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## API 文档

### 认证

- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/register` - 注册

### Bibi (笔记)

- `GET /api/v1/bibis` - 获取笔记列表
- `POST /api/v1/bibis` - 创建笔记
- `GET /api/v1/bibis/:id` - 获取笔记详情
- `PUT /api/v1/bibis/:id` - 更新笔记
- `DELETE /api/v1/bibis/:id` - 删除笔记
- `POST /api/v1/bibis/:id/pin` - 切换置顶状态
- `GET /api/v1/bibis/search` - 搜索笔记

### 标签

- `GET /api/v1/tags` - 获取标签列表
- `POST /api/v1/tags` - 创建标签
- `PUT /api/v1/tags/:id` - 更新标签
- `DELETE /api/v1/tags/:id` - 删除标签

### 评论

- `GET /api/v1/bibis/:id/comments` - 获取笔记评论
- `POST /api/v1/bibis/:id/comments` - 添加评论
- `PUT /api/v1/comments/:id` - 更新评论
- `DELETE /api/v1/comments/:id` - 删除评论

## 评论功能

评论支持以下字段：
- **名称** (必填): 评论者名称
- **邮箱** (必填): 用于获取 Gravatar 头像
- **网址** (选填): 评论者个人网站
- **内容** (必填): 评论内容

评论无需登录，使用邮箱自动获取 Gravatar 头像。

## 项目结构

```
bibibibi/
├── cmd/bibibibi/          # Go 入口
├── internal/
│   ├── api/              # API 路由和处理器
│   ├── model/            # 数据模型
│   ├── service/          # 业务逻辑
│   └── store/            # 数据库操作
├── web/                  # React 前端
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## 数据持久化

数据存储在 Docker 卷 `bibibibi-data` 中，容器重启后数据不会丢失。

### 备份数据

```bash
docker cp bibibibi-bibibibi-1:/data/bibibibi.db ./backup.db
```

### 恢复数据

```bash
docker cp ./backup.db bibibibi-bibibibi-1:/data/bibibibi.db
```

## 技术栈

- **后端**: Go + Gin + GORM + SQLite
- **前端**: React + TypeScript + Vite + Tailwind CSS
- **部署**: Docker

## 许可证

MIT License

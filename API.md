# bibibibi API 文档

## 基础信息

- **Base URL**: `/api/v1`
- **认证方式**: Bearer Token (支持 JWT 和 API Token)
- **Headers**: `Content-Type: application/json`

---

## 认证相关 `/api/v1/auth`

### POST /auth/login - 登录

**请求**
```json
{
  "username": "string",
  "password": "string"
}
```

**响应**
```json
{
  "token": "jwt_token_string",
  "user": {
    "id": 1,
    "username": "string",
    "nickname": "string",
    "email": "string",
    "website": "string",
    "is_admin": false,
    "avatar": "gravatar_url",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### POST /auth/register - 注册

**请求**
```json
{
  "username": "string",
  "password": "string",
  "nickname": "string (可选)",
  "email": "string"
}
```

**响应**: 返回用户对象

---

## 用户相关 `/api/v1/user`

### GET /user/me - 获取当前用户信息

**需要认证**: 是

**响应**
```json
{
  "id": 1,
  "username": "string",
  "nickname": "string",
  "email": "string",
  "website": "string",
  "is_admin": false,
  "avatar": "gravatar_url",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### PUT /user/me - 更新当前用户信息

**需要认证**: 是

**请求**
```json
{
  "username": "string",
  "nickname": "string",
  "email": "string",
  "website": "string",
  "password": "string (可选)"
}
```

---

## 笔记相关 `/api/v1/bibis`

### GET /bibis - 获取笔记列表

**需要认证**: 否

**Query 参数**

- `page` (int, 默认 1)
- `page_size` (int, 默认 20)
- `visibility` (string, 可选: PUBLIC/PRIVATE)
- `creator_id` (int, 可选, 按创建者筛选笔记)

**说明**:
- 当指定 `creator_id` 时，返回该用户的笔记
- 当不指定 `creator_id` 时，返回本地公开笔记 + 远程广场笔记（合并）

**响应**
```json
{
  "bibis": [
    {
      "id": "sha1_hash_string",
      "content": "markdown内容",
      "visibility": "PUBLIC",
      "pinned": false,
      "like_count": 10,
      "comment_count": 5,
      "liked": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "creator": {
        "id": 1,
        "username": "string",
        "nickname": "string",
        "avatar": "gravatar_url"
      },
      "tags": [{"id": 1, "name": "tag1"}],
      "comments": [
        {
          "id": 1,
          "parent_id": 0,
          "name": "评论者名称",
          "email": "评论者邮箱",
          "website": "评论者网站",
          "content": "评论内容",
          "avatar": "gravatar_url",
          "created_at": "2024-01-01T00:00:00Z"
        }
      ],
      "is_remote": false,
      "source_url": "https://remote.example.com"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

---

### GET /bibis/all - 获取所有公开笔记

**需要认证**: 否

**说明**: 未登录用户查看"我的笔记"时调用此接口，返回所有公开笔记

**Query 参数**

- `page` (int, 默认 1)
- `page_size` (int, 默认 20)

**响应**: 同 GET /bibis

---

### POST /bibis - 创建笔记

**需要认证**: 是

**请求**
```json
{
  "content": "string",
  "visibility": "PUBLIC | PRIVATE",
  "tag_ids": [1, 2]
}
```

---

### GET /bibis/:id - 获取笔记详情

**需要认证**: 否

---

### PUT /bibis/:id - 更新笔记

**需要认证**: 是 (仅作者)

**请求**
```json
{
  "content": "string",
  "visibility": "PUBLIC | PRIVATE",
  "tag_ids": [1, 2]
}
```

---

### DELETE /bibis/:id - 删除笔记

**需要认证**: 是 (仅作者)

---

### POST /bibis/:id/pin - 切换置顶状态

**需要认证**: 是 (仅作者)

---

### POST /bibis/:id/like - 点赞

**需要认证**: 否

**响应**
```json
{
  "liked": true
}
```

---

### GET /bibis/:id/comments - 获取评论列表

**需要认证**: 否

---

### POST /bibis/:id/comments - 创建评论

**需要认证**: 否

**请求**
```json
{
  "name": "string",
  "email": "string",
  "website": "string (可选)",
  "content": "string",
  "parent_id": 0
}
```

---

### GET /bibis/search - 搜索笔记

**需要认证**: 否

**Query 参数**

- `keyword` (string, 必填)
- `page` (int, 默认 1)
- `page_size` (int, 默认 20)

---

## 标签相关 `/api/v1/tags`

### GET /tags - 获取标签列表

**需要认证**: 否

**Query 参数**

- `creator_id` (int, 可选)

---

### POST /tags - 创建标签

**需要认证**: 是

**请求**
```json
{
  "name": "string"
}
```

---

### PUT /tags/:id - 更新标签

**需要认证**: 是 (仅创建者)

**请求**
```json
{
  "name": "string"
}
```

---

### DELETE /tags/:id - 删除标签

**需要认证**: 是 (仅创建者)

---

## 评论相关 `/api/v1/comments`

### PUT /comments/:id - 更新评论

**请求**
```json
{
  "name": "string",
  "email": "string",
  "website": "string",
  "content": "string"
}
```

---

### DELETE /comments/:id - 删除评论

**需要认证**: 否 (笔记作者可删除)

---

## 系统设置 `/api/v1/settings`

### GET /settings - 获取系统设置

**需要认证**: 是 (仅管理员)

**响应**
```json
{
  "registration_enabled": true,
  "gravatar_source": "https://weavatar.com/avatar/"
}
```

---

### PUT /settings - 更新系统设置

**需要认证**: 是 (仅管理员)

**请求**
```json
{
  "registration_enabled": true/false,
  "gravatar_source": "https://weavatar.com/avatar/"
}
```

---

## Token 管理 `/api/v1/tokens`

### GET /tokens - 获取 Token 列表

**需要认证**: 是

**响应**
```json
{
  "tokens": [
    {
      "id": 1,
      "token": "random_token_string",
      "description": "Token 描述",
      "expires_at": "2024-01-01T00:00:00Z 或 null（不过期）",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST /tokens - 创建 Token

**需要认证**: 是

**请求**
```json
{
  "description": "Token 描述",
  "expires_in_hours": 24 // 可选，不填或0表示不过期
}
```

**响应**
```json
{
  "token": {
    "id": 1,
    "token": "random_token_string",
    "description": "Token 描述",
    "expires_at": null,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### DELETE /tokens/:id - 删除 Token

**需要认证**: 是

**响应**
```json
{
  "message": "删除成功"
}
```

---

## 广场数据源 `/api/v1/feeds`

### GET /feeds - 获取广场数据源列表

**需要认证**: 是

**响应**
```json
{
  "feeds": [
    {
      "id": 1,
      "name": "数据源名称",
      "url": "https://bibi.example.com",
      "enabled": true,
      "last_fetch_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST /feeds - 创建广场数据源

**需要认证**: 是 (仅管理员)

**请求**
```json
{
  "name": "数据源名称",
  "url": "https://bibi.example.com"
}
```

---

### DELETE /feeds/:id - 删除广场数据源

**需要认证**: 是 (仅管理员)

---

### POST /feeds/sync - 同步广场数据源

**需要认证**: 是 (仅管理员)

**说明**: 手动触发从所有已启用的数据源获取最新笔记

---

## 公开系统设置 `/api/v1/public`

### GET /public/settings - 获取公开设置

**需要认证**: 否

**响应**
```json
{
  "registration_enabled": true/false
}
```

---

### GET /public/latest-bibi-card - 获取广场最新笔记图片

**需要认证**: 否

**说明**: 返回广场最新一条公开笔记渲染成的 PNG 图片（400x300 像素）

**响应**: PNG 图片数据

**使用场景**: 可用于在外部网站或平台展示 bibibibi 广场的最新笔记

---

## 备注

### Gravatar 头像

用户和评论的 `avatar` 字段通过 Gravatar 自动生成，基于邮箱的 MD5 哈希值:

```
https://www.gravatar.com/avatar/{md5(email)}?s=80&d=identicon
```

### 认证方式

- 支持两种 Token 认证方式：JWT Token 和 API Token
- API Token 可在系统设置中管理，支持描述和过期时间
- API Token 不会过期（除非手动删除或设置过期时间）

### 管理员说明

- 第一个注册的用户自动成为管理员
- 管理员可以在系统设置中开/关注册功能
- 管理员可以修改 Gravatar 头像源地址

### 点赞规则

- 匿名用户可以点赞
- 点赞后 10 秒内不可重复点赞
- 点赞不可取消

### 笔记 ID 说明

- 笔记 ID 使用 SHA1 全局唯一算法生成
- 算法: `username + timestamp + random -> SHA1`
- 确保不同用户的笔记 ID 不会冲突

### 远程笔记

- 笔记广场可聚合来自其他 bibibibi 实例的公开笔记
- 远程笔记通过 `is_remote` 字段标识
- 远程笔记的 `source_url` 字段标识来源实例地址
- 远程笔记的点赞/评论操作直接调用远程 API

### 数据库索引

| 表 | 索引类型 | 字段 |
|---|---|---|
| User | 唯一索引 | email |
| Tag | 复合唯一索引 | (creator_id, name) |
| Like | 复合唯一索引 | (bibi_id, user_id) |
| Token | 唯一索引 | token |

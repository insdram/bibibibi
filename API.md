# bibibibi API 文档

## 基础信息

- **Base URL**: `/api/v1`
- **认证方式**: JWT Bearer Token
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

**响应**
```json
{
  "bibis": [
    {
      "id": 1,
      "content": "markdown内容",
      "visibility": "PUBLIC",
      "pinned": false,
      "like_count": 10,
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
      "comments": []
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

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

## 备注

### Gravatar 头像

用户和评论的 `avatar` 字段通过 Gravatar 自动生成，基于邮箱的 MD5 哈希值:

```
https://www.gravatar.com/avatar/{md5(email)}?s=80&d=identicon
```

### 管理员说明

- 第一个注册的用户自动成为管理员
- 管理员可以在系统设置中开/关注册功能

### 点赞规则

- 匿名用户可以点赞
- 点赞后 10 秒内不可重复点赞
- 点赞不可取消

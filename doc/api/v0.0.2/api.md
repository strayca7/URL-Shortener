# URL Shortener API 文档

## 基本信息
- **标题**: URL Shortener API
- **描述**: URL缩短服务的API文档
- **版本**: 1.0.0
- **基础URL**: `http://localhost:8080`

## 认证方式
系统使用两种认证方式：
1. **Bearer Token认证**
   - 在请求头中添加: `Authorization: Bearer <access_token>`
2. **Refresh Token认证**
   - 在请求头中添加: `refresh_token: <refresh_token>`

## API端点

### 1. 系统相关接口

#### GET /healthz
健康检查端点

**响应**
- `200`: 成功响应
```json
{
    "message": "ok"
}
```

### 2. 公共接口 (/public)

#### POST /public/register
用户注册

**请求体**
```json
{
    "email": "user@example.com",     // 必填，有效的邮箱地址
    "password": "password123"        // 必填，最大32字符
}
```

**响应**
- `201`: 注册成功
```json
{
    "user_id": "xxx",
    "email": "user@example.com"
}
```

#### POST /public/login
用户登录（限制：每秒5次请求）

**请求体**
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**响应**
- `200`: 登录成功
```json
{
    "access_token": "xxx",
    "refresh_token": "xxx",
    "user": {
        "user_id": "xxx",
        "email": "user@example.com"
    }
}
```

#### GET /public/{code}
公共短链接重定向

**路径参数**
- `code`: 短链接代码

**响应**
- `302`: 重定向到原始URL
- `404`: 短链接不存在或已过期

#### GET /public/shortcodes
获取所有公共短链接列表

**响应**
- `200`: 成功获取短链接列表
```json
[
    {
        "id": "xxx",
        "original_url": "https://example.com",
        "short_code": "abc123",
        "created_at": "2024-05-06T12:00:00Z",
        "expires_at": "2024-06-06T12:00:00Z",
        "access_count": 0
    }
]
```

#### DELETE /public/short/{code}
删除公共短链接

**路径参数**
- `code`: 短链接代码

**响应**
- `200`: 成功删除
- `404`: 短链接不存在

### 3. 认证接口 (/auth)
> 以下所有接口都需要在请求头中包含：
> - `Authorization: Bearer <access_token>`
> - `refresh_token: <refresh_token>`

#### POST /auth/short/new
创建短链接

**请求体**
```json
{
    "url": "https://example.com/very/long/url"  // 必填，原始URL
}
```

**响应**
- `200`: 成功创建短链接
- `401`: 未授权

#### POST /auth/{code}
用户短链接重定向

**路径参数**
- `code`: 短链接代码

**响应**
- `302`: 重定向到原始URL
- `404`: 链接不存在

#### GET /auth/shortcodes
获取用户的所有短链接

**响应**
- `200`: 成功获取短链接列表
```json
[
    {
        "id": "xxx",
        "original_url": "https://example.com",
        "short_code": "abc123",
        "created_at": "2024-05-06T12:00:00Z",
        "expires_at": "2024-06-06T12:00:00Z",
        "access_count": 0
    }
]
```

#### POST /auth/refresh
刷新访问令牌

**响应**
- `200`: 成功刷新令牌
- `401`: 令牌无效或已过期

## 数据模型

### ShortURL
```json
{
    "id": "string",           // 短链接唯一标识符
    "original_url": "string", // 原始URL
    "short_code": "string",   // 短链接代码
    "created_at": "string",   // 创建时间(ISO 8601格式)
    "expires_at": "string",   // 过期时间(ISO 8601格式)
    "access_count": "integer" // 访问计数
}
```

### 通用错误响应
所有接口在发生错误时将返回以下格式：
```json
{
    "error": "错误描述信息"
}
```
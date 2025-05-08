# URL Shortener API 文档

## 基本信息
- 版本：0.0.1
- 基础URL：`http://localhost:8080`

## 认证方式
系统使用两种认证方式：
1. Bearer Token认证
   - 在请求头中添加：`Authorization: Bearer <access_token>`
2. Refresh Token认证
   - 在请求头中添加：`refresh_token: <refresh_token>`

## API端点

### 健康检查
#### GET /ping
检查服务是否正常运行。

**响应**
- 200: 成功
```json
{
    "message": "pong"
}
```

### 用户认证

#### POST /public/register
用户注册接口。

**请求体**
```json
{
    "email": "user@example.com",     // 必填，有效的邮箱地址
    "password": "password123"        // 必填，8-32字符
}
```

**响应**
- 201: 注册成功
```json
{
    "user_id": "xxx",
    "email": "user@example.com"
}
```
- 400: 参数格式错误
- 409: 邮箱已被注册
- 500: 服务器内部错误

#### POST /public/login
用户登录接口（限制：每秒5次请求）

**请求体**
```json
{
    "email": "user@example.com",     // 必填，邮箱地址
    "password": "password123"        // 必填，密码
}
```

**响应**
- 200: 登录成功
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
- 400: 请求格式无效
- 401: 认证失败
- 500: 服务器内部错误

### 短链接操作

#### POST /auth/shorten
创建短链接

**请求头**
- `Authorization: Bearer <access_token>` (必填)
- `refresh_token: <refresh_token>` (必填)

**请求体**
```json
{
    "url": "https://example.com/very/long/url"  // 必填，原始URL
}
```

**响应**
- 200: 成功创建短链接
- 401: 未授权
- 500: 服务器内部错误

#### POST /auth/short/{code}
短链接重定向

**请求头**
- `Authorization: Bearer <access_token>` (必填)
- `refresh_token: <refresh_token>` (必填)

**路径参数**
- `code`: 短链接代码

**响应**
- 302: 重定向到原始URL
- 404: 链接不存在或已过期
- 500: 服务器内部错误

#### GET /auth/shortcodes
获取用户的所有短链接

**请求头**
- `Authorization: Bearer <access_token>` (必填)
- `refresh_token: <refresh_token>` (必填)

**响应**
- 200: 成功获取短链接列表
```json
[
    {
        "id": "xxx",
        "original_url": "https://example.com",
        "short_code": "abc123",
        "created_at": "2024-05-06T12:00:00Z",
        "expires_at": "2024-06-06T12:00:00Z"
    }
]
```
- 401: 未授权
- 500: 服务器内部错误

#### POST /auth/refresh
刷新访问令牌

**请求头**
- `Authorization: Bearer <access_token>` (必填)
- `refresh_token: <refresh_token>` (必填)

**响应**
- 200: 成功刷新令牌
- 401: 令牌无效或已过期
- 500: 服务器内部错误
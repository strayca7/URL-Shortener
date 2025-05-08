# URL Shortener API 文档

## 基本信息
- **标题**: URL Shortener API
- **描述**: URL缩短服务的API文档
- **版本**: 0.0.2
- **基础URL**: `http://localhost:8080`

## 认证方式
系统使用两种认证方式：
1. **Bearer Token认证**
   - 在请求头中添加: `Authorization: Bearer <access_token>`
2. **Refresh Token认证**
   - 在请求头中添加: `refresh_token: <refresh_token>`

## 数据模型

### LoginRequest
```json
{
    "email": "string",     // 邮箱地址，必填
    "password": "string"   // 密码，8-32字符，必填
}
```

### LoginResponse
```json
{
    "access_token": "string",
    "refresh_token": "string",
    "user": {
        "user_id": "string",
        "email": "string"
    }
}
```

### RegisterRequest
```json
{
    "email": "string",     // 邮箱地址，必填
    "password": "string"   // 密码，8-32字符，必填
}
```

### RegisterResponse
```json
{
    "user_id": "string",
    "email": "string"
}
```

### CreateShortURLRequest
```json
{
    "url": "string"        // 原始URL，必填，必须是有效的URI格式
}
```

### ReturnShortURL
```json
{
    "original_url": "string",
    "short_code": "string"
}
```

### ShortURL
```json
{
    "id": "string",
    "original_url": "string",
    "short_code": "string",
    "created_at": "string",   // ISO 8601日期时间格式
    "expires_at": "string",   // ISO 8601日期时间格式
    "access_count": 0         // 整数类型
}
```

## API端点

### 1. 系统接口

#### GET /healthz
健康检查端点

**响应**
- `200`: 成功
```json
{
    "message": "ok"
}
```

### 2. 公共接口

#### POST /public/register
用户注册

**请求体**: `LoginRequest`

**响应**
- `201`: 注册成功 - `RegisterResponse`
- `400`: 参数格式错误
- `409`: 邮箱已被注册
- `500`: 服务器内部错误

#### POST /public/login
用户登录（限制：每秒5次请求）

**请求体**: `LoginRequest`

**响应**
- `200`: 登录成功 - `LoginResponse`
- `400`: 请求格式无效
- `401`: 认证失败
- `500`: 服务器内部错误

#### POST /public/short/new
创建公共短链接

**请求体**: `CreateShortURLRequest`

**响应**
- `201`: 创建成功 - `ReturnShortURL`
- `400`: 请求格式无效
- `500`: 服务器内部错误

#### GET /public/{code}
公共短链接重定向

**参数**
- `code`: 短链接代码 (path参数，必填)

**响应**
- `302`: 重定向到原始URL
- `404`: 短链接不存在或已过期
- `500`: 服务器内部错误

#### GET /public/shortcodes
获取所有公共短链接

**响应**
- `200`: 成功 - `ShortURL[]`
- `500`: 服务器内部错误

#### DELETE /public/short/{code}
删除公共短链接

**参数**
- `code`: 短链接代码 (path参数，必填)

**响应**
- `200`: 删除成功
- `404`: 短链接不存在
- `500`: 服务器内部错误

### 3. 认证接口
> 以下所有接口都需要在请求头中包含：
> - `Authorization: Bearer <access_token>`
> - `refresh_token: <refresh_token>`

#### POST /auth/short/new
创建用户短链接

**请求体**: `CreateShortURLRequest`

**响应**
- `200`: 创建成功 - `ReturnShortURL`
- `401`: 未授权
- `500`: 服务器内部错误

#### POST /auth/{code}
用户短链接重定向

**参数**
- `code`: 短链接代码 (path参数，必填)

**响应**
- `302`: 重定向到原始URL
- `404`: 链接不存在
- `500`: 服务器内部错误

#### GET /auth/shortcodes
获取用户的所有短链接

**响应**
- `200`: 成功 - `ShortURL[]`
- `401`: 未授权
- `500`: 服务器内部错误

#### POST /auth/refresh
刷新访问令牌

**响应**
- `200`: 刷新成功
- `401`: 令牌无效或已过期
- `500`: 服务器内部错误

### 错误响应
所有接口在发生错误时将返回以下格式：
```json
{
    "error": "错误描述信息"
}
```
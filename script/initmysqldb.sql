CREATE DATABASE IF NOT EXISTS miniurl CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE miniurl;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE COMMENT '用户唯一标识 (UUID)',
    email VARCHAR(255) NOT NULL UNIQUE COMMENT '用户邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS short_urls (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    original_url TEXT NOT NULL COMMENT '原始 URL',
    short_code VARCHAR(10) NOT NULL UNIQUE COMMENT '短链代码',
    user_id VARCHAR(36) NOT NULL COMMENT '用户 ID',
    access_count INT DEFAULT 0 COMMENT '访问次数',
    client_ips JSON COMMENT '访问者 IP 列表',
    expire_at TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO users (user_id, email, password_hash) VALUES
('test-user-id', 'test@example.com', 'hashed-password');

INSERT INTO short_urls (original_url, short_code, user_id, access_count, client_ips, expire_at) VALUES
('https://www.example.com', 'abc123', 'test-user-id', 0, '["127.0.0.1"]', NULL);
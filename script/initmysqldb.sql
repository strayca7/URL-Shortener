CREATE DATABASE IF NOT EXISTS miniurl CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE miniurl;

CREATE TABLE users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    user_id VARCHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    UNIQUE INDEX idx_user_id (user_id),
    UNIQUE INDEX idx_email (email)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE short_urls (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL,
    expire_at TIMESTAMP NULL,
    access_count INT DEFAULT 0,
    user_id VARCHAR(36) NOT NULL,
    UNIQUE INDEX idx_short_code (short_code),
    INDEX idx_expire_at (expire_at),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE client_ips (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    ip_address VARCHAR(45) NOT NULL,
    short_url_id INT UNSIGNED NOT NULL,
    INDEX idx_short_url_id (short_url_id),
    FOREIGN KEY (short_url_id) REFERENCES short_urls(id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
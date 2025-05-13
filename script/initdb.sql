CREATE DATABASE IF NOT EXISTS miniurl DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE miniurl;
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    user_id VARCHAR(36) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    INDEX idx_users_deleted_at (deleted_at)
);
CREATE TABLE IF NOT EXISTS user_short_urls (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    expire_at DATETIME(3),
    access_count INT DEFAULT 0,
    user_id VARCHAR(36) NOT NULL,
    INDEX idx_user_short_urls_deleted_at (deleted_at),
    INDEX idx_user_short_urls_expire_at (expire_at),
    INDEX idx_user_short_urls_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS client_ips (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    ip_address VARCHAR(45) NOT NULL,
    short_url_id BIGINT UNSIGNED NOT NULL,
    INDEX idx_client_ips_deleted_at (deleted_at),
    INDEX idx_client_ips_short_url_id (short_url_id),
    FOREIGN KEY (short_url_id) REFERENCES user_short_urls(id)
);
CREATE TABLE IF NOT EXISTS public_short_urls (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    short_code VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    expires_at DATETIME(3),
    access_count INT UNSIGNED DEFAULT 0,
    INDEX idx_public_short_urls_deleted_at (deleted_at)
);
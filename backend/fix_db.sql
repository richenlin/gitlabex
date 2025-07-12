-- 修复数据库迁移问题
-- 删除并重新创建 users 表

-- 删除现有的 users 表
DROP TABLE IF EXISTS users CASCADE;

-- 重新创建 users 表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar TEXT,
    role INTEGER NOT NULL DEFAULT 2,
    gitlab_id BIGINT UNIQUE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_users_gitlab_id ON users(gitlab_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- 插入测试用户
INSERT INTO users (username, email, name, gitlab_id, role) VALUES 
('admin', 'admin@example.com', '管理员', 1, 1),
('teacher', 'teacher@example.com', '教师', 2, 2),
('student', 'student@example.com', '学生', 3, 3);

-- 检查数据
SELECT * FROM users; 
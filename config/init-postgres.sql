-- PostgreSQL 初始化脚本
-- 为GitLabEx应用创建额外的数据库和用户

-- 创建gitlabex数据库
CREATE DATABASE gitlabex;

-- 创建gitlabex用户
CREATE USER gitlabex WITH ENCRYPTED PASSWORD 'password123';

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE gitlabex TO gitlabex;

-- 切换到gitlabex数据库
\c gitlabex;

-- 授予schema权限
GRANT ALL ON SCHEMA public TO gitlabex;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gitlabex;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gitlabex;

-- 等待应用启动并创建表结构后，插入测试数据
-- 这个脚本将在应用第一次启动后通过单独的初始化脚本执行

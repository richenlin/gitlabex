-- 创建GitLab用户（如果不存在）
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'gitlab') THEN
      CREATE ROLE gitlab LOGIN PASSWORD 'password123';
   END IF;
END
$do$;

-- 创建GitLabEx用户（如果不存在）
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'gitlabex') THEN
      CREATE ROLE gitlabex LOGIN PASSWORD 'password123';
   END IF;
END
$do$;

-- 创建GitLab数据库（如果不存在）
SELECT 'CREATE DATABASE gitlab'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'gitlab')\gexec

-- 创建GitLabEx数据库（如果不存在）
SELECT 'CREATE DATABASE gitlabex'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'gitlabex')\gexec

-- 授权用户访问数据库
GRANT ALL PRIVILEGES ON DATABASE gitlab TO gitlab;
GRANT ALL PRIVILEGES ON DATABASE gitlabex TO gitlabex;

-- 切换到gitlab数据库并设置所有者
\c gitlab
ALTER DATABASE gitlab OWNER TO gitlab;
GRANT ALL ON SCHEMA public TO gitlab;

-- 切换到gitlabex数据库并设置所有者
\c gitlabex
ALTER DATABASE gitlabex OWNER TO gitlabex;
GRANT ALL ON SCHEMA public TO gitlabex;

-- 显示创建的数据库
\l 
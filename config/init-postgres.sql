-- 创建GitLab数据库
CREATE DATABASE gitlab;

-- 创建GitLabEx数据库（如果不存在）
CREATE DATABASE gitlabex;

-- 授权用户访问两个数据库
GRANT ALL PRIVILEGES ON DATABASE gitlab TO gitlabex;
GRANT ALL PRIVILEGES ON DATABASE gitlabex TO gitlabex;

-- 显示创建的数据库
\l 
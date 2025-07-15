-- 为项目表添加互动开发相关字段
ALTER TABLE projects ADD COLUMN IF NOT EXISTS code_editor_enabled BOOLEAN DEFAULT true;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS main_branch_protected BOOLEAN DEFAULT true;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS allowed_file_types TEXT[];
ALTER TABLE projects ADD COLUMN IF NOT EXISTS max_file_size BIGINT DEFAULT 10485760;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS auto_save_interval INTEGER DEFAULT 30;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS student_branch_prefix VARCHAR(50) DEFAULT 'student';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS enable_real_time_collab BOOLEAN DEFAULT false;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS require_commit_message BOOLEAN DEFAULT true;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS editor_theme VARCHAR(50) DEFAULT 'vs-dark';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS editor_font_size INTEGER DEFAULT 14;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS editor_tab_size INTEGER DEFAULT 4;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS enable_linting BOOLEAN DEFAULT true;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS enable_formatting BOOLEAN DEFAULT true;

-- 为项目成员表添加互动开发相关字段
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS branch_created_at TIMESTAMP;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS last_edit_time TIMESTAMP;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS online_edit_enabled BOOLEAN DEFAULT true;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS local_clone_enabled BOOLEAN DEFAULT true;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS files_modified INTEGER DEFAULT 0;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS lines_added INTEGER DEFAULT 0;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS lines_deleted INTEGER DEFAULT 0;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS editor_preferences TEXT;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS last_active_time TIMESTAMP;

-- 创建项目文件表
CREATE TABLE IF NOT EXISTS project_files (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(50),
    file_size BIGINT,
    branch VARCHAR(100) DEFAULT 'main',
    content TEXT,
    content_hash VARCHAR(64),
    is_directory BOOLEAN DEFAULT false,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- GitLab 相关字段
    gitlab_file_id VARCHAR(100),
    gitlab_blob_id VARCHAR(100),
    gitlab_commit_id VARCHAR(100),
    gitlab_file_url VARCHAR(500),
    
    -- 编辑相关字段
    is_editable BOOLEAN DEFAULT true,
    last_edited_by INTEGER REFERENCES users(id),
    last_edited_at TIMESTAMP,
    edit_lock_by INTEGER REFERENCES users(id),
    edit_lock_at TIMESTAMP,
    edit_lock_expires TIMESTAMP,
    language VARCHAR(50),
    encoding VARCHAR(20) DEFAULT 'utf-8',
    
    UNIQUE(project_id, file_path, branch)
);

-- 创建代码编辑会话表
CREATE TABLE IF NOT EXISTS code_edit_sessions (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    branch VARCHAR(100) NOT NULL,
    session_id VARCHAR(100) UNIQUE NOT NULL,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    last_ping TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    changes_count INTEGER DEFAULT 0,
    saved_count INTEGER DEFAULT 0
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_project_files_project_id ON project_files(project_id);
CREATE INDEX IF NOT EXISTS idx_project_files_branch ON project_files(branch);
CREATE INDEX IF NOT EXISTS idx_project_files_file_path ON project_files(file_path);
CREATE INDEX IF NOT EXISTS idx_project_files_is_directory ON project_files(is_directory);
CREATE INDEX IF NOT EXISTS idx_project_files_last_edited_at ON project_files(last_edited_at);

CREATE INDEX IF NOT EXISTS idx_code_edit_sessions_project_id ON code_edit_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_code_edit_sessions_user_id ON code_edit_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_code_edit_sessions_status ON code_edit_sessions(status);
CREATE INDEX IF NOT EXISTS idx_code_edit_sessions_session_id ON code_edit_sessions(session_id);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_project_files_updated_at 
    BEFORE UPDATE ON project_files 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE project_files IS '项目文件记录表';
COMMENT ON COLUMN project_files.file_path IS '文件相对路径';
COMMENT ON COLUMN project_files.content IS '文件内容（小文件）';
COMMENT ON COLUMN project_files.content_hash IS '内容SHA256哈希';
COMMENT ON COLUMN project_files.edit_lock_by IS '编辑锁定者ID';
COMMENT ON COLUMN project_files.edit_lock_expires IS '编辑锁定过期时间';

COMMENT ON TABLE code_edit_sessions IS '代码编辑会话表';
COMMENT ON COLUMN code_edit_sessions.session_id IS '会话唯一标识';
COMMENT ON COLUMN code_edit_sessions.last_ping IS '最后心跳时间';
COMMENT ON COLUMN code_edit_sessions.changes_count IS '修改次数';
COMMENT ON COLUMN code_edit_sessions.saved_count IS '保存次数'; 
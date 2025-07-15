-- 修复表结构，添加缺失的字段

-- 为projects表添加缺失的字段
ALTER TABLE projects ADD COLUMN IF NOT EXISTS type VARCHAR(50) DEFAULT 'practice';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS max_members INTEGER DEFAULT 10;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS total_assignments INTEGER DEFAULT 0;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS completed_assignments INTEGER DEFAULT 0;

-- 为project_members表添加缺失的字段
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS user_id INTEGER;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'student';
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- 更新现有的project_members数据
UPDATE project_members SET user_id = student_id WHERE user_id IS NULL;
UPDATE project_members SET role = 'student' WHERE role IS NULL;
UPDATE project_members SET is_active = true WHERE is_active IS NULL;

-- 为assignments表添加缺失的字段
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS max_score INTEGER DEFAULT 100;

-- 为assignment_submissions表添加缺失的字段
ALTER TABLE assignment_submissions ADD COLUMN IF NOT EXISTS graded_by INTEGER REFERENCES users(id); 
-- Migration: Remove Class Management System
-- Purpose: Remove class-related tables and data, update Project table structure
-- Version: 005
-- Created: 2024-03-15

-- 开始事务
BEGIN;

-- 1. 备份数据（如果需要恢复）
-- 注释掉的备份语句，生产环境使用前请先备份数据

-- 2. 删除班级相关外键约束（如果存在）
-- 删除项目表中的班级外键约束
ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_projects_class;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_class_id_fkey;

-- 3. 移除项目表中的班级关联字段
ALTER TABLE projects DROP COLUMN IF EXISTS class_id;

-- 4. 删除班级相关表（按依赖关系顺序删除）
-- 删除班级成员表
DROP TABLE IF EXISTS class_members CASCADE;

-- 删除班级表
DROP TABLE IF EXISTS classes CASCADE;

-- 5. 确保项目表有必要的字段
-- 添加项目代码字段（如果不存在）
ALTER TABLE projects ADD COLUMN IF NOT EXISTS project_code VARCHAR(20) UNIQUE;

-- 添加最大成员数字段（如果不存在）
ALTER TABLE projects ADD COLUMN IF NOT EXISTS max_members INTEGER DEFAULT 50;

-- 添加当前成员数字段（如果不存在）
ALTER TABLE projects ADD COLUMN IF NOT EXISTS current_members INTEGER DEFAULT 0;

-- 6. 更新现有项目的项目代码（如果为空）
UPDATE projects 
SET project_code = 'PROJ' || TO_CHAR(EXTRACT(YEAR FROM created_at), 'YYYY') || LPAD(id::text, 4, '0')
WHERE project_code IS NULL OR project_code = '';

-- 7. 确保用户表有必要的权限相关字段
-- 添加GitLab用户ID字段（如果不存在）
ALTER TABLE users ADD COLUMN IF NOT EXISTS gitlab_id INTEGER;

-- 添加动态角色字段（如果不存在）
ALTER TABLE users ADD COLUMN IF NOT EXISTS dynamic_role VARCHAR(20) DEFAULT '';

-- 8. 确保作业表结构正确
-- 添加评审报告字段（如果不存在）
ALTER TABLE assignment_submissions ADD COLUMN IF NOT EXISTS review_report JSONB;

-- 添加详细评分字段（如果不存在）
ALTER TABLE assignment_submissions ADD COLUMN IF NOT EXISTS detailed_scores JSONB;

-- 9. 创建项目成员统计触发器函数
CREATE OR REPLACE FUNCTION update_project_member_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE projects 
        SET current_members = (
            SELECT COUNT(*) 
            FROM project_members 
            WHERE project_id = NEW.project_id
        )
        WHERE id = NEW.project_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE projects 
        SET current_members = (
            SELECT COUNT(*) 
            FROM project_members 
            WHERE project_id = OLD.project_id
        )
        WHERE id = OLD.project_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 10. 创建触发器来自动更新项目成员数
DROP TRIGGER IF EXISTS trigger_update_project_member_count ON project_members;
CREATE TRIGGER trigger_update_project_member_count
    AFTER INSERT OR DELETE ON project_members
    FOR EACH ROW
    EXECUTE FUNCTION update_project_member_count();

-- 11. 初始化现有项目的成员数
UPDATE projects 
SET current_members = (
    SELECT COUNT(*) 
    FROM project_members 
    WHERE project_members.project_id = projects.id
);

-- 12. 添加索引以提高性能
-- 项目代码索引
CREATE INDEX IF NOT EXISTS idx_projects_project_code ON projects(project_code);

-- 用户GitLab ID索引
CREATE INDEX IF NOT EXISTS idx_users_gitlab_id ON users(gitlab_id);

-- 作业项目ID索引
CREATE INDEX IF NOT EXISTS idx_assignments_project_id ON assignments(project_id);

-- 项目创建者索引
CREATE INDEX IF NOT EXISTS idx_projects_teacher_id ON projects(teacher_id);

-- 13. 更新表注释
COMMENT ON TABLE projects IS '课题表 - 教师直接管理的研究课题';
COMMENT ON COLUMN projects.project_code IS '课题代码 - 学生加入课题时使用';
COMMENT ON COLUMN projects.teacher_id IS '课题创建者（教师）ID';
COMMENT ON COLUMN projects.max_members IS '最大成员数限制';
COMMENT ON COLUMN projects.current_members IS '当前成员数（自动计算）';

COMMENT ON TABLE project_members IS '课题成员表 - 学生参与的课题';
COMMENT ON TABLE assignments IS '作业表 - 基于课题的作业管理';
COMMENT ON TABLE assignment_submissions IS '作业提交表 - 包含详细评审信息';

-- 14. 验证数据完整性
-- 检查是否有孤立的项目成员记录
DO $$
DECLARE
    orphaned_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphaned_count
    FROM project_members pm
    LEFT JOIN projects p ON pm.project_id = p.id
    WHERE p.id IS NULL;
    
    IF orphaned_count > 0 THEN
        RAISE WARNING '发现 % 条孤立的项目成员记录，需要手动清理', orphaned_count;
    END IF;
END $$;

-- 提交事务
COMMIT;

-- 输出迁移完成信息
DO $$
BEGIN
    RAISE NOTICE '=== 班级管理系统移除完成 ===';
    RAISE NOTICE '1. 已删除 classes 和 class_members 表';
    RAISE NOTICE '2. 已移除 projects 表中的 class_id 字段';
    RAISE NOTICE '3. 已添加项目代码和成员数管理功能';
    RAISE NOTICE '4. 已优化数据库索引和约束';
    RAISE NOTICE '5. 系统现在支持教师直接管理课题';
    RAISE NOTICE '================================';
END $$; 
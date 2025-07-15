-- 修复约束和字段问题

-- 修复project_members表
ALTER TABLE project_members ALTER COLUMN student_id DROP NOT NULL;
ALTER TABLE project_members ALTER COLUMN user_id SET NOT NULL;

-- 修复assignments表
ALTER TABLE assignments ALTER COLUMN teacher_id DROP NOT NULL;

-- 删除重复的示例数据
DELETE FROM project_members WHERE user_id IS NULL;
DELETE FROM assignments WHERE teacher_id IS NULL;
DELETE FROM assignment_submissions WHERE assignment_id NOT IN (SELECT id FROM assignments);

-- 清理现有数据
TRUNCATE TABLE assignment_submissions, assignments, project_members, projects, classes, users RESTART IDENTITY CASCADE; 
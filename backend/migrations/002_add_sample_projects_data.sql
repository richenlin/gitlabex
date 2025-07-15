-- 添加课题管理示例数据
-- 这个迁移文件包含了课题、成员、作业等示例数据

-- 首先确保基础表存在
CREATE TABLE IF NOT EXISTS classes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    gitlab_group_id INTEGER,
    gitlab_group_path VARCHAR(255),
    join_code VARCHAR(20) UNIQUE,
    max_students INTEGER DEFAULT 50,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    code VARCHAR(50) UNIQUE,
    class_id INTEGER REFERENCES classes(id),
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    gitlab_project_id INTEGER,
    gitlab_project_path VARCHAR(255),
    gitlab_web_url VARCHAR(500),
    status VARCHAR(50) DEFAULT 'active',
    type VARCHAR(50) DEFAULT 'practice',
    start_date DATE,
    end_date DATE,
    max_members INTEGER DEFAULT 10,
    total_assignments INTEGER DEFAULT 0,
    completed_assignments INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS project_members (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'student', -- 'teacher' or 'student'
    gitlab_access_level INTEGER DEFAULT 20, -- GitLab权限级别
    branch_name VARCHAR(100),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(project_id, user_id)
);

CREATE TABLE IF NOT EXISTS assignments (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMP,
    max_score INTEGER DEFAULT 100,
    status VARCHAR(50) DEFAULT 'active',
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS assignment_submissions (
    id SERIAL PRIMARY KEY,
    assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    files_submitted TEXT[],
    score INTEGER,
    feedback TEXT,
    status VARCHAR(50) DEFAULT 'submitted', -- 'submitted', 'graded', 'returned'
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    graded_at TIMESTAMP,
    graded_by INTEGER REFERENCES users(id),
    UNIQUE(assignment_id, student_id)
);

-- 插入示例用户（如果不存在）
INSERT INTO users (id, gitlab_id, username, email, name, role, active, created_at) 
VALUES 
    (1, 1, 'admin', 'admin@example.com', '系统管理员', 1, true, CURRENT_TIMESTAMP),
    (2, 2, 'zhang_teacher', 'zhang@example.com', '张教授', 2, true, CURRENT_TIMESTAMP),
    (3, 3, 'li_teacher', 'li@example.com', '李研究员', 2, true, CURRENT_TIMESTAMP),
    (4, 4, 'wang_teacher', 'wang@example.com', '王老师', 2, true, CURRENT_TIMESTAMP),
    (5, 5, 'wang_student', 'wang.student@example.com', '王同学', 3, true, CURRENT_TIMESTAMP),
    (6, 6, 'zhang_student', 'zhang.student@example.com', '张同学', 3, true, CURRENT_TIMESTAMP),
    (7, 7, 'li_student', 'li.student@example.com', '李同学', 3, true, CURRENT_TIMESTAMP),
    (8, 8, 'chen_student', 'chen.student@example.com', '陈同学', 3, true, CURRENT_TIMESTAMP),
    (9, 9, 'zhao_student', 'zhao.student@example.com', '赵同学', 3, true, CURRENT_TIMESTAMP),
    (10, 10, 'liu_student', 'liu.student@example.com', '刘同学', 3, true, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- 插入示例班级
INSERT INTO classes (id, name, description, teacher_id, code, active, created_at) 
VALUES 
    (1, '计算机科学基础班', '计算机科学与技术专业基础课程班级', 2, 'CS2024001', true, CURRENT_TIMESTAMP),
    (2, '软件工程实践班', '软件工程专业实践课程班级', 3, 'SE2024001', true, CURRENT_TIMESTAMP),
    (3, '人工智能研究班', '人工智能方向研究生班级', 4, 'AI2024001', true, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- 插入示例课题
INSERT INTO projects (id, name, description, code, class_id, teacher_id, status, type, start_date, end_date, max_members, total_assignments, completed_assignments, created_at) 
VALUES 
    (1, '基于Vue3的在线教育平台设计与实现', '设计并实现一个现代化的在线教育平台，支持视频播放、在线测试、作业管理等功能', 'WEB2024001', 1, 2, 'ongoing', 'graduation', '2024-02-01', '2024-06-01', 4, 6, 4, '2024-01-15'),
    (2, '机器学习在图像识别中的应用研究', '研究深度学习算法在图像分类和目标检测中的应用，提出改进方案', 'ML2024001', 2, 3, 'ongoing', 'research', '2024-01-01', '2024-12-01', 3, 4, 2, '2023-12-15'),
    (3, '智能物流管理系统', '为参加全国大学生软件设计大赛开发的智能物流管理系统，包含路径优化、库存管理等功能', 'LOG2024001', 3, 4, 'planning', 'competition', '2024-03-01', '2024-08-01', 5, 3, 1, '2024-02-20')
ON CONFLICT (id) DO NOTHING;

-- 插入课题成员
INSERT INTO project_members (project_id, user_id, role, is_active) 
VALUES 
    -- 课题1成员
    (1, 2, 'teacher', true),
    (1, 5, 'student', true),
    (1, 6, 'student', true),
    (1, 7, 'student', true),
    
    -- 课题2成员
    (2, 3, 'teacher', true),
    (2, 8, 'student', true),
    (2, 9, 'student', true),
    
    -- 课题3成员
    (3, 4, 'teacher', true),
    (3, 10, 'student', true),
    (3, 5, 'student', true),
    (3, 6, 'student', true),
    (3, 7, 'student', true);

-- 插入示例作业
INSERT INTO assignments (id, project_id, title, description, due_date, max_score, status) 
VALUES 
    -- 课题1的作业
    (1, 1, 'HTML基础练习', '完成HTML页面设计，包含基本标签和结构', '2024-02-15 23:59:59', 100, 'active'),
    (2, 1, 'CSS样式设计', '学习CSS布局和样式，实现响应式设计', '2024-02-28 23:59:59', 100, 'active'),
    (3, 1, 'JavaScript基础', '实现页面交互功能，掌握DOM操作', '2024-03-15 23:59:59', 100, 'active'),
    (4, 1, 'Vue3组件开发', '使用Vue3开发可复用组件', '2024-03-30 23:59:59', 100, 'active'),
    (5, 1, '前端路由实现', '实现单页面应用的路由功能', '2024-04-15 23:59:59', 100, 'active'),
    (6, 1, '完整项目整合', '整合所有功能，完成最终项目', '2024-05-30 23:59:59', 100, 'active'),
    
    -- 课题2的作业
    (7, 2, '文献调研报告', '完成相关领域的文献调研和分析', '2024-01-31 23:59:59', 100, 'completed'),
    (8, 2, '数据集准备', '收集和预处理图像数据集', '2024-02-28 23:59:59', 100, 'completed'),
    (9, 2, '模型设计与实现', '设计并实现深度学习模型', '2024-04-30 23:59:59', 100, 'active'),
    (10, 2, '实验结果分析', '分析实验结果并撰写报告', '2024-06-30 23:59:59', 100, 'active'),
    
    -- 课题3的作业
    (11, 3, '需求分析文档', '分析系统需求并撰写需求文档', '2024-03-15 23:59:59', 100, 'completed'),
    (12, 3, '系统架构设计', '设计系统整体架构', '2024-04-01 23:59:59', 100, 'active'),
    (13, 3, '核心功能实现', '实现系统核心功能模块', '2024-05-15 23:59:59', 100, 'active')
ON CONFLICT (id) DO NOTHING;

-- 插入示例作业提交
INSERT INTO assignment_submissions (assignment_id, student_id, content, files_submitted, score, feedback, status, submitted_at, graded_at) 
VALUES 
    -- 课题1的作业提交
    (1, 5, 'HTML基础练习完成', ARRAY['index.html', 'about.html'], 85, '结构清晰，但需要改进语义化标签的使用', 'graded', '2024-02-14 10:30:00', '2024-02-16 14:20:00'),
    (1, 6, 'HTML基础练习完成', ARRAY['index.html', 'contact.html'], 90, '很好的HTML结构，语义化标签使用恰当', 'graded', '2024-02-14 15:45:00', '2024-02-16 14:25:00'),
    (1, 7, 'HTML基础练习完成', ARRAY['index.html', 'portfolio.html'], 88, '整体不错，需要注意代码缩进', 'graded', '2024-02-15 09:15:00', '2024-02-16 14:30:00'),
    
    (2, 5, 'CSS样式设计完成', ARRAY['style.css', 'responsive.css'], 82, '布局合理，响应式设计需要改进', 'graded', '2024-02-27 16:20:00', '2024-03-01 10:15:00'),
    (2, 6, 'CSS样式设计完成', ARRAY['main.css', 'mobile.css'], 95, '优秀的响应式设计，代码规范', 'graded', '2024-02-28 11:30:00', '2024-03-01 10:20:00'),
    (2, 7, 'CSS样式设计完成', ARRAY['styles.css'], 78, '基本功能实现，但缺少移动端适配', 'graded', '2024-02-28 20:45:00', '2024-03-01 10:25:00'),
    
    (3, 5, 'JavaScript基础完成', ARRAY['script.js', 'utils.js'], 87, 'DOM操作熟练，需要改进错误处理', 'graded', '2024-03-14 14:10:00', '2024-03-16 09:30:00'),
    (3, 6, 'JavaScript基础完成', ARRAY['main.js', 'components.js'], 92, '代码结构清晰，功能实现完整', 'graded', '2024-03-14 18:25:00', '2024-03-16 09:35:00'),
    
    (4, 5, 'Vue3组件开发完成', ARRAY['components/', 'views/'], 89, '组件设计合理，可复用性好', 'graded', '2024-03-29 16:40:00', '2024-03-31 11:15:00'),
    (4, 6, 'Vue3组件开发完成', ARRAY['src/components/', 'src/views/'], 94, '优秀的组件架构，代码质量高', 'graded', '2024-03-30 10:20:00', '2024-03-31 11:20:00'),
    
    -- 课题2的作业提交
    (7, 8, '文献调研报告完成', ARRAY['literature_review.pdf'], 88, '调研全面，分析深入', 'graded', '2024-01-30 14:30:00', '2024-02-01 10:00:00'),
    (7, 9, '文献调研报告完成', ARRAY['research_report.pdf'], 85, '内容充实，需要改进引用格式', 'graded', '2024-01-31 16:45:00', '2024-02-01 10:15:00'),
    
    (8, 8, '数据集准备完成', ARRAY['dataset/', 'preprocessing.py'], 90, '数据处理规范，质量高', 'graded', '2024-02-27 11:20:00', '2024-03-01 15:30:00'),
    (8, 9, '数据集准备完成', ARRAY['data/', 'clean_data.py'], 87, '数据清洗完整，需要改进文档', 'graded', '2024-02-28 09:15:00', '2024-03-01 15:35:00'),
    
    -- 课题3的作业提交
    (11, 10, '需求分析文档完成', ARRAY['requirements.pdf'], 86, '需求分析详细，用例设计合理', 'graded', '2024-03-14 13:45:00', '2024-03-16 14:20:00');

-- 更新序列值
SELECT setval('users_id_seq', 10, true);
SELECT setval('classes_id_seq', 3, true);
SELECT setval('projects_id_seq', 3, true);
SELECT setval('project_members_id_seq', 12, true);
SELECT setval('assignments_id_seq', 13, true);
SELECT setval('assignment_submissions_id_seq', 15, true); 
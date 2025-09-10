-- GitLabEx 简化测试数据
-- 兼容当前数据库结构

SET CLIENT_ENCODING TO 'UTF8';

-- 插入测试用户数据
INSERT INTO users (id, created_at, updated_at, git_lab_id, username, email, name, role, edu_role, is_active, avatar_url) VALUES
-- 管理员
('550e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 1, 'admin', 'admin@gitlabex.com', '系统管理员', 'admin', 50, true, 'https://via.placeholder.com/64'),

-- 教师
('550e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 2, 'prof_wang', 'wang.prof@university.edu', '王教授', 'teacher', 40, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), 3, 'prof_li', 'li.prof@university.edu', '李教授', 'teacher', 40, true, 'https://via.placeholder.com/64'),

-- 助教
('550e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), 5, 'ta_chen', 'chen.ta@university.edu', '陈助教', 'assistant', 30, true, 'https://via.placeholder.com/64'),

-- 学生
('550e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), 7, 'student_001', 'student001@university.edu', '学生001', 'student', 20, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), 8, 'student_002', 'student002@university.edu', '学生002', 'student', 20, true, 'https://via.placeholder.com/64')
ON CONFLICT (id) DO NOTHING;

-- 插入研究课题数据
INSERT INTO research_projects (id, created_at, updated_at, name, description, status, creator_id, start_date, is_public) VALUES
-- 公开课题
('660e8400-e29b-41d4-a716-446655440001', '2023-11-14 10:00:00', NOW(), 'Starlink星座分析', 'SpaceX Starlink卫星星座的轨道分析和性能评估研究', 'active', '550e8400-e29b-41d4-a716-446655440002', '2023-11-14', true),
('660e8400-e29b-41d4-a716-446655440002', '2023-11-12 09:30:00', NOW(), '卫星通信链路优化', '针对LEO/GEO卫星通信链路的优化算法研究', 'active', '550e8400-e29b-41d4-a716-446655440003', '2023-11-12', true),
('660e8400-e29b-41d4-a716-446655440003', '2023-11-09 14:20:00', NOW(), '轨道机动计算器', '卫星轨道机动计算与仿真工具开发', 'active', '550e8400-e29b-41d4-a716-446655440002', '2023-11-09', true),

-- 专有课题
('660e8400-e29b-41d4-a716-446655440005', '2023-11-13 16:15:00', NOW(), 'LEO轨道碰撞预警', '低轨卫星碰撞风险评估与预警系统', 'active', '550e8400-e29b-41d4-a716-446655440003', '2023-11-13', false)
ON CONFLICT (id) DO NOTHING;

-- 插入话题数据
INSERT INTO topics (id, created_at, updated_at, title, content, project_id, author_id, status, priority, tags, view_count, like_count) VALUES
('880e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink卫星轨道参数分析方法讨论', '大家对于Starlink卫星轨道参数分析有什么好的方法推荐吗？特别是在轨道衰减预测方面。', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440007', 'active', 'normal', ARRAY['轨道分析', 'Starlink', '卫星'], 256, 12),
('880e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO卫星碰撞概率计算模型', '关于低轨卫星碰撞概率计算，我们应该采用哪种数学模型比较准确？', '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440008', 'active', 'high', ARRAY['碰撞预警', 'LEO', '概率计算'], 189, 8),
('880e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路性能优化策略', '针对不同轨道高度的卫星通信链路，有哪些有效的性能优化策略？', '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440008', 'active', 'normal', ARRAY['通信链路', '性能优化', '卫星通信'], 312, 15)
ON CONFLICT (id) DO NOTHING;

-- 插入文档数据
INSERT INTO documents (id, created_at, updated_at, title, description, file_path, file_size, file_type, mime_type, status, uploader_id, project_id, category, tags, download_count, auto_indexed) VALUES
('990e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink轨道分析报告.pdf', 'SpaceX Starlink卫星星座轨道特性分析的详细报告', '/documents/starlink-orbit-analysis.pdf', 2048576, 'document', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', '研究报告', ARRAY['Starlink', '轨道分析', '卫星星座'], 45, true),
('990e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO碰撞预警算法设计文档.docx', '低轨卫星碰撞预警系统算法设计与实现文档', '/documents/leo-collision-algorithm.docx', 1536000, 'document', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'approved', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440005', '技术文档', ARRAY['LEO', '碰撞预警', '算法设计'], 23, true),
('990e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路建模手册.pdf', '卫星通信链路数学建模和仿真方法手册', '/documents/satellite-comm-modeling.pdf', 3072000, 'document', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440002', '技术手册', ARRAY['卫星通信', '链路建模', '仿真'], 67, true)
ON CONFLICT (id) DO NOTHING;

-- 插入作业数据
INSERT INTO homeworks (id, created_at, updated_at, title, description, project_id, creator_id, status, due_date, max_grade, instructions, requirements, tags) VALUES
('aa0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink轨道参数分析作业', '分析SpaceX Starlink卫星的轨道参数，并计算其覆盖范围', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'published', '2023-12-15 23:59:59', 100, '1. 下载最新的Starlink TLE数据\n2. 计算轨道参数\n3. 分析覆盖特性\n4. 提交分析报告', ARRAY['使用Python进行计算', '报告格式为PDF', '包含可视化图表'], ARRAY['轨道分析', '作业']),
('aa0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO碰撞风险评估实验', '基于给定的卫星轨道数据，评估碰撞风险并设计预警算法', '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', 'published', '2023-12-20 23:59:59', 100, '1. 分析两颗卫星的轨道数据\n2. 计算最近接近距离\n3. 评估碰撞概率\n4. 设计预警机制', ARRAY['使用MATLAB或Python', '提交源代码和报告', '包含算法流程图'], ARRAY['碰撞预警', '风险评估', '作业'])
ON CONFLICT (id) DO NOTHING;

-- 插入公告数据
INSERT INTO announcements (id, created_at, updated_at, title, content, author_id, priority, is_active, valid_from, valid_to, target_roles) VALUES
('cc0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), '系统维护通知', '系统将于本周六晚上10:00-12:00进行维护升级，期间可能无法访问。请大家提前保存工作内容。', '550e8400-e29b-41d4-a716-446655440001', 'high', true, NOW(), '2023-12-31 23:59:59', ARRAY['admin', 'teacher', 'assistant', 'student']),
('cc0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), '新课题申请通知', '2024年春季学期新课题申请已开放，欢迎各位老师和同学申请。申请截止日期为12月30日。', '550e8400-e29b-41d4-a716-446655440001', 'normal', true, NOW(), '2023-12-30 23:59:59', ARRAY['teacher', 'assistant'])
ON CONFLICT (id) DO NOTHING;

-- 插入通知数据
INSERT INTO notifications (id, created_at, updated_at, type, title, content, recipient_id, sender_id, project_id, topic_id, homework_id, is_read, action_url) VALUES
('dd0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'project_invite', '项目邀请', '您被邀请加入"Starlink星座分析"项目', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, NULL, false, '/research-projects/660e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'homework_assigned', '新作业分配', '您有新的作业："Starlink轨道参数分析作业"', '550e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, 'aa0e8400-e29b-41d4-a716-446655440001', true, '/homework/aa0e8400-e29b-41d4-a716-446655440001')
ON CONFLICT (id) DO NOTHING;

-- 显示插入结果
DO $$
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'GitLabEx 简化测试数据初始化完成！';
    RAISE NOTICE '==============================================';
    RAISE NOTICE '已插入数据：';
    RAISE NOTICE '- 用户: % 个', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '- 研究课题: % 个', (SELECT COUNT(*) FROM research_projects);  
    RAISE NOTICE '- 话题: % 个', (SELECT COUNT(*) FROM topics);
    RAISE NOTICE '- 文档: % 个', (SELECT COUNT(*) FROM documents);
    RAISE NOTICE '- 作业: % 个', (SELECT COUNT(*) FROM homeworks);
    RAISE NOTICE '- 公告: % 个', (SELECT COUNT(*) FROM announcements);
    RAISE NOTICE '- 通知: % 个', (SELECT COUNT(*) FROM notifications);
    RAISE NOTICE '==============================================';
    RAISE NOTICE '测试账号信息：';
    RAISE NOTICE '管理员: admin / admin@gitlabex.com';
    RAISE NOTICE '教师: prof_wang / wang.prof@university.edu';
    RAISE NOTICE '学生: student_001 / student001@university.edu';
    RAISE NOTICE '==============================================';
END $$;

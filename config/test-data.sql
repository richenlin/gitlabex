-- GitLabEx 测试数据初始化脚本
-- 基于设计原型插入演示数据

-- 设置客户端编码
SET CLIENT_ENCODING TO 'UTF8';

-- 插入测试用户数据
INSERT INTO users (id, created_at, updated_at, gitlab_id, username, email, name, role, edu_role, is_active, avatar_url) VALUES
-- 管理员
('550e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 1, 'admin', 'admin@gitlabex.com', '系统管理员', 'admin', 50, true, 'https://via.placeholder.com/64'),

-- 教师
('550e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 2, 'prof_wang', 'wang.prof@university.edu', '王教授', 'teacher', 40, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), 3, 'prof_li', 'li.prof@university.edu', '李教授', 'teacher', 40, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), 4, 'prof_zhang', 'zhang.prof@university.edu', '张教授', 'teacher', 40, true, 'https://via.placeholder.com/64'),

-- 助教
('550e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), 5, 'ta_chen', 'chen.ta@university.edu', '陈助教', 'assistant', 30, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), 6, 'ta_liu', 'liu.ta@university.edu', '刘助教', 'assistant', 30, true, 'https://via.placeholder.com/64'),

-- 学生
('550e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), 7, 'student_001', 'student001@university.edu', '学生001', 'student', 20, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), 8, 'student_002', 'student002@university.edu', '学生002', 'student', 20, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), 9, 'student_003', 'student003@university.edu', '学生003', 'student', 20, true, 'https://via.placeholder.com/64'),
('550e8400-e29b-41d4-a716-446655440010', NOW(), NOW(), 10, 'student_004', 'student004@university.edu', '学生004', 'student', 20, true, 'https://via.placeholder.com/64')
ON CONFLICT (id) DO NOTHING;

-- 插入研究课题数据（基于设计原型中的航天主题）
INSERT INTO research_projects (id, created_at, updated_at, name, description, status, creator_id, gitlab_project_id, gitlab_url, start_date, is_public) VALUES
-- 公开课题
('660e8400-e29b-41d4-a716-446655440001', '2023-11-14 10:00:00', NOW(), 'Starlink星座分析', 'SpaceX Starlink卫星星座的轨道分析和性能评估研究', 'active', '550e8400-e29b-41d4-a716-446655440002', 101, 'http://localhost:8081/aerospace/starlink-analysis', '2023-11-14', true),
('660e8400-e29b-41d4-a716-446655440002', '2023-11-12 09:30:00', NOW(), '卫星通信链路优化', '针对LEO/GEO卫星通信链路的优化算法研究', 'active', '550e8400-e29b-41d4-a716-446655440003', 102, 'http://localhost:8081/aerospace/satellite-comm-optimization', '2023-11-12', true),
('660e8400-e29b-41d4-a716-446655440003', '2023-11-09 14:20:00', NOW(), '轨道机动计算器', '卫星轨道机动计算与仿真工具开发', 'active', '550e8400-e29b-41d4-a716-446655440002', 103, 'http://localhost:8081/aerospace/orbital-maneuver-calc', '2023-11-09', true),
('660e8400-e29b-41d4-a716-446655440004', '2023-11-08 11:45:00', NOW(), '太空交通管制系统', '近地轨道太空交通管制算法与系统设计', 'active', '550e8400-e29b-41d4-a716-446655440004', 104, 'http://localhost:8081/aerospace/space-traffic-control', '2023-11-08', true),

-- 专有课题
('660e8400-e29b-41d4-a716-446655440005', '2023-11-13 16:15:00', NOW(), 'LEO轨道碰撞预警', '低轨卫星碰撞风险评估与预警系统', 'active', '550e8400-e29b-41d4-a716-446655440003', 105, 'http://localhost:8081/aerospace/leo-collision-warning', '2023-11-13', false),
('660e8400-e29b-41d4-a716-446655440006', '2023-11-11 13:30:00', NOW(), 'GEO卫星通信干扰分析', '地球同步轨道卫星通信干扰建模与分析', 'active', '550e8400-e29b-41d4-a716-446655440002', 106, 'http://localhost:8081/aerospace/geo-interference-analysis', '2023-11-11', false),
('660e8400-e29b-41d4-a716-446655440007', '2023-11-10 08:00:00', NOW(), '太空碎片追踪系统', '空间碎片轨道预测与跟踪算法研究', 'active', '550e8400-e29b-41d4-a716-446655440004', 107, 'http://localhost:8081/aerospace/space-debris-tracking', '2023-11-10', false),
('660e8400-e29b-41d4-a716-446655440008', '2023-11-07 15:45:00', NOW(), '卫星星座部署模拟', '大型卫星星座部署策略仿真与优化', 'active', '550e8400-e29b-41d4-a716-446655440003', 108, 'http://localhost:8081/aerospace/constellation-deployment', '2023-11-07', false)
ON CONFLICT (id) DO NOTHING;

-- 插入项目成员关系
INSERT INTO project_members (id, created_at, updated_at, project_id, user_id, role, joined_at) VALUES
-- Starlink星座分析项目成员
('770e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'maintainer', '2023-11-14 10:00:00'), -- 王教授（创建者）
('770e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440005', 'developer', '2023-11-14 11:00:00'), -- 陈助教
('770e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440007', 'reporter', '2023-11-14 12:00:00'), -- 学生001
('770e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440008', 'reporter', '2023-11-14 13:00:00'), -- 学生002

-- LEO轨道碰撞预警项目成员
('770e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', 'maintainer', '2023-11-13 16:15:00'), -- 李教授（创建者）
('770e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440006', 'developer', '2023-11-13 17:00:00'), -- 刘助教
('770e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440009', 'reporter', '2023-11-13 18:00:00'), -- 学生003

-- 卫星通信链路优化项目成员
('770e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', 'maintainer', '2023-11-12 09:30:00'), -- 李教授（创建者）
('770e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440010', 'reporter', '2023-11-12 10:00:00') -- 学生004
ON CONFLICT (id) DO NOTHING;

-- 插入话题数据
INSERT INTO topics (id, created_at, updated_at, title, content, project_id, author_id, gitlab_issue_id, status, priority, tags, view_count, like_count) VALUES
('880e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink卫星轨道参数分析方法讨论', '大家对于Starlink卫星轨道参数分析有什么好的方法推荐吗？特别是在轨道衰减预测方面。', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440007', 1001, 'active', 'normal', ARRAY['轨道分析', 'Starlink', '卫星'], 256, 12),
('880e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO卫星碰撞概率计算模型', '关于低轨卫星碰撞概率计算，我们应该采用哪种数学模型比较准确？', '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440009', 1002, 'active', 'high', ARRAY['碰撞预警', 'LEO', '概率计算'], 189, 8),
('880e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路性能优化策略', '针对不同轨道高度的卫星通信链路，有哪些有效的性能优化策略？', '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440010', 1003, 'active', 'normal', ARRAY['通信链路', '性能优化', '卫星通信'], 312, 15),
('880e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '轨道机动算法实现难点', '在实现轨道机动计算算法时遇到的主要技术难点和解决方案分享', '660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440008', 1004, 'active', 'normal', ARRAY['轨道机动', '算法实现', '技术难点'], 156, 6)
ON CONFLICT (id) DO NOTHING;

-- 插入文档数据
INSERT INTO documents (id, created_at, updated_at, title, description, file_path, file_size, file_type, mime_type, status, uploader_id, project_id, category, tags, download_count, gitlab_file_path, gitlab_branch, auto_indexed) VALUES
('990e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink轨道分析报告.pdf', 'SpaceX Starlink卫星星座轨道特性分析的详细报告', '/documents/starlink-orbit-analysis.pdf', 2048576, 'document', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', '研究报告', ARRAY['Starlink', '轨道分析', '卫星星座'], 45, 'docs/starlink-orbit-analysis.pdf', 'main', true),
('990e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO碰撞预警算法设计文档.docx', '低轨卫星碰撞预警系统算法设计与实现文档', '/documents/leo-collision-algorithm.docx', 1536000, 'document', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'approved', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440005', '技术文档', ARRAY['LEO', '碰撞预警', '算法设计'], 23, 'docs/leo-collision-algorithm.docx', 'main', true),
('990e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路建模手册.pdf', '卫星通信链路数学建模和仿真方法手册', '/documents/satellite-comm-modeling.pdf', 3072000, 'document', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440002', '技术手册', ARRAY['卫星通信', '链路建模', '仿真'], 67, 'docs/satellite-comm-modeling.pdf', 'main', true),
('990e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '轨道机动计算代码示例.zip', '轨道机动计算的Python代码实现示例', '/documents/orbital-maneuver-code.zip', 512000, 'archive', 'application/zip', 'approved', '550e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440003', '代码示例', ARRAY['轨道机动', 'Python', '代码示例'], 34, 'src/examples/orbital-maneuver-code.zip', 'main', false),
('990e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), '太空碎片数据集.csv', '历史太空碎片轨道数据集，用于机器学习训练', '/documents/space-debris-dataset.csv', 10240000, 'data', 'text/csv', 'approved', '550e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440007', '数据集', ARRAY['太空碎片', '数据集', '机器学习'], 89, 'data/space-debris-dataset.csv', 'main', true)
ON CONFLICT (id) DO NOTHING;

-- 插入作业数据
INSERT INTO homeworks (id, created_at, updated_at, title, description, project_id, creator_id, status, due_date, max_grade, instructions, requirements, tags, gitlab_branch, gitlab_path) VALUES
('aa0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink轨道参数分析作业', '分析SpaceX Starlink卫星的轨道参数，并计算其覆盖范围', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'published', '2023-12-15 23:59:59', 100, '1. 下载最新的Starlink TLE数据\n2. 计算轨道参数\n3. 分析覆盖特性\n4. 提交分析报告', ARRAY['使用Python进行计算', '报告格式为PDF', '包含可视化图表'], ARRAY['轨道分析', '作业'], 'homework/starlink-analysis', 'assignments/starlink-orbit-analysis'),
('aa0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO碰撞风险评估实验', '基于给定的卫星轨道数据，评估碰撞风险并设计预警算法', '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', 'published', '2023-12-20 23:59:59', 100, '1. 分析两颗卫星的轨道数据\n2. 计算最近接近距离\n3. 评估碰撞概率\n4. 设计预警机制', ARRAY['使用MATLAB或Python', '提交源代码和报告', '包含算法流程图'], ARRAY['碰撞预警', '风险评估', '作业'], 'homework/collision-risk', 'assignments/collision-risk-assessment'),
('aa0e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路预算计算', '计算不同条件下的卫星通信链路预算，分析影响因素', '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', 'draft', '2023-12-25 23:59:59', 100, '1. 学习链路预算基本概念\n2. 计算上行和下行链路预算\n3. 分析不同因素的影响\n4. 优化通信性能', ARRAY['使用提供的计算模板', '考虑天气影响因素', '包含敏感性分析'], ARRAY['通信链路', '链路预算', '作业'], 'homework/link-budget', 'assignments/satellite-link-budget')
ON CONFLICT (id) DO NOTHING;

-- 插入作业提交数据
INSERT INTO submissions (id, created_at, updated_at, homework_id, student_id, status, content, file_path, gitlab_commit_sha, gitlab_branch, submitted_at, grade, feedback) VALUES
('bb0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'aa0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440007', 'submitted', 'Starlink轨道分析作业提交', '/submissions/starlink-analysis-student001.pdf', 'abc123def456', 'homework-starlink-analysis-student001', '2023-12-10 14:30:00', NULL, NULL),
('bb0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'aa0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440008', 'graded', 'Starlink轨道分析作业提交', '/submissions/starlink-analysis-student002.pdf', 'def456ghi789', 'homework-starlink-analysis-student002', '2023-12-08 16:45:00', 85, '分析思路清晰，计算过程正确，但可视化部分可以进一步改进。'),
('bb0e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), 'aa0e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440009', 'submitted', 'LEO碰撞风险评估实验提交', '/submissions/collision-risk-student003.zip', 'ghi789jkl012', 'homework-collision-risk-student003', '2023-12-12 09:15:00', NULL, NULL)
ON CONFLICT (id) DO NOTHING;

-- 插入公告数据
INSERT INTO announcements (id, created_at, updated_at, title, content, author_id, priority, is_active, valid_from, valid_to, target_roles) VALUES
('cc0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), '系统维护通知', '系统将于本周六晚上10:00-12:00进行维护升级，期间可能无法访问。请大家提前保存工作内容。', '550e8400-e29b-41d4-a716-446655440001', 'high', true, NOW(), '2023-12-31 23:59:59', ARRAY['admin', 'teacher', 'assistant', 'student']),
('cc0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), '新课题申请通知', '2024年春季学期新课题申请已开放，欢迎各位老师和同学申请。申请截止日期为12月30日。', '550e8400-e29b-41d4-a716-446655440001', 'normal', true, NOW(), '2023-12-30 23:59:59', ARRAY['teacher', 'assistant']),
('cc0e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '学期作业提交提醒', '请各位同学注意，本学期所有作业提交截止日期为12月25日，请合理安排时间。', '550e8400-e29b-41d4-a716-446655440002', 'normal', true, NOW(), '2023-12-25 23:59:59', ARRAY['student'])
ON CONFLICT (id) DO NOTHING;

-- 插入通知数据
INSERT INTO notifications (id, created_at, updated_at, type, title, content, recipient_id, sender_id, project_id, topic_id, homework_id, is_read, action_url) VALUES
('dd0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'project_invite', '项目邀请', '您被邀请加入"Starlink星座分析"项目', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, NULL, false, '/research-projects/660e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'homework_assigned', '新作业分配', '您有新的作业："Starlink轨道参数分析作业"', '550e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, 'aa0e8400-e29b-41d4-a716-446655440001', true, '/homework/aa0e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), 'topic_reply', '话题回复', '您的话题"Starlink卫星轨道参数分析方法讨论"有新回复', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440001', '880e8400-e29b-41d4-a716-446655440001', NULL, false, '/topics/880e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), 'grade_published', '成绩发布', '您的作业"Starlink轨道参数分析作业"已评分', '550e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, 'aa0e8400-e29b-41d4-a716-446655440001', false, '/homework/aa0e8400-e29b-41d4-a716-446655440001/submissions')
ON CONFLICT (id) DO NOTHING;

-- 更新序列值，确保新插入的数据不会与自动生成的ID冲突
SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users WHERE id ~ '^[0-9]+$') + 1, false);

-- 插入完成提示
DO $$
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'GitLabEx 测试数据初始化完成！';
    RAISE NOTICE '==============================================';
    RAISE NOTICE '已插入数据：';
    RAISE NOTICE '- 用户: % 个', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '- 研究课题: % 个', (SELECT COUNT(*) FROM research_projects);  
    RAISE NOTICE '- 项目成员: % 个', (SELECT COUNT(*) FROM project_members);
    RAISE NOTICE '- 话题: % 个', (SELECT COUNT(*) FROM topics);
    RAISE NOTICE '- 文档: % 个', (SELECT COUNT(*) FROM documents);
    RAISE NOTICE '- 作业: % 个', (SELECT COUNT(*) FROM homeworks);
    RAISE NOTICE '- 提交: % 个', (SELECT COUNT(*) FROM submissions);
    RAISE NOTICE '- 公告: % 个', (SELECT COUNT(*) FROM announcements);
    RAISE NOTICE '- 通知: % 个', (SELECT COUNT(*) FROM notifications);
    RAISE NOTICE '==============================================';
    RAISE NOTICE '测试账号信息：';
    RAISE NOTICE '管理员: admin / admin@gitlabex.com';
    RAISE NOTICE '教师: prof_wang / wang.prof@university.edu';
    RAISE NOTICE '学生: student_001 / student001@university.edu';
    RAISE NOTICE '==============================================';
END $$;

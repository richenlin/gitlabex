-- GitLabEx 修正版测试数据
-- 基于当前数据库结构，避免与现有数据冲突

SET CLIENT_ENCODING TO 'UTF8';

-- 插入更多用户数据（避免与现有git_lab_id冲突）
INSERT INTO users (id, created_at, updated_at, git_lab_id, username, email, name, role, edu_role, is_active, avatar_url) VALUES
-- 管理员（使用不冲突的git_lab_id）
('550e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 100, 'admin', 'admin@gitlabex.com', '系统管理员', 'admin', 50, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin'),

-- 教师团队
('550e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 101, 'prof_wang', 'wang.prof@university.edu', '王明教授', 'teacher', 40, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=wangming'),
('550e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), 102, 'prof_li', 'li.prof@university.edu', '李华教授', 'teacher', 40, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=lihua'),
('550e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), 103, 'prof_zhang', 'zhang.prof@university.edu', '张伟教授', 'teacher', 40, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhangwei'),
('550e8400-e29b-41d4-a716-446655440010', NOW(), NOW(), 104, 'prof_liu', 'liu.prof@university.edu', '刘芳教授', 'teacher', 40, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=liufang'),

-- 助教团队
('550e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), 105, 'ta_chen', 'chen.ta@university.edu', '陈小明助教', 'assistant', 30, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=chenxiaoming'),
('550e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), 106, 'ta_wu', 'wu.ta@university.edu', '吴晓丽助教', 'assistant', 30, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=wuxiaoli'),
('550e8400-e29b-41d4-a716-446655440011', NOW(), NOW(), 107, 'ta_huang', 'huang.ta@university.edu', '黄志强助教', 'assistant', 30, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=huangzhiqiang'),

-- 学生团队
('550e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), 108, 'student_001', 'student001@university.edu', '张三', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhangsan'),
('550e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), 109, 'student_002', 'student002@university.edu', '李四', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=lisi'),
('550e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), 110, 'student_003', 'student003@university.edu', '王五', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=wangwu'),
('550e8400-e29b-41d4-a716-446655440012', NOW(), NOW(), 111, 'student_004', 'student004@university.edu', '赵六', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhaoliu'),
('550e8400-e29b-41d4-a716-446655440013', NOW(), NOW(), 112, 'student_005', 'student005@university.edu', '钱七', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=qianqi'),
('550e8400-e29b-41d4-a716-446655440014', NOW(), NOW(), 113, 'student_006', 'student006@university.edu', '孙八', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=sunba'),
('550e8400-e29b-41d4-a716-446655440015', NOW(), NOW(), 114, 'student_007', 'student007@university.edu', '周九', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhoujiu'),
('550e8400-e29b-41d4-a716-446655440016', NOW(), NOW(), 115, 'student_008', 'student008@university.edu', '吴十', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=wushi'),
('550e8400-e29b-41d4-a716-446655440017', NOW(), NOW(), 116, 'student_009', 'student009@university.edu', '郑十一', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhengshiyi'),
('550e8400-e29b-41d4-a716-446655440018', NOW(), NOW(), 117, 'student_010', 'student010@university.edu', '王十二', 'student', 20, true, 'https://api.dicebear.com/7.x/avataaars/svg?seed=wangshier')
ON CONFLICT (id) DO NOTHING;

-- 插入研究课题数据（使用正确的字段名）
INSERT INTO research_projects (id, created_at, updated_at, name, description, status, creator_id, start_date, is_public, git_lab_project_id, git_lab_url) VALUES
-- 航天工程类
('660e8400-e29b-41d4-a716-446655440001', '2023-11-14 10:00:00', NOW(), 'Starlink星座分析', 'SpaceX Starlink卫星星座的轨道分析和性能评估研究，包括覆盖范围分析、星间链路优化、地面站配置等关键技术研究', 'active', '550e8400-e29b-41d4-a716-446655440002', '2023-11-14', true, 1001, 'http://localhost:8081/aerospace/starlink-analysis'),
('660e8400-e29b-41d4-a716-446655440002', '2023-11-12 09:30:00', NOW(), '卫星通信链路优化', '针对LEO/MEO/GEO不同轨道高度卫星通信链路的优化算法研究，重点解决多普勒频移补偿、功率控制、自适应调制等问题', 'active', '550e8400-e29b-41d4-a716-446655440003', '2023-11-12', true, 1002, 'http://localhost:8081/aerospace/satellite-comm-optimization'),
('660e8400-e29b-41d4-a716-446655440003', '2023-11-09 14:20:00', NOW(), '轨道机动计算器', '卫星轨道机动计算与仿真工具开发，支持霍曼转移、双椭圆转移、平面变换等多种机动策略', 'active', '550e8400-e29b-41d4-a716-446655440002', '2023-11-09', true, 1003, 'http://localhost:8081/aerospace/orbital-maneuver-calculator'),
('660e8400-e29b-41d4-a716-446655440004', '2023-11-08 11:45:00', NOW(), '火箭回收技术研究', 'SpaceX猎鹰9号式火箭垂直回收技术的建模与仿真分析，包括制导导航控制算法设计', 'active', '550e8400-e29b-41d4-a716-446655440004', '2023-11-08', true, 1004, 'http://localhost:8081/aerospace/rocket-recovery'),
('660e8400-e29b-41d4-a716-446655440005', '2023-11-13 16:15:00', NOW(), 'LEO轨道碰撞预警', '低轨卫星碰撞风险评估与预警系统，基于机器学习的轨道预测和碰撞概率计算', 'active', '550e8400-e29b-41d4-a716-446655440003', '2023-11-13', false, 1005, 'http://localhost:8081/aerospace/leo-collision-warning'),

-- 人工智能类
('660e8400-e29b-41d4-a716-446655440006', '2023-11-10 13:30:00', NOW(), '深度学习图像识别', '基于卷积神经网络的卫星图像目标识别系统，应用于地面目标检测和分类', 'active', '550e8400-e29b-41d4-a716-446655440010', '2023-11-10', true, 1006, 'http://localhost:8081/ai/satellite-image-recognition'),
('660e8400-e29b-41d4-a716-446655440007', '2023-11-07 15:20:00', NOW(), '自然语言处理平台', '多语言文本分析和情感分析平台开发，支持中英文混合文本处理', 'active', '550e8400-e29b-41d4-a716-446655440010', '2023-11-07', true, 1007, 'http://localhost:8081/ai/nlp-platform'),
('660e8400-e29b-41d4-a716-446655440008', '2023-11-05 10:15:00', NOW(), '强化学习游戏AI', '基于深度Q网络的游戏AI智能体开发，实现自主学习和策略优化', 'active', '550e8400-e29b-41d4-a716-446655440010', '2023-11-05', true, 1008, 'http://localhost:8081/ai/reinforcement-learning-game'),

-- 软件工程类
('660e8400-e29b-41d4-a716-446655440009', '2023-11-06 09:00:00', NOW(), '微服务架构实践', '基于Spring Cloud的微服务架构设计与实现，包括服务发现、负载均衡、熔断器等组件', 'active', '550e8400-e29b-41d4-a716-446655440004', '2023-11-06', true, 1009, 'http://localhost:8081/software/microservices-practice'),
('660e8400-e29b-41d4-a716-446655440010', '2023-11-04 14:40:00', NOW(), 'React全栈开发', '基于React + Node.js的全栈Web应用开发，集成Redux状态管理和JWT认证', 'active', '550e8400-e29b-41d4-a716-446655440004', '2023-11-04', true, 1010, 'http://localhost:8081/software/react-fullstack'),

-- 数据科学类
('660e8400-e29b-41d4-a716-446655440011', '2023-11-03 11:25:00', NOW(), '大数据分析平台', '基于Hadoop和Spark的大数据处理平台构建，支持实时数据流处理和批量数据分析', 'active', '550e8400-e29b-41d4-a716-446655440010', '2023-11-03', true, 1011, 'http://localhost:8081/data/big-data-platform'),
('660e8400-e29b-41d4-a716-446655440012', '2023-11-01 16:50:00', NOW(), '机器学习预测模型', '时间序列预测和回归分析模型开发，应用于股票价格和气象数据预测', 'active', '550e8400-e29b-41d4-a716-446655440010', '2023-11-01', true, 1012, 'http://localhost:8081/data/ml-prediction-models')
ON CONFLICT (id) DO NOTHING;

-- 插入项目成员关系
INSERT INTO project_members (id, created_at, updated_at, project_id, user_id, role, joined_at) VALUES
-- Starlink项目成员
('770e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'owner', NOW()),
('770e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440005', 'maintainer', NOW()),
('770e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440007', 'developer', NOW()),
('770e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440008', 'developer', NOW()),
('770e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440009', 'reporter', NOW()),

-- 卫星通信项目成员
('770e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', 'owner', NOW()),
('770e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440006', 'maintainer', NOW()),
('770e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440012', 'developer', NOW()),
('770e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440013', 'developer', NOW()),

-- AI图像识别项目成员
('770e8400-e29b-41d4-a716-446655440010', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440010', 'owner', NOW()),
('770e8400-e29b-41d4-a716-446655440011', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440011', 'maintainer', NOW()),
('770e8400-e29b-41d4-a716-446655440012', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440014', 'developer', NOW()),
('770e8400-e29b-41d4-a716-446655440013', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440015', 'developer', NOW()),
('770e8400-e29b-41d4-a716-446655440014', NOW(), NOW(), '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440016', 'reporter', NOW())
ON CONFLICT (id) DO NOTHING;

-- 插入丰富的话题讨论数据（移除不存在的comments_count字段）
INSERT INTO topics (id, created_at, updated_at, title, content, project_id, author_id, status, priority, tags, view_count, like_count) VALUES
-- 航天工程相关话题
('880e8400-e29b-41d4-a716-446655440001', '2023-11-15 10:30:00', NOW(), 'Starlink卫星轨道参数分析方法讨论', '大家对于Starlink卫星轨道参数分析有什么好的方法推荐吗？特别是在轨道衰减预测方面。我目前使用的是SGP4模型，但感觉在长期预测上精度不够。有没有人尝试过使用机器学习方法来改进预测精度？', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440007', 'active', 'normal', ARRAY['轨道分析', 'Starlink', '卫星', 'SGP4'], 256, 12),
('880e8400-e29b-41d4-a716-446655440002', '2023-11-14 14:20:00', NOW(), 'LEO卫星碰撞概率计算模型', '关于低轨卫星碰撞概率计算，我们应该采用哪种数学模型比较准确？目前看到有蒙特卡洛方法、解析方法等，各有什么优缺点？在实际工程应用中哪种更实用？', '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440008', 'active', 'high', ARRAY['碰撞预警', 'LEO', '概率计算', '蒙特卡洛'], 189, 8),
('880e8400-e29b-41d4-a716-446655440003', '2023-11-13 16:45:00', NOW(), '卫星通信链路性能优化策略', '针对不同轨道高度的卫星通信链路，有哪些有效的性能优化策略？我在研究LEO卫星通信时发现多普勒频移是个大问题，大家是怎么解决的？', '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440012', 'active', 'normal', ARRAY['通信链路', '性能优化', '卫星通信', '多普勒'], 312, 15),
('880e8400-e29b-41d4-a716-446655440004', '2023-11-12 11:15:00', NOW(), '火箭垂直回收控制算法讨论', 'SpaceX的火箭回收技术很厉害，想了解一下垂直着陆的控制算法。是否有开源的仿真平台可以学习？PID控制够用吗还是需要更高级的控制方法？', '660e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440009', 'active', 'normal', ARRAY['火箭回收', '控制算法', 'SpaceX', 'PID'], 145, 9),

-- AI相关话题
('880e8400-e29b-41d4-a716-446655440005', '2023-11-11 09:20:00', NOW(), 'CNN在卫星图像识别中的应用', '最近在做卫星图像目标检测，想问问大家用哪种CNN架构效果比较好？ResNet、VGG还是更新的EfficientNet？数据增强有什么好的策略？', '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440014', 'active', 'normal', ARRAY['CNN', '图像识别', '卫星图像', 'ResNet'], 203, 11),
('880e8400-e29b-41d4-a716-446655440006', '2023-11-10 15:30:00', NOW(), 'BERT模型在中文文本分析中的优化', '在做中文情感分析项目，发现BERT模型在某些领域词汇上效果不太好。有没有人尝试过领域自适应的方法？或者推荐其他适合中文的预训练模型？', '660e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440015', 'active', 'normal', ARRAY['BERT', '中文NLP', '情感分析', '预训练模型'], 178, 6),
('880e8400-e29b-41d4-a716-446655440007', '2023-11-09 13:45:00', NOW(), '强化学习在游戏AI中的调参经验', 'DQN训练游戏AI时总是不收敛，reward设计和网络结构都检查过了。有经验的同学能分享一下调参技巧吗？特别是学习率和探索策略的设置。', '660e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440016', 'active', 'high', ARRAY['强化学习', 'DQN', '游戏AI', '调参'], 234, 13),

-- 软件工程相关话题
('880e8400-e29b-41d4-a716-446655440008', '2023-11-08 10:50:00', NOW(), 'Spring Cloud微服务拆分原则', '在设计微服务架构时，服务拆分的粒度一直很难把握。太细了管理复杂，太粗了又失去了微服务的优势。大家有什么好的拆分原则和实践经验？', '660e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440017', 'active', 'normal', ARRAY['微服务', 'Spring Cloud', '架构设计', '服务拆分'], 167, 8),
('880e8400-e29b-41d4-a716-446655440009', '2023-11-07 14:25:00', NOW(), 'React Hooks最佳实践分享', '从Class组件迁移到Hooks后，发现性能优化和状态管理变得更复杂了。useCallback、useMemo什么时候用？自定义Hook怎么设计比较好？', '660e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440018', 'active', 'normal', ARRAY['React', 'Hooks', '性能优化', '状态管理'], 198, 10),

-- 数据科学相关话题
('880e8400-e29b-41d4-a716-446655440010', '2023-11-06 16:10:00', NOW(), 'Spark大数据处理性能调优', '在处理TB级数据时Spark作业经常OOM或者运行很慢，除了调整executor内存还有什么优化策略？数据倾斜问题怎么解决比较好？', '660e8400-e29b-41d4-a716-446655440011', '550e8400-e29b-41d4-a716-446655440013', 'active', 'high', ARRAY['Spark', '大数据', '性能调优', '数据倾斜'], 289, 14),
('880e8400-e29b-41d4-a716-446655440011', '2023-11-05 12:35:00', NOW(), '时间序列预测模型选择', '在做股票价格预测，尝试了ARIMA、LSTM等模型，效果都不理想。有没有人用过Transformer或者其他新的时间序列模型？特征工程有什么技巧？', '660e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440012', 'active', 'normal', ARRAY['时间序列', '股票预测', 'LSTM', 'Transformer'], 156, 7),

-- 一般性技术讨论
('880e8400-e29b-41d4-a716-446655440012', '2023-11-04 09:15:00', NOW(), '开源项目贡献经验分享', '想为开源项目做贡献但不知道从哪里开始，有经验的同学能分享一下如何选择项目、如何提交PR、需要注意什么吗？', NULL, '550e8400-e29b-41d4-a716-446655440014', 'active', 'low', ARRAY['开源', 'GitHub', 'PR', '贡献'], 134, 9),
('880e8400-e29b-41d4-a716-446655440013', '2023-11-03 17:40:00', NOW(), '技术面试准备攻略', '马上要开始找实习了，想问问大家技术面试都会考什么？算法题、系统设计、项目经验哪个更重要？有什么好的准备资料推荐？', NULL, '550e8400-e29b-41d4-a716-446655440015', 'active', 'normal', ARRAY['面试', '算法', '系统设计', '实习'], 278, 18)
ON CONFLICT (id) DO NOTHING;

-- 插入更多文档数据
INSERT INTO documents (id, created_at, updated_at, title, description, file_path, file_size, file_type, mime_type, status, uploader_id, project_id, category, tags, download_count, auto_indexed) VALUES
-- 航天工程文档
('990e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink轨道分析报告.pdf', 'SpaceX Starlink卫星星座轨道特性分析的详细报告，包含轨道参数计算、覆盖范围分析和星间链路设计', '/documents/aerospace/starlink-orbit-analysis.pdf', 2048576, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', '研究报告', ARRAY['Starlink', '轨道分析', '卫星星座'], 45, true),
('990e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO碰撞预警算法设计文档.docx', '低轨卫星碰撞预警系统算法设计与实现文档，详细描述了碰撞概率计算模型和预警机制', '/documents/aerospace/leo-collision-algorithm.docx', 1536000, 'word', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'approved', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440005', '技术文档', ARRAY['LEO', '碰撞预警', '算法设计'], 23, true),
('990e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路建模手册.pdf', '卫星通信链路数学建模和仿真方法手册，涵盖信道建模、功率预算、调制解调等关键技术', '/documents/aerospace/satellite-comm-modeling.pdf', 3072000, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440002', '技术手册', ARRAY['卫星通信', '链路建模', '仿真'], 67, true),
('990e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '火箭回收制导控制算法.pdf', 'SpaceX猎鹰9号火箭垂直回收的制导导航控制算法研究，包含仿真验证结果', '/documents/aerospace/rocket-recovery-control.pdf', 1843200, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440004', '研究报告', ARRAY['火箭回收', '制导控制', 'SpaceX'], 34, true),
('990e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), '轨道机动策略对比分析.xlsx', '不同轨道机动策略的燃料消耗和时间成本对比分析数据表', '/documents/aerospace/orbital-maneuver-comparison.xlsx', 512000, 'excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 'approved', '550e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440003', '数据分析', ARRAY['轨道机动', '燃料优化', '对比分析'], 19, true),

-- AI/机器学习文档
('990e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), 'CNN卫星图像识别模型训练指南.pdf', '基于卷积神经网络的卫星图像目标识别模型训练完整指南，包含数据预处理、模型设计、训练技巧等', '/documents/ai/cnn-satellite-image-guide.pdf', 2560000, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440010', '660e8400-e29b-41d4-a716-446655440006', '技术指南', ARRAY['CNN', '图像识别', '卫星图像', '深度学习'], 89, true),
('990e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), 'BERT中文文本分析实践.docx', 'BERT模型在中文文本情感分析中的应用实践，包含模型微调和性能优化方法', '/documents/ai/bert-chinese-text-analysis.docx', 1280000, 'word', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'approved', '550e8400-e29b-41d4-a716-446655440010', '660e8400-e29b-41d4-a716-446655440007', '实践指南', ARRAY['BERT', 'NLP', '中文处理', '情感分析'], 56, true),
('990e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), '强化学习DQN算法实现.py', 'Deep Q-Network强化学习算法的Python实现代码，包含完整的训练和测试流程', '/documents/ai/dqn-implementation.py', 45600, 'code', 'text/x-python', 'approved', '550e8400-e29b-41d4-a716-446655440011', '660e8400-e29b-41d4-a716-446655440008', '源代码', ARRAY['强化学习', 'DQN', 'Python', '游戏AI'], 78, true),

-- 软件工程文档
('990e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), 'Spring Cloud微服务架构设计.pdf', 'Spring Cloud微服务架构设计最佳实践，包含服务拆分原则、配置管理、服务发现等', '/documents/software/spring-cloud-architecture.pdf', 1920000, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440009', '架构设计', ARRAY['微服务', 'Spring Cloud', '架构设计', '最佳实践'], 92, true),
('990e8400-e29b-41d4-a716-446655440010', NOW(), NOW(), 'React Hooks开发规范.md', 'React Hooks开发规范和最佳实践指南，包含性能优化和常见陷阱避免方法', '/documents/software/react-hooks-best-practices.md', 28800, 'other', 'text/markdown', 'approved', '550e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440010', '开发规范', ARRAY['React', 'Hooks', '开发规范', '最佳实践'], 123, true),

-- 数据科学文档
('990e8400-e29b-41d4-a716-446655440011', NOW(), NOW(), 'Spark大数据处理优化手册.pdf', 'Apache Spark大数据处理性能优化完整手册，涵盖内存管理、数据倾斜处理、作业调优等', '/documents/data/spark-optimization-handbook.pdf', 3584000, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440010', '660e8400-e29b-41d4-a716-446655440011', '技术手册', ARRAY['Spark', '大数据', '性能优化', '调优'], 145, true),
('990e8400-e29b-41d4-a716-446655440012', NOW(), NOW(), '时间序列预测模型对比.xlsx', '多种时间序列预测模型在股票价格预测中的性能对比分析数据', '/documents/data/time-series-model-comparison.xlsx', 768000, 'excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 'approved', '550e8400-e29b-41d4-a716-446655440010', '660e8400-e29b-41d4-a716-446655440012', '数据分析', ARRAY['时间序列', '预测模型', '股票分析', '对比研究'], 67, true),

-- 通用技术文档（project_id为NULL的通用文档）
('990e8400-e29b-41d4-a716-446655440013', NOW(), NOW(), 'Git协作开发流程规范.pdf', 'Git版本控制和团队协作开发流程规范，包含分支管理、代码审查、发布流程等', '/documents/general/git-workflow-guide.pdf', 1024000, 'pdf', 'application/pdf', 'approved', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440009', '开发规范', ARRAY['Git', '协作开发', '版本控制', '流程规范'], 234, true),
('990e8400-e29b-41d4-a716-446655440014', NOW(), NOW(), '技术面试题库.docx', '计算机科学技术面试常见题目汇总，包含算法、系统设计、数据库等各个方向', '/documents/general/tech-interview-questions.docx', 896000, 'word', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'approved', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440010', '学习资料', ARRAY['面试', '算法题', '系统设计', '技术问答'], 456, true),
('990e8400-e29b-41d4-a716-446655440015', NOW(), NOW(), '开源项目贡献指南.md', '如何为开源项目做贡献的详细指南，从选择项目到提交PR的完整流程', '/documents/general/open-source-contribution-guide.md', 34560, 'other', 'text/markdown', 'approved', '550e8400-e29b-41d4-a716-446655440006', '660e8400-e29b-41d4-a716-446655440011', '开发指南', ARRAY['开源', 'GitHub', 'PR', '贡献指南'], 189, true)
ON CONFLICT (id) DO NOTHING;

-- 插入作业数据
INSERT INTO homeworks (id, created_at, updated_at, title, description, project_id, creator_id, status, due_date, max_grade, instructions, requirements, tags) VALUES
-- 航天工程作业
('aa0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), 'Starlink轨道参数分析作业', '分析SpaceX Starlink卫星的轨道参数，计算其覆盖范围并评估星座性能', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'published', '2023-12-15 23:59:59', 100, 
'1. 下载最新的Starlink TLE数据\n2. 使用SGP4模型计算轨道参数\n3. 分析全球覆盖特性\n4. 计算星间链路可见性\n5. 提交详细分析报告', 
ARRAY['使用Python或MATLAB进行计算', '报告格式为PDF，不少于10页', '包含可视化图表和轨道仿真结果', '需要附上源代码'], 
ARRAY['轨道分析', '作业', 'Starlink']),

('aa0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), 'LEO碰撞风险评估实验', '基于给定的卫星轨道数据，评估碰撞风险并设计预警算法', '660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', 'published', '2023-12-20 23:59:59', 100, 
'1. 分析两颗卫星的轨道数据\n2. 计算最近接近距离和时间\n3. 评估碰撞概率\n4. 设计自动预警机制\n5. 验证算法有效性', 
ARRAY['使用MATLAB或Python实现', '提交源代码和详细报告', '包含算法流程图和仿真结果', '报告需包含误差分析'], 
ARRAY['碰撞预警', '风险评估', '作业']),

('aa0e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '卫星通信链路预算计算', '针对特定卫星通信场景，完成完整的链路预算计算和性能分析', '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', 'published', '2023-12-18 23:59:59', 100, 
'1. 选择具体的卫星通信场景\n2. 计算上行链路和下行链路预算\n3. 分析雨衰、大气损耗等影响因素\n4. 优化系统参数\n5. 提交计算报告', 
ARRAY['需要考虑实际工程约束', '使用标准的链路预算公式', '报告包含敏感性分析', '提供Excel计算表格'], 
ARRAY['卫星通信', '链路预算', '作业']),

-- AI/机器学习作业
('aa0e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '卫星图像目标检测实验', '使用深度学习方法实现卫星图像中的目标检测，识别建筑物、道路、车辆等', '660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440010', 'published', '2023-12-22 23:59:59', 100, 
'1. 收集和标注卫星图像数据集\n2. 设计CNN模型架构\n3. 训练目标检测模型\n4. 评估模型性能\n5. 优化检测精度', 
ARRAY['使用PyTorch或TensorFlow框架', '数据集不少于1000张图像', '模型准确率需达到85%以上', '提交训练代码和模型文件'], 
ARRAY['深度学习', '目标检测', '卫星图像', '作业']),

('aa0e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), '中文情感分析系统开发', '基于BERT模型开发中文文本情感分析系统，支持多分类情感识别', '660e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440010', 'published', '2023-12-25 23:59:59', 100, 
'1. 准备中文情感分析数据集\n2. 微调预训练BERT模型\n3. 实现多分类情感识别\n4. 开发Web界面\n5. 部署和测试系统', 
ARRAY['使用transformers库', '支持积极、消极、中性三分类', 'Web界面使用Flask或Django', '系统响应时间小于2秒'], 
ARRAY['NLP', '情感分析', 'BERT', '作业']),

-- 软件工程作业
('aa0e8400-e29b-41d4-a716-446655440006', NOW(), NOW(), '微服务电商系统设计', '使用Spring Cloud设计和实现一个简化版的微服务电商系统', '660e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440004', 'published', '2023-12-28 23:59:59', 100, 
'1. 设计微服务架构\n2. 实现用户、商品、订单等核心服务\n3. 集成服务发现和配置中心\n4. 实现API网关\n5. 添加监控和日志', 
ARRAY['使用Spring Cloud全家桶', '至少包含5个微服务', '支持Docker容器化部署', '提供完整的API文档'], 
ARRAY['微服务', 'Spring Cloud', '系统设计', '作业']),

('aa0e8400-e29b-41d4-a716-446655440007', NOW(), NOW(), 'React全栈项目开发', '使用React + Node.js开发一个完整的全栈Web应用', '660e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440004', 'published', '2023-12-30 23:59:59', 100, 
'1. 设计应用功能和界面\n2. 实现React前端\n3. 开发Node.js后端API\n4. 集成数据库\n5. 实现用户认证和授权', 
ARRAY['前端使用React Hooks', '后端使用Express框架', '数据库使用MongoDB或MySQL', '实现响应式设计'], 
ARRAY['React', 'Node.js', '全栈开发', '作业']),

-- 数据科学作业
('aa0e8400-e29b-41d4-a716-446655440008', NOW(), NOW(), '大数据分析平台搭建', '使用Hadoop和Spark搭建大数据分析平台，处理和分析大规模数据集', '660e8400-e29b-41d4-a716-446655440011', '550e8400-e29b-41d4-a716-446655440010', 'published', '2024-01-05 23:59:59', 100, 
'1. 搭建Hadoop集群环境\n2. 配置Spark计算引擎\n3. 实现数据ETL流程\n4. 开发数据分析应用\n5. 生成分析报告', 
ARRAY['使用Docker搭建集群', '处理数据量不少于1GB', '实现实时和批处理两种模式', '提供数据可视化界面'], 
ARRAY['大数据', 'Hadoop', 'Spark', '作业']),

('aa0e8400-e29b-41d4-a716-446655440009', NOW(), NOW(), '股票价格预测模型', '开发多种机器学习模型预测股票价格，比较不同算法的预测效果', '660e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440010', 'published', '2024-01-08 23:59:59', 100, 
'1. 收集股票历史数据\n2. 进行特征工程\n3. 实现多种预测模型\n4. 比较模型性能\n5. 分析预测结果', 
ARRAY['至少实现3种不同算法', '使用交叉验证评估性能', '包含LSTM时间序列模型', '提供模型解释性分析'], 
ARRAY['机器学习', '股票预测', '时间序列', '作业'])
ON CONFLICT (id) DO NOTHING;

-- 插入公告数据
INSERT INTO announcements (id, created_at, updated_at, title, content, author_id, priority, is_active, valid_from, valid_to, target_roles) VALUES
('cc0e8400-e29b-41d4-a716-446655440001', NOW(), NOW(), '系统维护通知', '系统将于本周六晚上22:00-24:00进行维护升级，期间可能无法访问。请大家提前保存工作内容。维护内容包括数据库优化、新功能上线等。', '550e8400-e29b-41d4-a716-446655440001', 'high', true, NOW(), '2024-01-31 23:59:59', ARRAY['admin', 'teacher', 'assistant', 'student']),
('cc0e8400-e29b-41d4-a716-446655440002', NOW(), NOW(), '2024春季学期课题申请开放', '2024年春季学期新课题申请已开放，欢迎各位老师和同学申请。申请截止日期为1月15日。优秀课题将获得资金支持。', '550e8400-e29b-41d4-a716-446655440001', 'medium', true, NOW(), '2024-01-15 23:59:59', ARRAY['teacher', 'assistant']),
('cc0e8400-e29b-41d4-a716-446655440003', NOW(), NOW(), '学术会议征稿通知', '第十届航天工程国际会议现征集论文，截止日期2024年2月29日。鼓励师生积极投稿，优秀论文将推荐至期刊发表。', '550e8400-e29b-41d4-a716-446655440002', 'medium', true, NOW(), '2024-02-29 23:59:59', ARRAY['teacher', 'assistant', 'student']),
('cc0e8400-e29b-41d4-a716-446655440004', NOW(), NOW(), '寒假实习招聘信息', '多家知名科技公司提供寒假实习岗位，包括SpaceX、华为、腾讯等。有意向的同学请关注就业指导中心通知。', '550e8400-e29b-41d4-a716-446655440001', 'low', true, NOW(), '2024-01-20 23:59:59', ARRAY['student']),
('cc0e8400-e29b-41d4-a716-446655440005', NOW(), NOW(), '图书馆数据库访问升级', '图书馆已升级IEEE、ACM等学术数据库访问权限，新增多个AI和航天领域的专业数据库。欢迎师生使用。', '550e8400-e29b-41d4-a716-446655440001', 'low', true, NOW(), '2024-06-30 23:59:59', ARRAY['teacher', 'assistant', 'student'])
ON CONFLICT (id) DO NOTHING;

-- 插入丰富的通知数据
INSERT INTO notifications (id, created_at, updated_at, type, title, content, recipient_id, sender_id, project_id, topic_id, homework_id, is_read, action_url) VALUES
-- 项目相关通知
('dd0e8400-e29b-41d4-a716-446655440001', '2023-11-16 10:30:00', NOW(), 'project', '项目邀请', '您被邀请加入"Starlink星座分析"项目，担任开发者角色', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, NULL, true, '/scenes/660e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440002', '2023-11-15 14:20:00', NOW(), 'project', '项目更新', 'Starlink星座分析项目有新的文档上传', '550e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, NULL, false, '/scenes/660e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440003', '2023-11-14 16:45:00', NOW(), 'project', '成员加入', '张三已加入LEO轨道碰撞预警项目', '550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440007', '660e8400-e29b-41d4-a716-446655440005', NULL, NULL, true, '/scenes/660e8400-e29b-41d4-a716-446655440005'),

-- 作业相关通知
('dd0e8400-e29b-41d4-a716-446655440004', '2023-11-13 09:15:00', NOW(), 'homework', '新作业分配', '您有新的作业："Starlink轨道参数分析作业"，截止日期：12月15日', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, 'aa0e8400-e29b-41d4-a716-446655440001', true, '/homeworks/aa0e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440005', '2023-11-12 11:30:00', NOW(), 'homework', '作业提醒', 'LEO碰撞风险评估实验作业即将截止，请及时提交', '550e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440005', NULL, 'aa0e8400-e29b-41d4-a716-446655440002', false, '/homeworks/aa0e8400-e29b-41d4-a716-446655440002'),

-- 话题相关通知
('dd0e8400-e29b-41d4-a716-446655440006', '2023-11-11 13:45:00', NOW(), 'topic', '话题回复', '您的话题"Starlink卫星轨道参数分析方法讨论"有新回复', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440008', '660e8400-e29b-41d4-a716-446655440001', '880e8400-e29b-41d4-a716-446655440001', NULL, true, '/topics/880e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440007', '2023-11-10 10:25:00', NOW(), 'topic', '话题点赞', '您的话题"CNN在卫星图像识别中的应用"收到了点赞', '550e8400-e29b-41d4-a716-446655440014', '550e8400-e29b-41d4-a716-446655440015', '660e8400-e29b-41d4-a716-446655440006', '880e8400-e29b-41d4-a716-446655440005', NULL, false, '/topics/880e8400-e29b-41d4-a716-446655440005'),

-- 系统通知
('dd0e8400-e29b-41d4-a716-446655440008', '2023-11-09 08:00:00', NOW(), 'system', '系统更新', '系统已更新到v2.1.0，新增文档预览功能和实时通知', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440001', NULL, NULL, NULL, true, '/'),
('dd0e8400-e29b-41d4-a716-446655440009', '2023-11-08 18:30:00', NOW(), 'system', '安全提醒', '检测到异常登录尝试，请注意账户安全', '550e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440001', NULL, NULL, NULL, false, '/settings'),

-- 文档相关通知
('dd0e8400-e29b-41d4-a716-446655440010', '2023-11-07 14:15:00', NOW(), 'document', '文档分享', '王明教授分享了文档"Starlink轨道分析报告.pdf"', '550e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', NULL, NULL, true, '/documents/990e8400-e29b-41d4-a716-446655440001'),
('dd0e8400-e29b-41d4-a716-446655440011', '2023-11-06 16:40:00', NOW(), 'document', '文档更新', 'LEO碰撞预警算法设计文档已更新', '550e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440005', NULL, NULL, false, '/documents/990e8400-e29b-41d4-a716-446655440002'),

-- 更多通知
('dd0e8400-e29b-41d4-a716-446655440012', '2023-11-05 12:20:00', NOW(), 'info', '会议通知', '航天工程研讨会将于下周三举行，欢迎参加', '550e8400-e29b-41d4-a716-446655440014', '550e8400-e29b-41d4-a716-446655440002', NULL, NULL, NULL, true, '/'),
('dd0e8400-e29b-41d4-a716-446655440013', '2023-11-04 09:50:00', NOW(), 'achievement', '成就解锁', '恭喜！您已完成5个项目，解锁"项目达人"成就', '550e8400-e29b-41d4-a716-446655440015', '550e8400-e29b-41d4-a716-446655440001', NULL, NULL, NULL, false, '/profile'),
('dd0e8400-e29b-41d4-a716-446655440014', '2023-11-03 17:30:00', NOW(), 'reminder', '截止提醒', '您有3个作业即将截止，请及时完成', '550e8400-e29b-41d4-a716-446655440016', '550e8400-e29b-41d4-a716-446655440001', NULL, NULL, NULL, false, '/homeworks')
ON CONFLICT (id) DO NOTHING;

-- 显示插入结果统计
DO $$
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'GitLabEx 修正版测试数据初始化完成！';
    RAISE NOTICE '==============================================';
    RAISE NOTICE '已插入数据统计：';
    RAISE NOTICE '- 用户: % 个', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '- 研究课题: % 个', (SELECT COUNT(*) FROM research_projects);  
    RAISE NOTICE '- 项目成员: % 个', (SELECT COUNT(*) FROM project_members);
    RAISE NOTICE '- 话题讨论: % 个', (SELECT COUNT(*) FROM topics);
    RAISE NOTICE '- 文档资料: % 个', (SELECT COUNT(*) FROM documents);
    RAISE NOTICE '- 课程作业: % 个', (SELECT COUNT(*) FROM homeworks);
    RAISE NOTICE '- 系统公告: % 个', (SELECT COUNT(*) FROM announcements);
    RAISE NOTICE '- 用户通知: % 个', (SELECT COUNT(*) FROM notifications);
    RAISE NOTICE '==============================================';
    RAISE NOTICE '测试账号信息：';
    RAISE NOTICE '管理员: admin / admin@gitlabex.com';
    RAISE NOTICE '教师: prof_wang / wang.prof@university.edu';
    RAISE NOTICE '教师: prof_li / li.prof@university.edu';
    RAISE NOTICE '助教: ta_chen / chen.ta@university.edu';
    RAISE NOTICE '学生: student_001 / student001@university.edu';
    RAISE NOTICE '学生: student_002 / student002@university.edu';
    RAISE NOTICE '==============================================';
    RAISE NOTICE '项目涵盖领域：';
    RAISE NOTICE '- 航天工程 (轨道分析、卫星通信、碰撞预警等)';
    RAISE NOTICE '- 人工智能 (深度学习、NLP、强化学习等)';
    RAISE NOTICE '- 软件工程 (微服务、全栈开发等)';
    RAISE NOTICE '- 数据科学 (大数据、机器学习等)';
    RAISE NOTICE '==============================================';
END $$;

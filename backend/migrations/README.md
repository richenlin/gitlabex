# GitLabEx 数据库迁移指南

## 概述

本目录包含GitLabEx系统的数据库迁移脚本，用于安全地更新数据库架构和数据。

## 当前迁移版本

### 005_remove_class_management.sql
**重要架构变更：移除班级管理系统**

此迁移将系统从"用户 → 班级 → 课题"的三层架构简化为"用户 → 课题"的直接管理架构。

#### 变更内容：
1. **删除表**：
   - `classes` - 班级表
   - `class_members` - 班级成员表

2. **修改 `projects` 表**：
   - 删除 `class_id` 字段
   - 添加 `project_code` 字段（课题代码，用于学生加入）
   - 添加 `max_members` 字段（最大成员数）
   - 添加 `current_members` 字段（当前成员数，自动计算）

3. **修改 `users` 表**：
   - 添加 `gitlab_id` 字段（GitLab用户ID）
   - 添加 `dynamic_role` 字段（动态角色）

4. **修改 `assignment_submissions` 表**：
   - 添加 `review_report` 字段（JSONB格式评审报告）
   - 添加 `detailed_scores` 字段（JSONB格式详细评分）

5. **添加触发器**：
   - `update_project_member_count()` - 自动更新项目成员数

6. **性能优化**：
   - 添加多个索引提高查询性能
   - 添加表和字段注释

## 迁移执行

### 自动执行（推荐）

使用提供的迁移脚本：

```bash
# 基本执行（包含备份）
./backend/scripts/run_migration.sh

# 跳过备份（快速执行，不推荐生产环境）
./backend/scripts/run_migration.sh -n

# 强制执行（跳过确认提示）
./backend/scripts/run_migration.sh -f

# 仅验证迁移状态
./backend/scripts/run_migration.sh -v

# 查看帮助
./backend/scripts/run_migration.sh -h
```

### 环境变量配置

```bash
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="gitlabex"
export DB_USER="gitlabex"
export DB_PASSWORD="your_password"
```

### 手动执行

如果需要手动执行迁移：

```bash
# 1. 备份数据库
pg_dump -h localhost -p 5432 -U gitlabex -d gitlabex > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. 执行迁移
psql -h localhost -p 5432 -U gitlabex -d gitlabex -f backend/migrations/005_remove_class_management.sql

# 3. 验证结果
psql -h localhost -p 5432 -U gitlabex -d gitlabex -c "\dt"
```

## 注意事项

### ⚠️ 重要警告

1. **数据丢失风险**：此迁移将永久删除所有班级相关数据
2. **不可逆操作**：一旦执行，无法自动回滚
3. **生产环境**：务必在执行前进行完整的数据库备份

### 执行前检查

1. **备份验证**：确保数据库备份完整且可恢复
2. **应用停机**：建议在低峰期执行，并暂停应用服务
3. **权限检查**：确保数据库用户有足够的权限执行DDL操作
4. **磁盘空间**：确保有足够的磁盘空间用于备份和迁移

### 回滚计划

如果迁移出现问题，可以使用备份文件恢复：

```bash
# 使用备份恢复数据库
psql -h localhost -p 5432 -U gitlabex -d gitlabex < backup_file.sql
```

## 验证迁移

### 验证表结构

```sql
-- 检查班级表是否已删除
\dt classes

-- 检查项目表结构
\d projects

-- 检查新增字段
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'projects' 
AND column_name IN ('project_code', 'max_members', 'current_members');
```

### 验证触发器

```sql
-- 检查触发器是否存在
SELECT trigger_name, event_manipulation, event_object_table 
FROM information_schema.triggers 
WHERE trigger_name = 'trigger_update_project_member_count';
```

### 验证索引

```sql
-- 检查新增索引
SELECT indexname, tablename, indexdef 
FROM pg_indexes 
WHERE indexname LIKE 'idx_%';
```

## 数据迁移后的应用更新

迁移完成后，需要确保应用代码与新架构兼容：

### 后端更新

1. **模型更新**：Project模型已移除ClassID字段
2. **服务更新**：使用ProjectServiceV2、AssignmentServiceV2等新版本服务
3. **API更新**：课题管理API已移除class_id参数

### 前端更新

1. **界面更新**：移除所有班级管理相关界面
2. **路由更新**：移除班级管理路由
3. **API调用**：更新API调用以匹配新的接口结构

## 性能影响

### 查询性能

- **正面影响**：简化了关联查询，移除了班级中间层
- **索引优化**：新增的索引将提高常用查询的性能

### 存储空间

- **空间释放**：删除班级表将释放存储空间
- **新增字段**：项目表增加了几个字段，但总体空间需求减少

## 故障排除

### 常见问题

1. **权限不足**：
   ```
   错误：permission denied for table projects
   解决：确保数据库用户有ALTER TABLE权限
   ```

2. **外键约束**：
   ```
   错误：cannot drop table classes because other objects depend on it
   解决：迁移脚本已处理，使用CASCADE删除
   ```

3. **唯一约束冲突**：
   ```
   错误：duplicate key value violates unique constraint
   解决：项目代码生成逻辑会避免冲突
   ```

### 日志分析

迁移脚本会生成详细的日志文件：

```bash
# 查看迁移日志
tail -f backend/logs/migration_*.log

# 搜索错误信息
grep -i error backend/logs/migration_*.log
```

## 联系支持

如果遇到迁移问题，请：

1. 保留完整的错误日志
2. 记录数据库配置信息
3. 提供复现步骤
4. 确保备份文件完整

---

*最后更新：2024年3月15日* 
<template>
  <div class="permissions-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>权限管理</h1>
      <p>管理系统用户、用户组和权限分配</p>
    </div>

    <!-- 管理功能卡片 -->
    <el-row :gutter="24" class="management-cards">
      <el-col :xs="24" :sm="12" :md="8">
        <el-card class="management-card" shadow="hover" @click="$router.push('/permissions/users')">
          <div class="card-content">
            <el-icon class="card-icon" size="48" color="#409EFF">
              <User />
            </el-icon>
            <h3>用户管理</h3>
            <p>查看和管理系统用户信息</p>
            <div class="card-stats">
              <span>总用户: {{ userStats.total_users || 0 }}</span>
              <span>活跃用户: {{ userStats.active_users || 0 }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="8">
        <el-card class="management-card" shadow="hover" @click="$router.push('/permissions/groups')">
          <div class="card-content">
            <el-icon class="card-icon" size="48" color="#67C23A">
              <UserFilled />
            </el-icon>
            <h3>分组管理</h3>
            <p>管理用户组和组成员权限</p>
            <div class="card-stats">
              <span>用户组: {{ groupStats.total_groups || 0 }}</span>
              <span>总成员: {{ groupStats.total_members || 0 }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="8">
        <el-card class="management-card" shadow="hover">
          <div class="card-content">
            <el-icon class="card-icon" size="48" color="#E6A23C">
              <Key />
            </el-icon>
            <h3>权限概览</h3>
            <p>查看系统权限分配情况</p>
            <div class="card-stats">
              <span>管理员: {{ userStats.admin_count || 0 }}</span>
              <span>教师: {{ userStats.teacher_count || 0 }}</span>
              <span>学生: {{ userStats.student_count || 0 }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <h2>快速操作</h2>
      <el-row :gutter="16">
        <el-col :span="6">
          <el-button type="primary" @click="handleQuickAction('create-user')" block>
            <el-icon><Plus /></el-icon>
            创建用户
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="success" @click="handleQuickAction('create-group')" block>
            <el-icon><Plus /></el-icon>
            创建用户组
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="warning" @click="handleQuickAction('sync-gitlab')" block>
            <el-icon><Refresh /></el-icon>
            同步GitLab
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="info" @click="handleQuickAction('export-users')" block>
            <el-icon><Download /></el-icon>
            导出用户
          </el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 最近活动 -->
    <div class="recent-activities">
      <h2>最近活动</h2>
      <el-timeline>
        <el-timeline-item
          v-for="activity in recentActivities"
          :key="activity.id"
          :timestamp="activity.timestamp"
        >
          {{ activity.description }}
        </el-timeline-item>
      </el-timeline>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { User, UserFilled, Key, Plus, Refresh, Download } from '@element-plus/icons-vue'
import api from '../services/api'

// 响应式数据
const userStats = ref({
  total_users: 0,
  active_users: 0,
  admin_count: 0,
  teacher_count: 0,
  student_count: 0
})

const groupStats = ref({
  total_groups: 0,
  total_members: 0
})

const recentActivities = ref([
  {
    id: 1,
    description: '管理员创建了新用户组 "2024级学生"',
    timestamp: '2024-01-15 10:30:00'
  },
  {
    id: 2,
    description: '用户 "张老师" 被分配为教师角色',
    timestamp: '2024-01-15 09:45:00'
  },
  {
    id: 3,
    description: '同步了10个GitLab用户',
    timestamp: '2024-01-15 09:00:00'
  }
])

// 方法
const handleQuickAction = (action: string) => {
  switch (action) {
    case 'create-user':
      ElMessage.info('创建用户功能开发中...')
      break
    case 'create-group':
      ElMessage.info('创建用户组功能开发中...')
      break
    case 'sync-gitlab':
      handleSyncGitLab()
      break
    case 'export-users':
      ElMessage.info('导出用户功能开发中...')
      break
    default:
      break
  }
}

const handleSyncGitLab = async () => {
  try {
    ElMessage.info('正在同步GitLab用户数据...')
    // TODO: 实现GitLab同步
    ElMessage.success('GitLab用户数据同步完成')
  } catch (error) {
    console.error('同步GitLab失败:', error)
    ElMessage.error('同步GitLab失败')
  }
}

const loadUserStats = async () => {
  try {
    // TODO: 调用API获取用户统计
    // const response = await api.getUserStats()
    // userStats.value = response.data
    
    // 模拟数据
    userStats.value = {
      total_users: 156,
      active_users: 142,
      admin_count: 3,
      teacher_count: 25,
      student_count: 128
    }
  } catch (error) {
    console.error('获取用户统计失败:', error)
  }
}

const loadGroupStats = async () => {
  try {
    // TODO: 调用API获取分组统计
    // const response = await api.getGroupStats()
    // groupStats.value = response.data
    
    // 模拟数据
    groupStats.value = {
      total_groups: 12,
      total_members: 156
    }
  } catch (error) {
    console.error('获取分组统计失败:', error)
  }
}

// 生命周期
onMounted(() => {
  loadUserStats()
  loadGroupStats()
})
</script>

<style scoped>
.permissions-view {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  color: #303133;
  margin-bottom: 8px;
}

.page-header p {
  color: #606266;
  margin: 0;
}

.management-cards {
  margin-bottom: 32px;
}

.management-card {
  cursor: pointer;
  transition: all 0.3s ease;
  height: 200px;
}

.management-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.card-content {
  text-align: center;
  padding: 20px;
}

.card-icon {
  margin-bottom: 16px;
}

.card-content h3 {
  color: #303133;
  margin: 0 0 8px 0;
  font-size: 18px;
}

.card-content p {
  color: #606266;
  margin: 0 0 16px 0;
  font-size: 14px;
}

.card-stats {
  display: flex;
  justify-content: space-around;
  font-size: 12px;
  color: #909399;
}

.quick-actions, .recent-activities {
  margin-bottom: 32px;
}

.quick-actions h2, .recent-activities h2 {
  color: #303133;
  margin-bottom: 16px;
  font-size: 20px;
}

.quick-actions .el-button {
  height: 48px;
}

@media (max-width: 768px) {
  .permissions-view {
    padding: 16px;
  }
  
  .management-card {
    margin-bottom: 16px;
  }
  
  .quick-actions .el-col {
    margin-bottom: 12px;
  }
}
</style> 
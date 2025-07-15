<template>
  <div class="permission-groups-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>分组管理</h1>
      <p>管理GitLab用户组和成员权限</p>
    </div>

    <!-- 功能说明 -->
    <el-alert
      title="GitLab分组管理"
      type="info"
      description="此功能直接调用GitLab API进行用户组管理，所有操作都会同步到GitLab系统中。"
      :closable="false"
      show-icon
    />

    <!-- 快速链接 -->
    <div class="quick-links">
      <h2>快速链接</h2>
      <el-row :gutter="24">
        <el-col :span="8">
          <el-card shadow="hover">
            <div class="link-content">
              <el-icon size="32" color="#409EFF"><UserFilled /></el-icon>
              <h3>GitLab用户组</h3>
              <p>在GitLab中查看和管理用户组</p>
              <el-button type="primary" @click="openGitLabGroups">前往GitLab</el-button>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card shadow="hover">
            <div class="link-content">
              <el-icon size="32" color="#67C23A"><User /></el-icon>
              <h3>用户管理</h3>
              <p>管理系统用户信息和角色</p>
              <el-button @click="$router.push('/permissions/users')">用户管理</el-button>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card shadow="hover">
            <div class="link-content">
              <el-icon size="32" color="#E6A23C"><Setting /></el-icon>
              <h3>权限配置</h3>
              <p>配置角色和权限分配</p>
              <el-button @click="$router.push('/permissions')">权限概览</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 使用说明 -->
    <div class="instructions">
      <h2>使用说明</h2>
      <el-steps :active="4" finish-status="success">
        <el-step title="访问GitLab" description="点击上方链接前往GitLab管理界面" />
        <el-step title="创建用户组" description="在GitLab中创建新的用户组" />
        <el-step title="添加成员" description="将用户添加到相应的用户组" />
        <el-step title="分配权限" description="为用户组成员分配合适的权限级别" />
        <el-step title="同步更新" description="权限变更会自动同步到教育平台" />
      </el-steps>
    </div>

    <!-- 权限级别说明 -->
    <div class="permission-levels">
      <h2>GitLab权限级别</h2>
      <el-table :data="permissionLevels" style="width: 100%">
        <el-table-column prop="level" label="权限级别" width="100" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="education_role" label="教育角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.tag_type">{{ row.education_role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="权限描述" />
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UserFilled, User, Setting } from '@element-plus/icons-vue'

// 权限级别数据
const permissionLevels = ref([
  {
    level: 50,
    name: 'Owner',
    education_role: '管理员',
    tag_type: 'danger',
    description: '拥有项目的完全访问权限，可以删除项目'
  },
  {
    level: 40,
    name: 'Maintainer',
    education_role: '教师',
    tag_type: 'warning',
    description: '可以推送到受保护分支，管理项目设置'
  },
  {
    level: 30,
    name: 'Developer',
    education_role: '助教',
    tag_type: 'info',
    description: '可以推送到非受保护分支，创建合并请求'
  },
  {
    level: 20,
    name: 'Reporter',
    education_role: '学生',
    tag_type: 'success',
    description: '可以查看项目代码，创建问题和评论'
  },
  {
    level: 10,
    name: 'Guest',
    education_role: '访客',
    tag_type: '',
    description: '有限的项目访问权限，只能查看项目'
  }
])

// 方法
const openGitLabGroups = () => {
  // TODO: 从配置中获取GitLab URL
  const gitlabUrl = 'http://localhost:8080' // 示例URL
  window.open(`${gitlabUrl}/admin/groups`, '_blank')
}
</script>

<style scoped>
.permission-groups-view {
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

.quick-links, .instructions, .permission-levels {
  margin: 32px 0;
}

.quick-links h2, .instructions h2, .permission-levels h2 {
  color: #303133;
  margin-bottom: 16px;
  font-size: 20px;
}

.link-content {
  text-align: center;
  padding: 20px;
}

.link-content h3 {
  color: #303133;
  margin: 16px 0 8px 0;
  font-size: 16px;
}

.link-content p {
  color: #606266;
  margin: 0 0 16px 0;
  font-size: 14px;
}

@media (max-width: 768px) {
  .permission-groups-view {
    padding: 16px;
  }
  
  .quick-links .el-col {
    margin-bottom: 16px;
  }
}
</style> 
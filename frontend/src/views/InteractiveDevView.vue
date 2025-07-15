<template>
  <div class="interactive-dev">
    <!-- 顶部工具栏 -->
    <div class="dev-toolbar">
      <div class="toolbar-left">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>
            <router-link to="/projects">课题</router-link>
          </el-breadcrumb-item>
          <el-breadcrumb-item>
            <router-link :to="`/projects/${projectId}`">
              {{ projectInfo?.name }}
            </router-link>
          </el-breadcrumb-item>
          <el-breadcrumb-item>互动开发</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      
      <div class="toolbar-right">
        <el-button-group>
          <el-button 
            size="small" 
            @click="createStudentBranch"
            v-if="needCreateBranch"
            type="primary"
          >
            <el-icon><Plus /></el-icon>
            创建我的分支
          </el-button>
          
          <el-button 
            size="small" 
            @click="submitAssignment"
            v-if="hasPersonalBranch"
            :disabled="!hasChanges"
          >
            <el-icon><Upload /></el-icon>
            提交作业
          </el-button>
          
          <el-button 
            size="small" 
            @click="viewGitLabRepo"
          >
            <el-icon><Link /></el-icon>
            GitLab仓库
          </el-button>
        </el-button-group>
        
        <el-dropdown @command="handleToolbarCommand">
          <el-button size="small">
            <el-icon><MoreFilled /></el-icon>
            更多
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="clone">
                <el-icon><Download /></el-icon>
                克隆到本地
              </el-dropdown-item>
              <el-dropdown-item command="history">
                <el-icon><Clock /></el-icon>
                提交历史
              </el-dropdown-item>
              <el-dropdown-item command="settings">
                <el-icon><Setting /></el-icon>
                项目设置
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <div class="dev-content">
      <!-- 左侧文件树 -->
      <div class="file-panel" :style="{ width: filePanelWidth + 'px' }">
        <FileTree
          ref="fileTreeRef"
          :project-id="projectId"
          :project-name="projectInfo?.name || ''"
          :branch="currentBranch"
          :readonly="isReadonly"
          @file-selected="handleFileSelected"
          @branch-changed="handleBranchChanged"
          @file-created="handleFileCreated"
        />
      </div>

      <!-- 分割线 -->
      <div 
        class="panel-divider"
        @mousedown="startResize"
      ></div>

      <!-- 右侧代码编辑器 -->
      <div class="editor-panel">
        <div class="editor-tabs" v-if="openFiles.length > 0">
          <div 
            v-for="file in openFiles"
            :key="file.path"
            class="editor-tab"
            :class="{ active: file.path === currentFile }"
            @click="switchFile(file.path)"
          >
            <span class="tab-label">{{ file.name }}</span>
            <el-icon 
              class="tab-close" 
              @click.stop="closeFile(file.path)"
            >
              <Close />
            </el-icon>
          </div>
        </div>

        <div class="editor-content">
          <div class="editor-welcome" v-if="!currentFile">
            <div class="welcome-content">
              <el-icon class="welcome-icon"><Document /></el-icon>
              <h3>欢迎使用互动开发环境</h3>
              <p>请从左侧文件树中选择文件开始编辑</p>
              
              <div class="welcome-actions">
                <el-button 
                  v-if="needCreateBranch"
                  type="primary"
                  @click="createStudentBranch"
                >
                  创建我的分支
                </el-button>
                
                <el-button 
                  @click="showQuickStart"
                  text
                >
                  快速开始指南
                </el-button>
              </div>
            </div>
          </div>

          <CodeEditor
            v-else
            ref="codeEditorRef"
            :project-id="projectId"
            :project-name="projectInfo?.name || ''"
            :file-path="currentFile"
            :branch="currentBranch"
            :readonly="isReadonly"
            @content-changed="handleContentChanged"
            @saved="handleFileSaved"
          />
        </div>
      </div>
    </div>

    <!-- 底部状态栏 -->
    <div class="dev-statusbar">
      <div class="statusbar-left">
        <span class="status-item">
          <el-icon><User /></el-icon>
          {{ userInfo?.name }}
        </span>
        
                 <span class="status-item">
           <el-icon><Document /></el-icon>
           {{ currentBranch }}
         </span>
        
        <span class="status-item" v-if="currentFile">
          <el-icon><Document /></el-icon>
          {{ currentFile }}
        </span>
      </div>
      
      <div class="statusbar-right">
        <span class="status-item" v-if="hasChanges">
          <el-icon><EditPen /></el-icon>
          有未保存的更改
        </span>
        
        <span class="status-item" v-if="activeEditors > 0">
          <el-icon><View /></el-icon>
          {{ activeEditors }} 人在线编辑
        </span>
        
        <span class="status-item">
          <el-icon><Clock /></el-icon>
          {{ formatTime(new Date()) }}
        </span>
      </div>
    </div>

    <!-- 提交作业对话框 -->
    <el-dialog 
      v-model="submitDialog" 
      title="提交作业" 
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="submitForm" label-width="100px">
        <el-form-item label="作业说明">
          <el-input 
            v-model="submitForm.description" 
            placeholder="请描述您的作业内容..."
            type="textarea"
            :rows="4"
          />
        </el-form-item>
        
        <el-form-item label="提交文件">
          <div class="file-list">
            <div 
              v-for="file in changedFiles"
              :key="file.path"
              class="file-item"
            >
              <el-icon><Document /></el-icon>
              <span>{{ file.path }}</span>
              <el-tag size="small" type="success">已修改</el-tag>
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="submitDialog = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="confirmSubmitAssignment"
          :loading="submitting"
        >
          确认提交
        </el-button>
      </template>
    </el-dialog>

    <!-- 快速开始对话框 -->
    <el-dialog 
      v-model="quickStartDialog" 
      title="快速开始指南" 
      width="700px"
    >
      <div class="quick-start-content">
        <el-steps :active="quickStartStep" direction="vertical">
          <el-step title="创建个人分支" description="为您的开发工作创建独立的分支">
            <template #icon>
              <el-icon><Plus /></el-icon>
            </template>
          </el-step>
          
          <el-step title="编辑代码" description="在线编辑或本地克隆进行开发">
            <template #icon>
              <el-icon><Edit /></el-icon>
            </template>
          </el-step>
          
          <el-step title="提交作业" description="完成开发后提交您的作业">
            <template #icon>
              <el-icon><Upload /></el-icon>
            </template>
          </el-step>
        </el-steps>
      </div>
      
      <template #footer>
        <el-button @click="quickStartDialog = false">关闭</el-button>
        <el-button 
          v-if="needCreateBranch"
          type="primary" 
          @click="createStudentBranch"
        >
          开始使用
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Upload,
  Link,
  MoreFilled,
  Download,
  Clock,
  Setting,
  Document,
  Close,
  User,
  EditPen,
  View,
  Edit
} from '@element-plus/icons-vue'

import FileTree from '../components/FileTree.vue'
import CodeEditor from '../components/CodeEditor.vue'

// 路由
const route = useRoute()
const router = useRouter()

// Props
const projectId = computed(() => Number(route.params.id))

// Refs
const fileTreeRef = ref()
const codeEditorRef = ref()
const submitDialog = ref(false)
const quickStartDialog = ref(false)
const submitting = ref(false)
const loading = ref(false)

// 数据
const projectInfo = ref<any>(null)
const userInfo = ref<any>(null)
const currentBranch = ref('main')
const currentFile = ref('')
const openFiles = ref<any[]>([])
const hasChanges = ref(false)
const activeEditors = ref(0)
const filePanelWidth = ref(300)
const hasPersonalBranch = ref(false)
const quickStartStep = ref(0)

// 表单数据
const submitForm = ref({
  description: ''
})

// 计算属性
const needCreateBranch = computed(() => {
  return userInfo.value?.role === 3 && !hasPersonalBranch.value // 学生且没有个人分支
})

const isReadonly = computed(() => {
  // 如果是主分支，只有创建者和管理员可以编辑
  if (currentBranch.value === projectInfo.value?.default_branch) {
    return userInfo.value?.role !== 1 && projectInfo.value?.teacher_id !== userInfo.value?.id
  }
  return false
})

const changedFiles = computed(() => {
  return openFiles.value.filter(file => file.hasChanges)
})

// 监听路由变化
watch(() => route.params.id, (newId) => {
  if (newId) {
    loadProjectInfo()
  }
})

// 加载项目信息
const loadProjectInfo = async () => {
  loading.value = true
  try {
    const response = await fetch(`/api/projects/${projectId.value}`)
    const data = await response.json()
    
    if (data.success) {
      projectInfo.value = data.project
      currentBranch.value = data.project.default_branch || 'main'
      
      // 检查是否有个人分支
      await checkPersonalBranch()
    } else {
      ElMessage.error(data.error || '加载项目信息失败')
    }
  } catch (error) {
    console.error('Load project error:', error)
    ElMessage.error('加载项目信息失败')
  } finally {
    loading.value = false
  }
}

// 检查个人分支
const checkPersonalBranch = async () => {
  try {
    const response = await fetch(`/api/projects/${projectId.value}/members/current`)
    const data = await response.json()
    
    if (data.success && data.member?.personal_branch) {
      hasPersonalBranch.value = true
      currentBranch.value = data.member.personal_branch
    }
  } catch (error) {
    console.error('Check personal branch error:', error)
  }
}

// 创建学生分支
const createStudentBranch = async () => {
  try {
    const response = await fetch(`/api/projects/${projectId.value}/code/student-branch`, {
      method: 'POST'
    })
    
    const data = await response.json()
    
    if (data.success) {
      ElMessage.success('个人分支创建成功')
      hasPersonalBranch.value = true
      
      // 刷新项目信息
      await checkPersonalBranch()
      
      // 刷新文件树
      if (fileTreeRef.value) {
        fileTreeRef.value.refresh()
      }
    } else {
      ElMessage.error(data.error || '创建分支失败')
    }
  } catch (error) {
    console.error('Create branch error:', error)
    ElMessage.error('创建分支失败')
  }
}

// 处理文件选择
const handleFileSelected = (filePath: string, fileType: string) => {
  if (fileType === 'file') {
    openFile(filePath)
  }
}

// 打开文件
const openFile = (filePath: string) => {
  // 检查文件是否已经打开
  const existingFile = openFiles.value.find(f => f.path === filePath)
  if (existingFile) {
    currentFile.value = filePath
    return
  }
  
  // 添加到打开文件列表
  const fileName = filePath.split('/').pop() || filePath
  openFiles.value.push({
    path: filePath,
    name: fileName,
    hasChanges: false
  })
  
  currentFile.value = filePath
}

// 切换文件
const switchFile = (filePath: string) => {
  currentFile.value = filePath
}

// 关闭文件
const closeFile = (filePath: string) => {
  const index = openFiles.value.findIndex(f => f.path === filePath)
  if (index >= 0) {
    openFiles.value.splice(index, 1)
    
    // 如果关闭的是当前文件，切换到其他文件
    if (currentFile.value === filePath) {
      if (openFiles.value.length > 0) {
        currentFile.value = openFiles.value[0].path
      } else {
        currentFile.value = ''
      }
    }
  }
}

// 处理分支变化
const handleBranchChanged = (branch: string) => {
  currentBranch.value = branch
  
  // 关闭所有打开的文件
  openFiles.value = []
  currentFile.value = ''
}

// 处理文件创建
const handleFileCreated = (filePath: string) => {
  // 自动打开新创建的文件
  openFile(filePath)
}

// 处理内容变化
const handleContentChanged = (content: string) => {
  hasChanges.value = true
  
  // 更新文件状态
  const file = openFiles.value.find(f => f.path === currentFile.value)
  if (file) {
    file.hasChanges = true
  }
}

// 处理文件保存
const handleFileSaved = (filePath: string, content: string) => {
  // 更新文件状态
  const file = openFiles.value.find(f => f.path === filePath)
  if (file) {
    file.hasChanges = false
  }
  
  // 检查是否还有未保存的文件
  hasChanges.value = openFiles.value.some(f => f.hasChanges)
}

// 提交作业
const submitAssignment = () => {
  if (!hasPersonalBranch.value) {
    ElMessage.warning('请先创建个人分支')
    return
  }
  
  submitDialog.value = true
}

// 确认提交作业
const confirmSubmitAssignment = async () => {
  submitting.value = true
  try {
    const response = await fetch(`/api/projects/${projectId.value}/assignments/submit`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        branch: currentBranch.value,
        description: submitForm.value.description,
        files: changedFiles.value.map(f => f.path)
      })
    })
    
    const data = await response.json()
    
    if (data.success) {
      submitDialog.value = false
      ElMessage.success('作业提交成功')
      
      // 重置表单
      submitForm.value.description = ''
    } else {
      ElMessage.error(data.error || '提交失败')
    }
  } catch (error) {
    console.error('Submit assignment error:', error)
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

// 查看GitLab仓库
const viewGitLabRepo = () => {
  if (projectInfo.value?.gitlab_url) {
    window.open(projectInfo.value.gitlab_url, '_blank')
  }
}

// 处理工具栏命令
const handleToolbarCommand = (command: string) => {
  switch (command) {
    case 'clone':
      showCloneInfo()
      break
    case 'history':
      showCommitHistory()
      break
    case 'settings':
      router.push(`/projects/${projectId.value}/settings`)
      break
  }
}

// 显示克隆信息
const showCloneInfo = () => {
  const cloneUrl = projectInfo.value?.repository_url
  if (cloneUrl) {
    ElMessageBox.alert(
      `git clone ${cloneUrl}`,
      '克隆仓库',
      {
        confirmButtonText: '复制',
        callback: () => {
          navigator.clipboard.writeText(`git clone ${cloneUrl}`)
          ElMessage.success('克隆命令已复制到剪贴板')
        }
      }
    )
  }
}

// 显示提交历史
const showCommitHistory = () => {
  ElMessage.info('提交历史功能待实现')
}

// 显示快速开始
const showQuickStart = () => {
  quickStartDialog.value = true
}

// 开始调整面板大小
const startResize = (e: MouseEvent) => {
  const startX = e.clientX
  const startWidth = filePanelWidth.value
  
  const handleMouseMove = (e: MouseEvent) => {
    const newWidth = startWidth + (e.clientX - startX)
    filePanelWidth.value = Math.max(200, Math.min(600, newWidth))
  }
  
  const handleMouseUp = () => {
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }
  
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

// 格式化时间
const formatTime = (date: Date) => {
  return date.toLocaleTimeString()
}

// 加载用户信息
const loadUserInfo = async () => {
  try {
    const response = await fetch('/api/users/current')
    const data = await response.json()
    
    if (data.success) {
      userInfo.value = data.user
    }
  } catch (error) {
    console.error('Load user info error:', error)
  }
}

// 生命周期
onMounted(() => {
  loadUserInfo()
  loadProjectInfo()
  
  // 定时更新活跃编辑器数量
  setInterval(async () => {
    try {
      const response = await fetch(`/api/projects/${projectId.value}/code/active-sessions`)
      const data = await response.json()
      
      if (data.success) {
        activeEditors.value = data.sessions?.length || 0
      }
    } catch (error) {
      // 忽略错误
    }
  }, 30000) // 30秒更新一次
})
</script>

<style scoped>
.interactive-dev {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #cccccc;
}

.dev-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #2d2d30;
  border-bottom: 1px solid #3e3e42;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dev-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.file-panel {
  background: #252526;
  border-right: 1px solid #3e3e42;
  overflow: hidden;
}

.panel-divider {
  width: 4px;
  background: #3e3e42;
  cursor: col-resize;
}

.panel-divider:hover {
  background: #409eff;
}

.editor-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.editor-tabs {
  display: flex;
  background: #2d2d30;
  border-bottom: 1px solid #3e3e42;
  overflow-x: auto;
}

.editor-tab {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #2d2d30;
  border-right: 1px solid #3e3e42;
  cursor: pointer;
  white-space: nowrap;
}

.editor-tab.active {
  background: #1e1e1e;
}

.editor-tab:hover {
  background: #37373d;
}

.tab-label {
  margin-right: 8px;
  font-size: 13px;
}

.tab-close {
  font-size: 12px;
  opacity: 0.7;
}

.tab-close:hover {
  opacity: 1;
  color: #f56c6c;
}

.editor-content {
  flex: 1;
  overflow: hidden;
}

.editor-welcome {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1e1e1e;
}

.welcome-content {
  text-align: center;
  max-width: 400px;
}

.welcome-icon {
  font-size: 48px;
  color: #6c6c6c;
  margin-bottom: 16px;
}

.welcome-content h3 {
  margin: 0 0 8px 0;
  color: #cccccc;
}

.welcome-content p {
  margin: 0 0 24px 0;
  color: #909399;
}

.welcome-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.dev-statusbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 16px;
  background: #007acc;
  color: white;
  font-size: 12px;
}

.statusbar-left,
.statusbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.file-list {
  max-height: 200px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.quick-start-content {
  padding: 20px;
}

/* Element Plus 样式覆盖 */
:deep(.el-breadcrumb__inner) {
  color: #cccccc !important;
}

:deep(.el-breadcrumb__separator) {
  color: #909399 !important;
}

:deep(.el-button--small) {
  background: #3c3c3c;
  border-color: #3c3c3c;
  color: #cccccc;
}

:deep(.el-button--small:hover) {
  background: #404040;
  border-color: #404040;
}

:deep(.el-button--small.is-disabled) {
  background: #2d2d30;
  border-color: #2d2d30;
  color: #6c6c6c;
}

:deep(.el-button--primary) {
  background: #409eff;
  border-color: #409eff;
}

:deep(.el-steps--vertical) {
  padding: 0 20px;
}
</style> 
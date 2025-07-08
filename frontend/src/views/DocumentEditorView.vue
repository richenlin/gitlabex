<template>
  <div class="editor-container">
    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-container">
      <el-card class="loading-card">
        <div class="loading-content">
          <el-icon class="loading-icon" :size="48"><Loading /></el-icon>
          <h3 class="loading-title">正在加载文档编辑器...</h3>
          <p class="loading-text">请稍候，正在为您准备文档编辑环境</p>
        </div>
      </el-card>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-container">
      <el-card class="error-card">
        <div class="error-content">
          <el-icon class="error-icon" :size="48" color="#f56c6c"><WarningFilled /></el-icon>
          <h3 class="error-title">加载失败</h3>
          <p class="error-text">{{ error }}</p>
          <div class="error-actions">
            <el-button @click="goBack">返回</el-button>
            <el-button type="primary" @click="reload">重新加载</el-button>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 编辑器工具栏 -->
    <div v-else class="editor-toolbar">
      <div class="toolbar-left">
        <el-button 
          :icon="ArrowLeft" 
          @click="goBack"
          type="text"
          class="back-button"
        >
          返回文档列表
        </el-button>
        
        <el-divider direction="vertical" />
        
        <div class="document-info">
          <el-icon class="document-icon"><Document /></el-icon>
          <span class="document-title">{{ documentTitle }}</span>
          <el-tag v-if="isReadOnly" type="info" size="small">只读模式</el-tag>
        </div>
      </div>
      
      <div class="toolbar-right">
        <el-button 
          :icon="Refresh" 
          @click="reload"
          type="text"
        >
          刷新
        </el-button>
        
        <el-button 
          :icon="FullScreen" 
          @click="toggleFullscreen"
          type="text"
        >
          {{ isFullscreen ? '退出全屏' : '全屏' }}
        </el-button>
        
        <el-dropdown @command="handleCommand">
          <el-button type="text">
            更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="save" :icon="DocumentChecked">保存文档</el-dropdown-item>
              <el-dropdown-item command="download" :icon="Download">下载文档</el-dropdown-item>
              <el-dropdown-item command="share" :icon="Share">分享文档</el-dropdown-item>
              <el-dropdown-item command="info" :icon="InfoFilled">文档信息</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- OnlyOffice编辑器容器 -->
    <div v-show="!isLoading && !error" class="editor-frame-container">
      <iframe
        ref="editorFrame"
        :src="editorUrl"
        class="editor-frame"
        frameborder="0"
        allowfullscreen
        @load="onEditorLoad"
      ></iframe>
    </div>

    <!-- 文档信息对话框 -->
    <el-dialog
      v-model="infoDialogVisible"
      title="文档信息"
      width="500px"
    >
      <div class="document-info-content" v-if="documentConfig">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="文档ID">
            {{ documentId }}
          </el-descriptions-item>
          <el-descriptions-item label="文档标题">
            {{ documentConfig.document.title }}
          </el-descriptions-item>
          <el-descriptions-item label="文件类型">
            <el-tag :type="getFileTypeTagType(documentConfig.document.fileType)">
              {{ documentConfig.document.fileType?.toUpperCase() }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="编辑模式">
            {{ documentConfig.editor.mode === 'edit' ? '编辑模式' : '只读模式' }}
          </el-descriptions-item>
          <el-descriptions-item label="语言">
            {{ documentConfig.editor.lang === 'zh-CN' ? '中文' : documentConfig.editor.lang }}
          </el-descriptions-item>
          <el-descriptions-item label="权限">
            <div class="permissions">
              <el-tag v-if="documentConfig.document.permissions.edit" type="success" size="small">编辑</el-tag>
              <el-tag v-if="documentConfig.document.permissions.comment" type="primary" size="small">评论</el-tag>
              <el-tag v-if="documentConfig.document.permissions.download" type="info" size="small">下载</el-tag>
              <el-tag v-if="documentConfig.document.permissions.print" type="warning" size="small">打印</el-tag>
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft,
  Document,
  Refresh,
  FullScreen,
  Loading,
  WarningFilled,
  DocumentChecked,
  Download,
  Share,
  InfoFilled,
  ArrowDown
} from '@element-plus/icons-vue'
import { ApiService } from '../services/api'
import type { DocumentConfig } from '../services/api'

const route = useRoute()
const router = useRouter()

// 响应式数据
const isLoading = ref(true)
const error = ref('')
const documentConfig = ref<DocumentConfig | null>(null)
const editorFrame = ref<HTMLIFrameElement>()
const isFullscreen = ref(false)
const infoDialogVisible = ref(false)

// 计算属性
const documentId = computed(() => {
  return Number(route.params.id)
})

const documentTitle = computed(() => {
  return documentConfig.value?.document.title || `文档_${documentId.value}`
})

const isReadOnly = computed(() => {
  return route.query.mode === 'view' || documentConfig.value?.editor.mode === 'view'
})

const editorUrl = computed(() => {
  if (!documentId.value) return ''
  return ApiService.getDocumentEditorUrl(documentId.value)
})

// 生命周期
onMounted(() => {
  loadDocument()
  
  // 监听全屏事件
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('mozfullscreenchange', handleFullscreenChange)
  document.addEventListener('MSFullscreenChange', handleFullscreenChange)
})

onUnmounted(() => {
  // 清理事件监听器
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.removeEventListener('mozfullscreenchange', handleFullscreenChange)
  document.removeEventListener('MSFullscreenChange', handleFullscreenChange)
})

// 方法
const loadDocument = async () => {
  if (!documentId.value) {
    error.value = '无效的文档ID'
    isLoading.value = false
    return
  }

  try {
    isLoading.value = true
    error.value = ''
    
    // 获取文档配置
    documentConfig.value = await ApiService.getDocumentConfig(documentId.value)
    
    // 延迟一下，确保iframe能正确加载
    await new Promise(resolve => setTimeout(resolve, 500))
    
    ElMessage.success('文档加载成功')
  } catch (err: any) {
    console.error('加载文档失败:', err)
    error.value = err.response?.data?.message || '加载文档失败，请稍后重试'
    ElMessage.error(error.value)
  } finally {
    isLoading.value = false
  }
}

const onEditorLoad = () => {
  console.log('编辑器加载完成')
  ElMessage.success('编辑器加载完成')
}

const goBack = () => {
  router.push('/documents')
}

const reload = () => {
  loadDocument()
}

const toggleFullscreen = () => {
  const element = document.documentElement
  
  if (!isFullscreen.value) {
    // 进入全屏
    if (element.requestFullscreen) {
      element.requestFullscreen()
    } else if ((element as any).webkitRequestFullscreen) {
      (element as any).webkitRequestFullscreen()
    } else if ((element as any).mozRequestFullScreen) {
      (element as any).mozRequestFullScreen()
    } else if ((element as any).msRequestFullscreen) {
      (element as any).msRequestFullscreen()
    }
  } else {
    // 退出全屏
    if (document.exitFullscreen) {
      document.exitFullscreen()
    } else if ((document as any).webkitExitFullscreen) {
      (document as any).webkitExitFullscreen()
    } else if ((document as any).mozCancelFullScreen) {
      (document as any).mozCancelFullScreen()
    } else if ((document as any).msExitFullscreen) {
      (document as any).msExitFullscreen()
    }
  }
}

const handleFullscreenChange = () => {
  isFullscreen.value = !!(
    document.fullscreenElement ||
    (document as any).webkitFullscreenElement ||
    (document as any).mozFullScreenElement ||
    (document as any).msFullscreenElement
  )
}

const handleCommand = (command: string) => {
  switch (command) {
    case 'save':
      handleSave()
      break
    case 'download':
      handleDownload()
      break
    case 'share':
      handleShare()
      break
    case 'info':
      infoDialogVisible.value = true
      break
  }
}

const handleSave = () => {
  // OnlyOffice会自动保存，这里只是提示
  ElMessage.success('文档已自动保存')
}

const handleDownload = () => {
  if (documentConfig.value?.document.permissions.download) {
    const url = ApiService.getDocumentContentUrl(documentId.value)
    window.open(url, '_blank')
  } else {
    ElMessage.warning('您没有下载权限')
  }
}

const handleShare = async () => {
  try {
    const shareUrl = `${window.location.origin}/documents/editor/${documentId.value}?mode=view`
    await navigator.clipboard.writeText(shareUrl)
    ElMessage.success('分享链接已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败，请手动复制链接')
  }
}

const getFileTypeTagType = (fileType: string) => {
  switch (fileType?.toLowerCase()) {
    case 'docx':
    case 'doc':
      return 'primary'
    case 'xlsx':
    case 'xls':
      return 'success'
    case 'pptx':
    case 'ppt':
      return 'warning'
    default:
      return 'info'
  }
}
</script>

<style scoped>
.editor-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
}

.loading-container,
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.loading-card,
.error-card {
  max-width: 400px;
  text-align: center;
}

.loading-content,
.error-content {
  padding: 40px 20px;
}

.loading-icon {
  color: #409eff;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-title,
.error-title {
  margin: 20px 0 10px 0;
  color: #2c3e50;
}

.loading-text,
.error-text {
  margin: 0 0 20px 0;
  color: #7f8c8d;
}

.error-icon {
  color: #f56c6c;
}

.error-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-button {
  color: #409eff;
  font-weight: 500;
}

.document-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.document-icon {
  color: #409eff;
  font-size: 16px;
}

.document-title {
  font-weight: 600;
  color: #2c3e50;
}

.editor-frame-container {
  flex: 1;
  position: relative;
  background: white;
  margin: 0;
  padding: 0;
}

.editor-frame {
  width: 100%;
  height: 100%;
  border: none;
}

.document-info-content {
  padding: 20px 0;
}

.permissions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 全屏样式 */
.editor-container:fullscreen {
  background: white;
}

.editor-container:fullscreen .editor-toolbar {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .editor-toolbar {
    padding: 8px 12px;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .toolbar-left,
  .toolbar-right {
    gap: 8px;
  }
  
  .document-title {
    display: none;
  }
  
  .loading-card,
  .error-card {
    margin: 20px;
  }
}
</style> 
<template>
  <div class="code-editor">
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>{{ projectName }}</el-breadcrumb-item>
          <el-breadcrumb-item v-if="currentBranch">
            <el-tag size="small" type="success">{{ currentBranch }}</el-tag>
          </el-breadcrumb-item>
          <el-breadcrumb-item v-if="currentFilePath">{{ currentFilePath }}</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      
      <div class="toolbar-right">
        <el-button-group>
          <el-button 
            size="small" 
            :disabled="!hasChanges" 
            @click="saveFile"
            :loading="saving"
          >
            <el-icon><Document /></el-icon>
            保存 (Ctrl+S)
          </el-button>
          
          <el-button 
            size="small" 
            @click="toggleFullscreen"
          >
            <el-icon><FullScreen /></el-icon>
            全屏
          </el-button>
        </el-button-group>
        
        <el-dropdown @command="handleSettingsCommand">
          <el-button size="small">
            <el-icon><Setting /></el-icon>
            设置
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="theme">切换主题</el-dropdown-item>
              <el-dropdown-item command="fontSize">字体大小</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div class="editor-container" ref="editorContainer">
      <div class="editor-loading" v-if="loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      
      <textarea
        ref="textEditor"
        v-model="currentContent"
        class="text-editor"
        :class="{ 'dark-theme': editorSettings.theme === 'dark' }"
        :style="{ 
          fontSize: editorSettings.fontSize + 'px',
          height: editorHeight + 'px'
        }"
        :readonly="props.readonly"
        @input="handleContentChange"
        @keydown="handleKeyDown"
        spellcheck="false"
        wrap="off"
        placeholder="请选择文件开始编辑..."
      />
    </div>

    <!-- 文件保存对话框 -->
    <el-dialog 
      v-model="saveDialog" 
      title="保存文件" 
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="saveForm" label-width="100px">
        <el-form-item label="提交消息">
          <el-input 
            v-model="saveForm.message" 
            placeholder="请输入提交消息..."
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="saveDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmSave" :loading="saving">
          确认保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑器设置对话框 -->
    <el-dialog 
      v-model="settingsDialog" 
      title="编辑器设置" 
      width="600px"
    >
      <el-form :model="editorSettings" label-width="120px">
        <el-form-item label="主题">
          <el-select v-model="editorSettings.theme">
            <el-option label="浅色主题" value="light" />
            <el-option label="深色主题" value="dark" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="字体大小">
          <el-slider 
            v-model="editorSettings.fontSize" 
            :min="10" 
            :max="24" 
            :step="1"
          />
          <span class="setting-value">{{ editorSettings.fontSize }}px</span>
        </el-form-item>
        
        <el-form-item label="自动保存">
          <el-switch 
            v-model="editorSettings.autoSave" 
            @change="toggleAutoSave"
          />
          <span class="setting-desc">每30秒自动保存</span>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="settingsDialog = false">关闭</el-button>
        <el-button type="primary" @click="saveSettings">保存设置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, FullScreen, Setting, Loading } from '@element-plus/icons-vue'

// Props
interface Props {
  projectId: number
  projectName: string
  filePath?: string
  branch?: string
  readonly?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  branch: 'main',
  readonly: false
})

// Emits
const emit = defineEmits<{
  fileChanged: [filePath: string]
  contentChanged: [content: string]
  saved: [filePath: string, content: string]
}>()

// Refs
const editorContainer = ref<HTMLElement>()
const textEditor = ref<HTMLTextAreaElement>()
const loading = ref(false)
const saving = ref(false)
const hasChanges = ref(false)
const saveDialog = ref(false)
const settingsDialog = ref(false)

// 编辑器相关
let currentContent = ref('')
let originalContent = ref('')
let autoSaveTimer: number | null = null

// 当前状态
const currentFilePath = ref(props.filePath || '')
const currentBranch = ref(props.branch)
const editorHeight = ref(600)

// 表单数据
const saveForm = ref({
  message: ''
})

// 编辑器设置
const editorSettings = ref({
  theme: 'dark',
  fontSize: 14,
  autoSave: true
})

// 监听文件路径变化
watch(() => props.filePath, (newPath) => {
  if (newPath && newPath !== currentFilePath.value) {
    currentFilePath.value = newPath
    loadFileContent()
  }
})

// 监听分支变化
watch(() => props.branch, (newBranch) => {
  if (newBranch && newBranch !== currentBranch.value) {
    currentBranch.value = newBranch
    if (currentFilePath.value) {
      loadFileContent()
    }
  }
})

// 处理内容变化
const handleContentChange = () => {
  hasChanges.value = currentContent.value !== originalContent.value
  emit('contentChanged', currentContent.value)
}

// 处理键盘事件
const handleKeyDown = (event: KeyboardEvent) => {
  // Ctrl+S 保存
  if ((event.ctrlKey || event.metaKey) && event.key === 's') {
    event.preventDefault()
    if (hasChanges.value) {
      saveFile()
    }
  }
  
  // Tab键处理
  if (event.key === 'Tab') {
    event.preventDefault()
    const textarea = textEditor.value
    if (textarea) {
      const start = textarea.selectionStart
      const end = textarea.selectionEnd
      const spaces = '    ' // 4个空格
      
      currentContent.value = currentContent.value.substring(0, start) + spaces + currentContent.value.substring(end)
      
      nextTick(() => {
        textarea.selectionStart = textarea.selectionEnd = start + spaces.length
      })
    }
  }
}

// 加载文件内容
const loadFileContent = async () => {
  if (!currentFilePath.value) return

  loading.value = true
  try {
    const response = await fetch(`/api/projects/${props.projectId}/code/files/content?path=${encodeURIComponent(currentFilePath.value)}&branch=${currentBranch.value}`)
    const data = await response.json()
    
    if (data.success) {
      const content = data.content || ''
      currentContent.value = content
      originalContent.value = content
      hasChanges.value = false
    } else {
      ElMessage.error(data.error || '加载文件失败')
    }
  } catch (error) {
    console.error('Load file error:', error)
    ElMessage.error('加载文件失败')
  } finally {
    loading.value = false
  }
}

// 保存文件
const saveFile = () => {
  if (!hasChanges.value || !currentFilePath.value) return
  
  if (editorSettings.value.autoSave) {
    // 自动保存，直接保存
    confirmSave()
  } else {
    // 手动保存，显示对话框
    saveDialog.value = true
  }
}

// 确认保存
const confirmSave = async () => {
  if (!currentFilePath.value) return

  saving.value = true
  try {
    const response = await fetch(`/api/projects/${props.projectId}/code/files/save`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        file_path: currentFilePath.value,
        branch: currentBranch.value,
        content: currentContent.value,
        message: saveForm.value.message || `Update ${currentFilePath.value}`
      })
    })
    
    const data = await response.json()
    
    if (data.success) {
      originalContent.value = currentContent.value
      hasChanges.value = false
      saveDialog.value = false
      saveForm.value.message = ''
      
      ElMessage.success('文件保存成功')
      emit('saved', currentFilePath.value, currentContent.value)
    } else {
      ElMessage.error(data.error || '保存失败')
    }
  } catch (error) {
    console.error('Save file error:', error)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 切换全屏
const toggleFullscreen = () => {
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    editorContainer.value?.requestFullscreen()
  }
}

// 处理设置命令
const handleSettingsCommand = (command: string) => {
  switch (command) {
    case 'theme':
    case 'fontSize':
      settingsDialog.value = true
      break
  }
}

// 切换自动保存
const toggleAutoSave = (enabled: boolean) => {
  if (enabled) {
    startAutoSave()
  } else {
    stopAutoSave()
  }
}

// 保存设置
const saveSettings = () => {
  localStorage.setItem('codeEditorSettings', JSON.stringify(editorSettings.value))
  settingsDialog.value = false
  ElMessage.success('设置已保存')
}

// 启动自动保存
const startAutoSave = () => {
  stopAutoSave()
  autoSaveTimer = window.setInterval(() => {
    if (hasChanges.value && currentFilePath.value) {
      confirmSave()
    }
  }, 30000) // 30秒
}

// 停止自动保存
const stopAutoSave = () => {
  if (autoSaveTimer) {
    clearInterval(autoSaveTimer)
    autoSaveTimer = null
  }
}

// 加载设置
const loadSettings = () => {
  const saved = localStorage.getItem('codeEditorSettings')
  if (saved) {
    try {
      const settings = JSON.parse(saved)
      editorSettings.value = { ...editorSettings.value, ...settings }
    } catch (error) {
      console.error('Load settings error:', error)
    }
  }
}

// 调整编辑器高度
const adjustEditorHeight = () => {
  if (editorContainer.value) {
    const containerHeight = editorContainer.value.clientHeight
    const toolbarHeight = 50
    editorHeight.value = containerHeight - toolbarHeight
  }
}

// 生命周期
onMounted(async () => {
  loadSettings()
  await nextTick()
  adjustEditorHeight()
  
  // 加载文件内容
  if (currentFilePath.value) {
    await loadFileContent()
  }
  
  // 启动自动保存
  if (editorSettings.value.autoSave) {
    startAutoSave()
  }
  
  // 监听窗口大小变化
  window.addEventListener('resize', adjustEditorHeight)
})

onUnmounted(() => {
  stopAutoSave()
  window.removeEventListener('resize', adjustEditorHeight)
})

// 暴露方法
defineExpose({
  loadFile: loadFileContent,
  saveFile: confirmSave,
  getContent: () => currentContent.value,
  hasUnsavedChanges: () => hasChanges.value
})
</script>

<style scoped>
.code-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #2d2d30;
  border-bottom: 1px solid #3e3e42;
  color: #cccccc;
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

.editor-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.editor-loading {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  gap: 8px;
  color: #cccccc;
  z-index: 10;
}

.text-editor {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  line-height: 1.5;
  padding: 16px;
  background: #1e1e1e;
  color: #d4d4d4;
  tab-size: 4;
  white-space: pre;
  overflow-wrap: normal;
  overflow-x: auto;
}

.text-editor.dark-theme {
  background: #1e1e1e;
  color: #d4d4d4;
}

.text-editor:not(.dark-theme) {
  background: #ffffff;
  color: #333333;
}

.text-editor:focus {
  background: #1e1e1e;
  color: #d4d4d4;
}

.text-editor::selection {
  background: #264f78;
}

.setting-value {
  margin-left: 8px;
  color: #409eff;
  font-weight: 500;
}

.setting-desc {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

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
</style> 
<template>
  <div class="file-tree">
    <div class="tree-header">
      <div class="header-left">
        <el-icon><Folder /></el-icon>
        <span>文件资源管理器</span>
      </div>
      
      <div class="header-right">
        <el-tooltip content="刷新">
          <el-button 
            size="small" 
            text 
            @click="refreshTree"
            :loading="loading"
          >
            <el-icon><Refresh /></el-icon>
          </el-button>
        </el-tooltip>
        
        <el-tooltip content="新建文件">
          <el-button 
            size="small" 
            text 
            @click="showCreateFileDialog"
          >
            <el-icon><DocumentAdd /></el-icon>
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <div class="branch-selector">
      <el-select 
        v-model="selectedBranch" 
        @change="handleBranchChange"
        size="small"
        placeholder="选择分支"
      >
        <el-option 
          v-for="branch in branches" 
          :key="branch.name" 
          :label="branch.name" 
          :value="branch.name"
        >
          <div class="branch-option">
            <el-icon><Document /></el-icon>
            <span>{{ branch.name }}</span>
            <el-tag 
              v-if="branch.name === defaultBranch" 
              size="small" 
              type="success"
            >
              主分支
            </el-tag>
          </div>
        </el-option>
      </el-select>
    </div>

    <div class="tree-container">
      <div class="tree-loading" v-if="loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      
      <div class="tree-content" v-else>
        <div class="tree-empty" v-if="!fileTree || !fileTree.children?.length">
          <el-icon><FolderOpened /></el-icon>
          <span>暂无文件</span>
        </div>
        
        <el-tree
          v-else
          :data="treeData"
          :props="treeProps"
          :expand-on-click-node="false"
          :default-expanded-keys="expandedKeys"
          node-key="path"
          @node-click="handleNodeClick"
          @node-contextmenu="handleNodeContextMenu"
          class="file-tree-component"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <div class="node-content">
                <el-icon class="node-icon">
                  <component :is="getFileIcon(data)" />
                </el-icon>
                
                <span class="node-label">{{ data.name }}</span>
                
                <div class="node-badges">
                  <el-tag 
                    v-if="data.language" 
                    size="small" 
                    type="info"
                    class="language-tag"
                  >
                    {{ data.language }}
                  </el-tag>
                  
                  <el-icon 
                    v-if="!data.is_editable" 
                    class="readonly-icon"
                    title="只读文件"
                  >
                    <Lock />
                  </el-icon>
                </div>
              </div>
            </div>
          </template>
        </el-tree>
      </div>
    </div>

    <!-- 右键菜单 -->
    <el-dropdown
      ref="contextMenuRef"
      trigger="contextmenu"
      :teleported="false"
      @command="handleContextMenuCommand"
    >
      <div></div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item 
            v-if="contextMenuData?.type === 'file'" 
            command="open"
          >
            <el-icon><View /></el-icon>
            打开文件
          </el-dropdown-item>
          
          <el-dropdown-item 
            v-if="contextMenuData?.type === 'file' && contextMenuData?.is_editable" 
            command="edit"
          >
            <el-icon><Edit /></el-icon>
            编辑文件
          </el-dropdown-item>
          
          <el-dropdown-item 
            v-if="contextMenuData?.type === 'directory'" 
            command="newFile"
          >
            <el-icon><DocumentAdd /></el-icon>
            新建文件
          </el-dropdown-item>
          
          <el-dropdown-item 
            v-if="contextMenuData?.type === 'directory'" 
            command="newFolder"
          >
            <el-icon><FolderAdd /></el-icon>
            新建文件夹
          </el-dropdown-item>
          
          <el-dropdown-item 
            command="rename"
            divided
          >
            <el-icon><EditPen /></el-icon>
            重命名
          </el-dropdown-item>
          
          <el-dropdown-item 
            command="delete"
            class="danger-item"
          >
            <el-icon><Delete /></el-icon>
            删除
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- 新建文件对话框 -->
    <el-dialog 
      v-model="createFileDialog" 
      title="新建文件" 
      width="500px"
    >
      <el-form :model="createFileForm" label-width="80px">
        <el-form-item label="文件名">
          <el-input 
            v-model="createFileForm.name" 
            placeholder="请输入文件名"
            @keyup.enter="confirmCreateFile"
          />
        </el-form-item>
        
        <el-form-item label="文件类型">
          <el-select v-model="createFileForm.type" placeholder="选择文件类型">
            <el-option label="JavaScript" value="js" />
            <el-option label="TypeScript" value="ts" />
            <el-option label="Vue" value="vue" />
            <el-option label="HTML" value="html" />
            <el-option label="CSS" value="css" />
            <el-option label="JSON" value="json" />
            <el-option label="Markdown" value="md" />
            <el-option label="Python" value="py" />
            <el-option label="Java" value="java" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="createFileDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmCreateFile">
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder,
  FolderOpened,
  Document,
  DocumentAdd,
  FolderAdd,
  Refresh,
  Loading,
  View,
  Edit,
  EditPen,
  Delete,
  Lock
} from '@element-plus/icons-vue'

// Props
interface Props {
  projectId: number
  projectName: string
  branch?: string
  readonly?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  branch: 'main',
  readonly: false
})

// Emits
const emit = defineEmits<{
  fileSelected: [filePath: string, fileType: string]
  branchChanged: [branch: string]
  fileCreated: [filePath: string]
  fileDeleted: [filePath: string]
  fileRenamed: [oldPath: string, newPath: string]
}>()

// 接口定义
interface FileTreeNode {
  name: string
  path: string
  type: 'file' | 'directory'
  size?: number
  last_modified?: string
  language?: string
  is_editable?: boolean
  children?: FileTreeNode[]
}

interface Branch {
  name: string
  is_default?: boolean
  last_commit?: string
}

// Refs
const loading = ref(false)
const fileTree = ref<FileTreeNode | null>(null)
const selectedBranch = ref(props.branch)
const branches = ref<Branch[]>([])
const defaultBranch = ref('main')
const expandedKeys = ref<string[]>([])
const contextMenuRef = ref()
const contextMenuData = ref<FileTreeNode | null>(null)
const createFileDialog = ref(false)

// 表单数据
const createFileForm = ref({
  name: '',
  type: 'js'
})

// 树形组件配置
const treeProps = {
  children: 'children',
  label: 'name',
  isLeaf: (data: FileTreeNode) => data.type === 'file'
}

// 计算属性
const treeData = computed(() => {
  return fileTree.value?.children || []
})

// 监听分支变化
watch(() => props.branch, (newBranch) => {
  if (newBranch && newBranch !== selectedBranch.value) {
    selectedBranch.value = newBranch
    loadFileTree()
  }
})

// 加载文件树
const loadFileTree = async () => {
  loading.value = true
  try {
    const response = await fetch(`/api/projects/${props.projectId}/code/tree?branch=${selectedBranch.value}`)
    const data = await response.json()
    
    if (data.success) {
      fileTree.value = data.file_tree
      
      // 自动展开根目录
      if (fileTree.value?.children) {
        expandedKeys.value = [fileTree.value.path]
      }
    } else {
      ElMessage.error(data.error || '加载文件树失败')
    }
  } catch (error) {
    console.error('Load file tree error:', error)
    ElMessage.error('加载文件树失败')
  } finally {
    loading.value = false
  }
}

// 加载分支列表
const loadBranches = async () => {
  try {
    const response = await fetch(`/api/projects/${props.projectId}/branches`)
    const data = await response.json()
    
    if (data.success) {
      branches.value = data.branches || []
      
      // 找到默认分支
      const defaultBranchInfo = branches.value.find(b => b.is_default)
      if (defaultBranchInfo) {
        defaultBranch.value = defaultBranchInfo.name
      }
    }
  } catch (error) {
    console.error('Load branches error:', error)
  }
}

// 刷新文件树
const refreshTree = () => {
  loadFileTree()
}

// 处理分支变化
const handleBranchChange = (branch: string) => {
  selectedBranch.value = branch
  loadFileTree()
  emit('branchChanged', branch)
}

// 处理节点点击
const handleNodeClick = (data: FileTreeNode) => {
  if (data.type === 'file') {
    emit('fileSelected', data.path, data.type)
  }
}

// 处理右键菜单
const handleNodeContextMenu = (event: MouseEvent, data: FileTreeNode) => {
  if (props.readonly) return
  
  event.preventDefault()
  contextMenuData.value = data
  
  // 显示右键菜单
  const contextMenu = contextMenuRef.value
  if (contextMenu) {
    contextMenu.handleOpen()
  }
}

// 处理右键菜单命令
const handleContextMenuCommand = (command: string) => {
  if (!contextMenuData.value) return
  
  const data = contextMenuData.value
  
  switch (command) {
    case 'open':
    case 'edit':
      emit('fileSelected', data.path, data.type)
      break
      
    case 'newFile':
      showCreateFileDialog(data.path)
      break
      
    case 'newFolder':
      createFolder(data.path)
      break
      
    case 'rename':
      renameFile(data)
      break
      
    case 'delete':
      deleteFile(data)
      break
  }
}

// 显示新建文件对话框
const showCreateFileDialog = (parentPath?: string) => {
  createFileForm.value = {
    name: '',
    type: 'js'
  }
  createFileDialog.value = true
}

// 确认创建文件
const confirmCreateFile = async () => {
  if (!createFileForm.value.name) {
    ElMessage.warning('请输入文件名')
    return
  }
  
  const fileName = createFileForm.value.name.includes('.') 
    ? createFileForm.value.name 
    : `${createFileForm.value.name}.${createFileForm.value.type}`
  
  try {
    const response = await fetch(`/api/projects/${props.projectId}/code/files`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        file_path: fileName,
        branch: selectedBranch.value,
        content: '',
        message: `Create ${fileName}`
      })
    })
    
    const data = await response.json()
    
    if (data.success) {
      createFileDialog.value = false
      ElMessage.success('文件创建成功')
      
      // 刷新文件树
      await loadFileTree()
      
      // 通知父组件
      emit('fileCreated', fileName)
    } else {
      ElMessage.error(data.error || '创建文件失败')
    }
  } catch (error) {
    console.error('Create file error:', error)
    ElMessage.error('创建文件失败')
  }
}

// 创建文件夹
const createFolder = async (parentPath: string) => {
  try {
    const { value: folderName } = await ElMessageBox.prompt(
      '请输入文件夹名称',
      '新建文件夹',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    
    if (folderName) {
      // 这里可以实现创建文件夹的逻辑
      ElMessage.info('文件夹创建功能待实现')
    }
  } catch (error) {
    // 用户取消
  }
}

// 重命名文件
const renameFile = async (data: FileTreeNode) => {
  try {
    const { value: newName } = await ElMessageBox.prompt(
      '请输入新的文件名',
      '重命名',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: data.name
      }
    )
    
    if (newName && newName !== data.name) {
      // 这里可以实现重命名的逻辑
      ElMessage.info('文件重命名功能待实现')
    }
  } catch (error) {
    // 用户取消
  }
}

// 删除文件
const deleteFile = async (data: FileTreeNode) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除${data.type === 'file' ? '文件' : '文件夹'} "${data.name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 这里可以实现删除的逻辑
    ElMessage.info('文件删除功能待实现')
  } catch (error) {
    // 用户取消
  }
}

// 获取文件图标
const getFileIcon = (data: FileTreeNode) => {
  if (data.type === 'directory') {
    return FolderOpened
  }
  
  // 根据文件扩展名返回不同图标
  const ext = data.name.split('.').pop()?.toLowerCase()
  
  switch (ext) {
    case 'js':
    case 'ts':
    case 'jsx':
    case 'tsx':
      return 'javascript'
    case 'vue':
      return 'vue'
    case 'html':
      return 'html'
    case 'css':
    case 'scss':
    case 'sass':
      return 'css'
    case 'json':
      return 'json'
    case 'md':
      return 'markdown'
    case 'py':
      return 'python'
    case 'java':
      return 'java'
    default:
      return Document
  }
}

// 生命周期
onMounted(() => {
  loadBranches()
  loadFileTree()
})

// 暴露方法
defineExpose({
  refresh: refreshTree,
  loadTree: loadFileTree
})
</script>

<style scoped>
.file-tree {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #252526;
  color: #cccccc;
}

.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #2d2d30;
  border-bottom: 1px solid #3e3e42;
  font-size: 13px;
  font-weight: 500;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 6px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.branch-selector {
  padding: 8px 12px;
  border-bottom: 1px solid #3e3e42;
}

.branch-option {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tree-container {
  flex: 1;
  overflow: auto;
}

.tree-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px;
  color: #cccccc;
}

.tree-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 20px;
  color: #6c6c6c;
}

.tree-node {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 2px 0;
}

.node-content {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}

.node-icon {
  font-size: 16px;
  color: #cccccc;
}

.node-label {
  font-size: 13px;
  color: #cccccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-badges {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
}

.language-tag {
  font-size: 10px;
  height: 16px;
  line-height: 14px;
  padding: 0 4px;
  opacity: 0.7;
}

.readonly-icon {
  font-size: 12px;
  color: #f56c6c;
}

.danger-item {
  color: #f56c6c;
}

/* Element Plus 样式覆盖 */
:deep(.el-tree) {
  background: transparent;
  color: #cccccc;
}

:deep(.el-tree-node__content) {
  height: 28px;
  background: transparent;
  color: #cccccc;
}

:deep(.el-tree-node__content:hover) {
  background: #2a2d2e;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background: #37373d;
}

:deep(.el-tree-node__expand-icon) {
  color: #cccccc;
}

:deep(.el-tree-node__expand-icon.is-leaf) {
  color: transparent;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-select .el-input__inner) {
  background: #3c3c3c;
  border-color: #3c3c3c;
  color: #cccccc;
}

:deep(.el-button--small.is-text) {
  color: #cccccc;
  padding: 4px;
}

:deep(.el-button--small.is-text:hover) {
  color: #409eff;
  background: #2a2d2e;
}
</style> 
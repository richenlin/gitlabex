<template>
  <div class="scene-detail" v-loading="loading">
    <div v-if="project" class="scene-container">
      <!-- 面包屑导航 -->
      <el-breadcrumb class="breadcrumb" separator=">">
        <el-breadcrumb-item :to="{ path: '/scenes' }">课题</el-breadcrumb-item>
        <el-breadcrumb-item>{{ project.name }}</el-breadcrumb-item>
      </el-breadcrumb>

      <!-- 课题头部信息 -->
      <div class="scene-header card">
        <div class="header-content">
          <div class="title-section">
            <h1 class="scene-title">{{ project.name }}</h1>
            <div class="scene-meta">
              <span>创建时间: {{ formatDate(project.created_at) }}</span>
              <span>创建者: {{ project.creator?.name }}</span>
              <el-tag :type="project.is_public ? 'success' : 'warning'" size="small">
                {{ project.is_public ? '公开课题' : '专有课题' }}
              </el-tag>
            </div>
          </div>
          <div class="header-actions" v-if="canManage">
            <el-button @click="editProject">编辑课题</el-button>
            <el-button type="danger" @click="showDeleteDialog">删除课题</el-button>
            <el-button @click="showMembersDialog">管理成员</el-button>
          </div>
        </div>
        <p class="scene-description">{{ project.description }}</p>
      </div>

      <!-- 标签页容器 -->
      <div class="tab-container card">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <!-- 文件标签页 -->
          <el-tab-pane label="文件" name="files">
            <div class="files-content">
              <div class="file-toolbar">
                <el-breadcrumb class="file-breadcrumb" separator="/">
                  <el-breadcrumb-item 
                    v-for="(part, index) in pathParts" 
                    :key="index"
                    @click="navigateToPath(index)"
                    class="breadcrumb-item"
                  >
                    {{ part || 'root' }}
                  </el-breadcrumb-item>
                </el-breadcrumb>
                <div class="file-actions">
                  <el-button size="small" @click="refreshFiles">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                  <el-button size="small" type="primary" @click="openWebIDE" v-if="canEdit">
                    <el-icon><Edit /></el-icon>
                    在Web IDE中编辑
                  </el-button>
                </div>
              </div>

              <div class="file-explorer">
                <div class="file-list" v-loading="filesLoading">
                  <!-- 文件列表表头 -->
                  <div class="file-header">
                    <div class="file-name-col">名称</div>
                    <div class="file-commit-col">最后提交</div>
                    <div class="file-update-col">更新时间</div>
                    <div class="file-size-col">大小</div>
                  </div>
                  <!-- 文件列表内容 -->
                  <div 
                    v-for="file in files" 
                    :key="file.name"
                    class="file-item"
                    @click="handleFileClick(file)"
                    @dblclick="handleFileDoubleClick(file)"
                  >
                    <div class="file-name-col">
                      <el-icon class="file-icon">
                        <Folder v-if="file.type === 'tree'" />
                        <Document v-else />
                      </el-icon>
                      <span class="file-name">{{ file.name }}</span>
                    </div>
                    <div class="file-commit-col">
                      <span class="commit-message">{{ file.last_commit_message || '-' }}</span>
                    </div>
                    <div class="file-update-col">
                      <span class="update-time">{{ formatRelativeTime(file.last_commit_date) }}</span>
                    </div>
                    <div class="file-size-col">
                      <span class="file-size">{{ formatFileSize(file.size || 0) }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 文件内容预览 -->
              <div class="file-preview" v-if="selectedFile && fileContent">
                <div class="preview-header">
                  <h3>{{ selectedFile.name }}</h3>
                  <div class="preview-actions">
                    <el-button size="small" @click="downloadFile" v-if="selectedFile.type === 'blob'">
                      <el-icon><Download /></el-icon>
                      下载
                    </el-button>
                    <el-button size="small" type="primary" @click="editFile" v-if="canEdit && isEditableFile(selectedFile)">
                      <el-icon><Edit /></el-icon>
                      编辑
                    </el-button>
                  </div>
                </div>
                <div class="preview-content">
                  <pre v-if="isTextFile(selectedFile)" class="code-preview"><code>{{ fileContent }}</code></pre>
                  <div v-else class="binary-file-info">
                    <el-icon><Document /></el-icon>
                    <p>二进制文件，无法预览</p>
                    <el-button @click="downloadFile">下载文件</el-button>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <!-- 话题标签页 -->
          <el-tab-pane label="话题" name="topics">
            <div class="topics-content">
              <div class="topics-header">
                <h3>课题话题讨论</h3>
                <div class="topics-actions">
                  <el-button type="primary" @click="showCreateTopicDialog" v-if="canCreateTopic">
                    发起话题
                  </el-button>
                </div>
              </div>

              <div class="topics-list" v-loading="topicsLoading">
                <div 
                  v-for="topic in topics" 
                  :key="topic.id"
                  class="topic-item"
                  @click="showTopicDetail(topic)"
                >
                  <div class="topic-status">
                    <div class="topic-stats">
                      <div class="stat-item">
                        <span class="stat-value">{{ topic.comments_count || 0 }}</span>
                        <span class="stat-label">回复</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-value">{{ (topic as any).view_count || 0 }}</span>
                        <span class="stat-label">浏览</span>
                      </div>
                    </div>
                  </div>
                  <div class="topic-content">
                    <h3 class="topic-title">{{ topic.title }}</h3>
                    <div class="topic-meta">
                      <span class="topic-author">
                        <img :src="topic.author?.avatar_url || '/default-avatar.png'" alt="用户头像">
                        {{ topic.author?.name }}
                      </span>
                      <span class="topic-time">{{ formatDate(topic.created_at) }}</span>
                      <el-tag 
                        v-for="label in topic.labels?.slice(0, 2)" 
                        :key="label" 
                        size="small"
                        class="topic-tag"
                      >
                        {{ label }}
                      </el-tag>
                    </div>
                    <div class="topic-excerpt">
                      <p>{{ topic.content?.substring(0, 100) }}...</p>
                    </div>
                    
                    <!-- 话题操作区域 -->
                    <div class="topic-actions">
                      <el-button 
                        size="small" 
                        :type="topic.user_liked ? 'primary' : 'default'"
                        :icon="topic.user_liked ? 'StarFilled' : 'Star'"
                        @click.stop="toggleLike(topic)"
                        :loading="topic.liking"
                      >
                        点赞 {{ topic.like_count || 0 }}
                      </el-button>
                      
                      <el-button 
                        size="small" 
                        :type="topic.user_disliked ? 'danger' : 'default'"
                        icon="CircleClose"
                        @click.stop="toggleDislike(topic)"
                        :loading="topic.disliking"
                      >
                        反对 {{ topic.dislike_count || 0 }}
                      </el-button>
                    </div>
                  </div>
                  <div class="topic-last-reply">
                    <div class="last-reply-info">
                      <span>最后回复：{{ formatDate(topic.updated_at) }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <el-pagination
                v-if="topicsTotal > topicsPageSize"
                v-model:current-page="topicsCurrentPage"
                :total="topicsTotal"
                :page-size="topicsPageSize"
                layout="prev, pager, next"
                @current-change="fetchTopics"
              />
            </div>
          </el-tab-pane>

          <!-- 作业标签页 -->
          <el-tab-pane label="作业" name="homework">
            <div class="homework-content">
              <div class="homework-header">
                <h3>课题作业</h3>
                <el-button type="primary" @click="showCreateHomeworkDialog" v-if="canCreateHomework">
                  创建作业
                </el-button>
              </div>

              <div class="homework-list" v-loading="homeworkLoading">
                <div 
                  v-for="homework in homeworks" 
                  :key="homework.id"
                  class="homework-item"
                >
                  <div class="homework-info" @click="viewHomework(homework.id)" style="cursor: pointer;">
                    <h4 class="homework-title">{{ homework.title }}</h4>
                    <div class="homework-meta">
                      <span>截止日期: {{ getHomeworkDueDate(homework) }}</span>
                      <span>状态: {{ getHomeworkStatusText(homework.status) }}</span>
                      <span>已提交人数: {{ homework.submissions?.length || 0 }}/{{ members.length }}</span>
                    </div>
                    <div class="homework-description">
                      <p>{{ homework.description }}</p>
                    </div>
                  </div>
                  <div class="homework-actions">
                    <el-button size="small" @click="viewHomework(homework.id)">查看详情</el-button>
                    <el-button 
                      size="small" 
                      type="primary" 
                      @click="submitHomework(homework.id)"
                      v-if="canSubmitHomework(homework)"
                    >
                      提交作业
                    </el-button>
                    <el-dropdown trigger="click" v-if="canGradeHomework">
                      <el-button size="small">
                        批改作业<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                      </el-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item @click="viewAllSubmissions(homework.id)">
                            查看所有提交
                          </el-dropdown-item>
                          <el-dropdown-item 
                            v-for="submission in getSubmissionsForHomework(homework.id)" 
                            :key="submission.id"
                            @click="gradeSubmission(submission.id)"
                          >
                            批改 {{ submission.student?.name || '学生' }} 的提交
                          </el-dropdown-item>
                          <el-dropdown-item v-if="getSubmissionsForHomework(homework.id).length === 0" disabled>
                            暂无提交
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </div>
                </div>
              </div>

              <!-- 作业分页 -->
              <el-pagination
                v-if="homeworkTotal > homeworkPageSize"
                v-model:current-page="homeworkCurrentPage"
                :total="homeworkTotal"
                :page-size="homeworkPageSize"
                layout="prev, pager, next"
                @current-change="fetchHomeworks"
                style="margin-top: 20px; justify-content: center;"
              />
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <!-- 成员管理对话框 -->
    <el-dialog v-model="membersDialogVisible" title="课题成员管理" width="600px">
      <div class="member-management">
        <div class="member-search">
          <el-input 
            v-model="memberSearchQuery" 
            placeholder="搜索用户..." 
            @input="searchUsers"
          />
          <el-button type="primary" @click="showAddMemberDialog">添加成员</el-button>
        </div>
        <div class="member-list">
          <div 
            v-for="member in members" 
            :key="member.id"
            class="member-item"
          >
            <div class="member-info">
              <div class="member-avatar">
                <img :src="member.avatar_url || '/default-avatar.png'" alt="用户头像">
              </div>
              <div class="member-details">
                <h4>{{ member.name || member.username }}</h4>
                <span :class="`member-role access-level-${member.access_level}`">
                  {{ getAccessLevelText(member.access_level) }}
                </span>
              </div>
            </div>
            <div class="member-actions">
              <el-button 
                size="small" 
                type="danger"
                @click="removeMember(member.id)"
                :disabled="member.access_level === 50"
                v-if="canManage"
              >
                移除
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 创建话题对话框 -->
    <el-dialog v-model="createTopicDialogVisible" title="发起新话题" width="600px">
      <el-form :model="topicForm" label-width="80px">
        <el-form-item label="话题标题">
          <el-input v-model="topicForm.title" placeholder="请输入话题标题" />
        </el-form-item>
        <el-form-item label="话题内容">
          <el-input 
            v-model="topicForm.content" 
            type="textarea" 
            :rows="6"
            placeholder="请输入话题内容"
          />
        </el-form-item>
        <el-form-item label="标签">
          <el-select 
            v-model="topicForm.labels" 
            multiple 
            filterable 
            allow-create
            placeholder="请选择或输入标签"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createTopicDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createTopic" :loading="creatingTopic">发布</el-button>
      </template>
    </el-dialog>

    <!-- 创建作业对话框 -->
    <el-dialog v-model="createHomeworkDialogVisible" title="创建新作业" width="600px">
      <el-form :model="homeworkForm" label-width="80px">
        <el-form-item label="作业标题">
          <el-input v-model="homeworkForm.title" placeholder="请输入作业标题" />
        </el-form-item>
        <el-form-item label="作业描述">
          <el-input 
            v-model="homeworkForm.description" 
            type="textarea" 
            :rows="4"
            placeholder="请输入作业描述"
          />
        </el-form-item>
        <el-form-item label="截止日期">
          <el-date-picker 
            v-model="homeworkForm.deadline" 
            type="datetime"
            placeholder="选择截止日期"
          />
        </el-form-item>
        <el-form-item label="满分">
          <el-input-number v-model="homeworkForm.max_grade" :min="1" :max="1000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createHomeworkDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createHomework" :loading="creatingHomework">创建</el-button>
      </template>
    </el-dialog>

    <!-- 删除课题确认对话框 -->
    <el-dialog v-model="deleteDialogVisible" title="⚠️ 危险操作：删除课题" width="500px">
      <div class="delete-warning">
        <el-icon class="warning-icon" color="#F56C6C" size="64">
          <WarningFilled />
        </el-icon>
        <p class="warning-text">
          确定要<strong>永久删除</strong>课题 <strong>"{{ project?.name }}"</strong> 吗？
        </p>
        
        <div class="warning-details">
          <p class="warning-subtitle">⚠️ 此操作将导致以下数据被永久删除：</p>
          <ul class="warning-list">
            <li>📝 课题的所有话题讨论</li>
            <li>📋 课题的所有作业和提交</li>
            <li>📁 课题的所有文档资料</li>
            <li>👥 课题的成员关系</li>
            <li>🔗 与GitLab项目的关联</li>
          </ul>
          <p class="warning-emphasis">
            <strong>注意：</strong>删除后这些数据将无法恢复，请谨慎操作！
          </p>
        </div>
        
        <div class="confirmation-input">
          <p class="input-label">请输入课题名称以确认删除：</p>
          <el-input 
            v-model="deleteConfirmInput" 
            placeholder="请输入课题名称"
            :disabled="deleting"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="cancelDelete" :disabled="deleting">取消</el-button>
        <el-button 
          type="danger" 
          @click="confirmDelete" 
          :loading="deleting"
          :disabled="deleteConfirmInput !== project?.name"
        >
          {{ deleting ? '删除中...' : '确认永久删除' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 回复话题对话框 -->
    <el-dialog v-model="replyDialogVisible" title="回复话题" width="500px">
      <div class="reply-context" v-if="replyingTopic">
        <h4>{{ replyingTopic.title }}</h4>
        <p class="context-excerpt">{{ replyingTopic.content?.substring(0, 200) }}...</p>
      </div>
      <el-form :model="replyForm" label-width="80px">
        <el-form-item label="回复内容">
          <el-input 
            v-model="replyForm.content" 
            type="textarea" 
            :rows="4"
            placeholder="请输入回复内容..."
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitReply" :loading="submittingReply">
          {{ submittingReply ? '提交中...' : '提交回复' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 话题详情模态框 -->
    <el-dialog 
      v-model="topicDetailDialogVisible" 
      :title="currentTopicDetail?.title || '话题详情'" 
      width="80%" 
      max-width="800px"
      top="5vh"
    >
      <div v-loading="topicDetailLoading">
        <!-- 话题内容 -->
        <div v-if="currentTopicDetail" class="topic-detail">
          <div class="topic-header">
            <div class="topic-meta">
              <el-avatar 
                :src="currentTopicDetail.author?.avatar_url" 
                :size="32"
                class="author-avatar"
              />
              <div class="author-info">
                <span class="author-name">{{ currentTopicDetail.author?.name }}</span>
                <span class="publish-time">{{ formatDate(currentTopicDetail.created_at) }}</span>
              </div>
            </div>
            <div class="topic-labels" v-if="currentTopicDetail.labels?.length">
              <el-tag 
                v-for="label in currentTopicDetail.labels" 
                :key="label" 
                size="small"
                class="topic-label"
              >
                {{ label }}
              </el-tag>
            </div>
          </div>
          
          <div class="topic-content">
            <p>{{ currentTopicDetail.content }}</p>
          </div>
          
          <div class="topic-stats">
            <span class="stat-item">
              <el-icon><ChatDotRound /></el-icon>
              {{ currentTopicDetail.comments_count || 0 }} 回复
            </span>
            <span class="stat-item">
              <el-icon><CaretTop /></el-icon>
              {{ currentTopicDetail.like_count || 0 }} 点赞
            </span>
            <span class="stat-item">
              <el-icon><CaretBottom /></el-icon>
              {{ currentTopicDetail.dislike_count || 0 }} 反对
            </span>
          </div>
        </div>
        
        <!-- 回复输入框 -->
        <div class="reply-section" v-if="userStore.isLoggedIn">
          <h4 class="reply-title">添加回复</h4>
          <el-input 
            v-model="replyForm.content" 
            type="textarea" 
            :rows="3"
            placeholder="请输入回复内容..."
            maxlength="1000"
            show-word-limit
            class="reply-input"
          />
          <div class="reply-actions">
            <el-button @click="replyForm.content = ''">清空</el-button>
            <el-button type="primary" @click="submitReply" :loading="submittingReply">
              {{ submittingReply ? '提交中...' : '提交回复' }}
            </el-button>
          </div>
        </div>
        
        <!-- 回复列表 -->
        <div class="comments-section" v-if="topicComments.length > 0 || commentsTotal > 0">
          <h4 class="comments-title">回复 ({{ commentsTotal || topicComments.length }})</h4>
          <div class="comments-list" v-loading="commentsLoading">
            <div 
              v-for="comment in topicComments" 
              :key="comment.id" 
              class="comment-item"
            >
              <el-avatar 
                :src="comment.author?.avatar_url" 
                :size="28"
                class="comment-avatar"
              />
              <div class="comment-content">
                <div class="comment-header">
                  <span class="comment-author">{{ comment.author?.name }}</span>
                  <span class="comment-time">{{ formatDate(comment.created_at) }}</span>
                </div>
                <div class="comment-body">
                  {{ comment.content }}
                </div>
              </div>
            </div>
          </div>
          
          <!-- 回复分页 -->
          <div class="comments-pagination" v-if="commentsTotal > commentsPageSize">
            <el-pagination
              v-model:current-page="commentsCurrentPage"
              :total="commentsTotal"
              :page-size="commentsPageSize"
              layout="prev, pager, next"
              @current-change="handleCommentsPageChange"
            />
          </div>
        </div>
        
        <div v-else class="login-prompt">
          <p>请先登录后再回复话题</p>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { researchService, gitlabService, topicService, homeworkService } from '@/services/api'
import type { ResearchProject, GitLabProjectMember, Topic, Homework } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Folder, 
  Document, 
  Refresh, 
  Upload, 
  Download, 
  Edit,
  ArrowDown,
  Plus,
  Delete,
  WarningFilled,
  Star,
  StarFilled,
  CircleClose,
  ChatDotRound
} from '@element-plus/icons-vue'
import { handleApiError, showSuccess } from '@/utils/errorHandler'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const project = ref<ResearchProject | null>(null)
const members = ref<GitLabProjectMember[]>([])
const topics = ref<Topic[]>([])
const homeworks = ref<Homework[]>([])
const files = ref<any[]>([])
const selectedFile = ref<any>(null)
const fileContent = ref('')

const loading = ref(false)
const filesLoading = ref(false)
const topicsLoading = ref(false)
const homeworkLoading = ref(false)

const activeTab = ref('files')
const currentPath = ref('')
const topicFilter = ref('all')
const topicsCurrentPage = ref(1)
const topicsPageSize = ref(10)
const topicsTotal = ref(0)

// 作业分页相关
const homeworkCurrentPage = ref(1)
const homeworkPageSize = ref(10)
const homeworkTotal = ref(0)

// 对话框状态
const membersDialogVisible = ref(false)
const createTopicDialogVisible = ref(false)
const createHomeworkDialogVisible = ref(false)
const deleteDialogVisible = ref(false)
const deleting = ref(false)
const deleteConfirmInput = ref('')
const replyDialogVisible = ref(false)
const submittingReply = ref(false)
const replyingTopic = ref<Topic | null>(null)
const topicDetailDialogVisible = ref(false)
const topicDetailLoading = ref(false)
const currentTopicDetail = ref<any>(null)
const topicComments = ref<any[]>([])
const memberSearchQuery = ref('')
const creatingTopic = ref(false)
const creatingHomework = ref(false)

// 表单数据
const topicForm = ref({
  title: '',
  content: '',
  labels: [] as string[]
})

const homeworkForm = ref({
  title: '',
  description: '',
  deadline: '',
  max_grade: 100
})

const replyForm = ref({
  content: ''
})

// 回复分页相关
const commentsCurrentPage = ref(1)
const commentsPageSize = ref(10)
const commentsTotal = ref(0)
const commentsLoading = ref(false)

// 计算属性
const projectId = computed(() => route.params.id as string)

const pathParts = computed(() => {
  return currentPath.value ? currentPath.value.split('/') : ['']
})

const canManage = ref(false)
const canEdit = ref(false)
const canCreateTopic = ref(false)
const canCreateHomework = ref(false)
const canGradeHomework = ref(false)

// 用户角色状态（由后端权限检查接口返回）
const isStudent = ref(false)
const isTeacher = ref(false)
const userRoles = ref<string[]>([])

// 检查权限的方法
const checkPermissions = async () => {
  if (!userStore.isLoggedIn || !project.value) {
    canManage.value = false
    canEdit.value = false
    canCreateTopic.value = false
    canCreateHomework.value = false
    canGradeHomework.value = false
    return
  }

  try {
    // 并行检查多个权限，使用原有的权限检查方式
    const [managePermission, editPermission, topicPermission, homeworkPermission, gradePermission] = await Promise.all([
      userStore.checkProjectPermission(projectId.value, 'manage'),
      userStore.checkPermission('update', 'project', projectId.value),
      userStore.checkPermission('create', 'topic'),
      userStore.checkPermission('create', 'homework'),
      userStore.checkPermission('grade', 'homework')
    ])

    canManage.value = managePermission
    canEdit.value = editPermission
    canCreateTopic.value = topicPermission
    canCreateHomework.value = homeworkPermission
    canGradeHomework.value = gradePermission
    
    // 同时获取详细的角色信息用于UI显示
    try {
      const detailedResponse = await (userStore as any).checkProjectPermissionDetailed(projectId.value)
      const roles = detailedResponse.roles || []
      userRoles.value = roles
      isStudent.value = roles.includes('student') || roles.includes('developer') || roles.includes('reporter')
      isTeacher.value = roles.includes('teacher') || roles.includes('maintainer') || roles.includes('owner') || userStore.isAdmin
    } catch (detailError) {
      console.log('获取详细角色信息失败，使用基本角色判断:', detailError)
      // 如果获取详细信息失败，使用基本的角色判断
      isStudent.value = !userStore.isAdmin && !canManage.value
      isTeacher.value = userStore.isAdmin || canManage.value || canGradeHomework.value
      userRoles.value = []
    }
    
  } catch (error) {
    console.error('权限检查失败:', error)
    // 权限检查失败时，默认为无权限
    canManage.value = false
    canEdit.value = false
    canCreateTopic.value = false
    canCreateHomework.value = false
    canGradeHomework.value = false
    isStudent.value = false
    isTeacher.value = false
    userRoles.value = []
  }
}

// 权限相关的计算属性已移至上面的响应式变量和checkPermissions方法

// 方法
const fetchProject = async () => {
  loading.value = true
  try {
    const response: any = await researchService.getProject(projectId.value)
    project.value = response
    
    // 项目加载完成后检查权限
    await checkPermissions()
    
    // 项目加载完成后再加载文件列表
    if (project.value?.gitlab_project_id) {
      fetchFiles()
    }
  } catch (error) {
    console.error('获取课题详情失败:', error)
    handleApiError(error, '获取课题详情')
  } finally {
    loading.value = false
  }
}

const fetchMembers = async () => {
  try {
    const response: any = await researchService.getMembers(projectId.value)
    // 后端现在返回GitLab项目成员，格式为 {members: [...], count: number}
    members.value = response.members || []
  } catch (error) {
    console.error('获取成员列表失败:', error)
    handleApiError(error, '获取成员列表')
    members.value = []
  }
}

const fetchFiles = async (path = '') => {
  if (!project.value?.gitlab_project_id) return
  
  filesLoading.value = true
  try {
    const response: any = await gitlabService.getRepositoryTree(
      project.value.gitlab_project_id.toString(),
      path
    )
    // 后端返回的数据格式是 {"tree": [...]}，需要提取 tree 字段
    files.value = response?.tree || response || []
    currentPath.value = path
    console.log('获取文件列表成功:', files.value)
  } catch (error: any) {
    console.error('获取文件列表失败:', error)
    // 如果是GitLab相关的认证错误，不显示错误提示，避免干扰用户体验
    if (error.response?.status !== 401) {
      handleApiError(error, '获取文件列表')
    }
  } finally {
    filesLoading.value = false
  }
}

const fetchTopics = async () => {
  topicsLoading.value = true
  try {
    const response: any = await topicService.getTopics({
      project_id: projectId.value,
      page: topicsCurrentPage.value,
      limit: topicsPageSize.value
    })
    topics.value = response.topics || []
    topicsTotal.value = response.pagination?.total || 0
  } catch (error) {
    console.error('获取话题列表失败:', error)
    handleApiError(error, '获取话题列表')
  } finally {
    topicsLoading.value = false
  }
}

const fetchHomeworks = async () => {
  homeworkLoading.value = true
  try {
    const response: any = await homeworkService.getHomeworks({
      projectId: projectId.value,
      page: homeworkCurrentPage.value,
      pageSize: homeworkPageSize.value
    })
    homeworks.value = Array.isArray(response) ? response : (response.homeworks || [])
    homeworkTotal.value = response.pagination?.total || response.total || homeworks.value.length
  } catch (error) {
    console.error('获取作业列表失败:', error)
  } finally {
    homeworkLoading.value = false
  }
}

const handleTabChange = (tabName: string) => {
  activeTab.value = tabName
  switch (tabName) {
    case 'files':
      fetchFiles()
      break
    case 'topics':
      fetchTopics()
      break
    case 'homework':
      fetchHomeworks()
      break
  }
}

const handleFileClick = async (file: any) => {
  selectedFile.value = file
  if (file.type === 'blob') {
    // 打开GitLab IDE而不是加载文件内容
    await openFileInIDE(file)
  }
}

const handleFileDoubleClick = (file: any) => {
  if (file.type === 'tree') {
    const newPath = currentPath.value ? `${currentPath.value}/${file.name}` : file.name
    fetchFiles(newPath)
  }
}

const openFileInIDE = async (file: any) => {
  if (!project.value?.id) return
  
  try {
    const filePath = currentPath.value ? `${currentPath.value}/${file.name}` : file.name
    const response: any = await researchService.getGitLabIDEURL(
      project.value.id,
      filePath
    )
    
    if (response.data?.ide_url) {
      // 在新窗口打开GitLab IDE
      window.open(response.data.ide_url, '_blank')
    } else {
      ElMessage.error('无法获取IDE链接')
    }
  } catch (error) {
    console.error('获取IDE链接失败:', error)
    ElMessage.error('获取IDE链接失败')
  }
}

const loadFileContent = async (file: any) => {
  if (!project.value?.gitlab_project_id) return
  
  try {
    const filePath = currentPath.value ? `${currentPath.value}/${file.name}` : file.name
    const response: any = await gitlabService.getFileContent(
      project.value.gitlab_project_id.toString(),
      filePath
    )
    fileContent.value = response?.content || ''
  } catch (error) {
    console.error('获取文件内容失败:', error)
    handleApiError(error, '获取文件内容')
  }
}

const navigateToPath = (index: number) => {
  const newPath = pathParts.value.slice(0, index + 1).join('/')
  fetchFiles(newPath)
}

const refreshFiles = () => {
  fetchFiles(currentPath.value)
}

const isTextFile = (file: any) => {
  const textExtensions = ['.txt', '.md', '.js', '.ts', '.vue', '.html', '.css', '.json', '.xml', '.yml', '.yaml']
  return textExtensions.some(ext => file.name.toLowerCase().endsWith(ext))
}

const isEditableFile = (file: any) => {
  return isTextFile(file) && file.type === 'blob'
}

const showMembersDialog = () => {
  membersDialogVisible.value = true
  fetchMembers()
}

const showCreateTopicDialog = () => {
  createTopicDialogVisible.value = true
  topicForm.value = {
    title: '',
    content: '',
    labels: []
  }
}

const showCreateHomeworkDialog = () => {
  createHomeworkDialogVisible.value = true
  homeworkForm.value = {
    title: '',
    description: '',
    deadline: '',
    max_grade: 100
  }
}

const createTopic = async () => {
  creatingTopic.value = true
  try {
    await topicService.createTopic({
      title: topicForm.value.title,
      content: topicForm.value.content,
      labels: topicForm.value.labels,
      project_id: projectId.value
    })
    showSuccess('话题创建成功')
    createTopicDialogVisible.value = false
    fetchTopics()
  } catch (error) {
    console.error('创建话题失败:', error)
    handleApiError(error, '创建话题')
  } finally {
    creatingTopic.value = false
  }
}

const createHomework = async () => {
  creatingHomework.value = true
  try {
    await homeworkService.createHomework({
      title: homeworkForm.value.title,
      description: homeworkForm.value.description,
      project_id: projectId.value,
      deadline: homeworkForm.value.deadline,
      max_grade: homeworkForm.value.max_grade
    })
    showSuccess('作业创建成功')
    createHomeworkDialogVisible.value = false
    fetchHomeworks()
  } catch (error) {
    console.error('创建作业失败:', error)
    handleApiError(error, '创建作业')
  } finally {
    creatingHomework.value = false
  }
}

const viewTopic = (topicId: string) => {
  router.push(`/topics/${topicId}`)
}

const viewHomework = (homeworkId: string) => {
  router.push(`/homeworks/${homeworkId}`)
}

const submitHomework = (homeworkId: string) => {
  router.push(`/homeworks/${homeworkId}/submit`)
}

const gradeHomework = (homeworkId: string) => {
  // 这个方法保留用于兼容性，但推荐使用具体的提交批改
  router.push(`/homeworks/${homeworkId}/grade`)
}

// 新增方法：查看所有提交
const viewAllSubmissions = (homeworkId: string) => {
  router.push(`/homeworks/${homeworkId}/submissions`)
}

// 新增方法：批改具体的提交
const gradeSubmission = (submissionId: string) => {
  router.push(`/homeworks/submissions/${submissionId}/grade`)
}

// 新增方法：获取特定作业的提交列表
const getSubmissionsForHomework = (homeworkId: string) => {
  // 从homeworks数组中找到对应作业的提交
  const homework = homeworks.value.find(h => h.id === homeworkId)
  return homework?.submissions?.filter(s => s.status !== 'pending') || []
}

// 获取当前用户对某个作业的提交状态
const getMySubmissionStatus = (homework: Homework) => {
  if (!userStore.user?.id) return null
  const userId = userStore.user.id.toString()
  return homework.submissions?.find(s => s.student_id === userId)
}

// 获取作业截止日期的显示文本
const getHomeworkDueDate = (homework: Homework) => {
  const dueDate = homework.deadline || homework.due_date
  return dueDate ? formatDate(dueDate) : '无限制'
}

const canSubmitHomework = (homework: Homework) => {
  if (!userStore.isLoggedIn || homework.status !== 'published') {
    return false
  }
  
  // 检查是否已过截止日期
  const dueDate = homework.deadline || homework.due_date
  if (dueDate && new Date(dueDate) < new Date()) {
    return false
  }
  
  // 检查是否已经提交过
  const mySubmission = getMySubmissionStatus(homework)
  return !mySubmission || mySubmission.status === 'pending'
}

const editProject = () => {
  router.push(`/scenes/${projectId.value}/edit`)
}

const showDeleteDialog = () => {
  deleteConfirmInput.value = ''
  deleteDialogVisible.value = true
}

const cancelDelete = () => {
  deleteConfirmInput.value = ''
  deleteDialogVisible.value = false
}

const confirmDelete = async () => {
  if (!project.value) return
  
  // 验证输入的课题名称
  if (deleteConfirmInput.value !== project.value?.name) {
    ElMessage.error('课题名称输入错误，请重新输入')
    return
  }
  
  deleting.value = true
  try {
    await researchService.deleteProject(projectId.value)
    showSuccess('课题已永久删除')
    router.push('/scenes')
  } catch (error) {
    console.error('删除课题失败:', error)
    handleApiError(error, '删除课题')
  } finally {
    deleting.value = false
    deleteDialogVisible.value = false
    deleteConfirmInput.value = ''
  }
}

const removeMember = async (userId: number) => {
  try {
    await researchService.removeMember(projectId.value, userId.toString())
    showSuccess('成员移除成功')
    fetchMembers()
  } catch (error) {
    console.error('移除成员失败:', error)
    handleApiError(error, '移除成员')
  }
}

// 获取GitLab访问级别对应的文本
const getAccessLevelText = (accessLevel: number): string => {
  switch (accessLevel) {
    case 10: return 'Guest'
    case 20: return 'Reporter'
    case 30: return 'Developer'
    case 40: return 'Maintainer'
    case 50: return 'Owner'
    default: return 'Unknown'
  }
}

// 话题操作相关方法
const toggleLike = async (topic: Topic) => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  
  if (!project.value?.id) {
    ElMessage.error('课题信息获取失败')
    return
  }
  
  topic.liking = true
  try {
    if (topic.user_liked) {
      await topicService.unlikeTopic(topic.id, project.value.id)
      topic.user_liked = false
      topic.like_count = (topic.like_count || 0) - 1
      showSuccess('取消点赞成功')
    } else {
      await topicService.likeTopic(topic.id, project.value.id)
      topic.user_liked = true
      topic.like_count = (topic.like_count || 0) + 1
      // 如果之前反对了，取消反对
      if (topic.user_disliked) {
        await topicService.undislikeTopic(topic.id, project.value.id)
        topic.user_disliked = false
        topic.dislike_count = Math.max(0, (topic.dislike_count || 0) - 1)
      }
      showSuccess('点赞成功')
    }
  } catch (error) {
    console.error('点赞操作失败:', error)
    handleApiError(error, '点赞操作')
  } finally {
    topic.liking = false
  }
}

const toggleDislike = async (topic: Topic) => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  
  if (!project.value?.id) {
    ElMessage.error('课题信息获取失败')
    return
  }
  
  topic.disliking = true
  try {
    if (topic.user_disliked) {
      await topicService.undislikeTopic(topic.id, project.value.id)
      topic.user_disliked = false
      topic.dislike_count = Math.max(0, (topic.dislike_count || 0) - 1)
      showSuccess('取消反对成功')
    } else {
      await topicService.dislikeTopic(topic.id, project.value.id)
      topic.user_disliked = true
      topic.dislike_count = (topic.dislike_count || 0) + 1
      // 如果之前点赞了，取消点赞
      if (topic.user_liked) {
        await topicService.unlikeTopic(topic.id, project.value.id)
        topic.user_liked = false
        topic.like_count = Math.max(0, (topic.like_count || 0) - 1)
      }
      showSuccess('反对成功')
    }
  } catch (error) {
    console.error('反对操作失败:', error)
    handleApiError(error, '反对操作')
  } finally {
    topic.disliking = false
  }
}

const showTopicDetail = async (topic: Topic) => {
  if (!project.value?.id) {
    ElMessage.error('课题信息获取失败')
    return
  }
  
  topicDetailLoading.value = true
  topicDetailDialogVisible.value = true
  
  try {
    const response: any = await topicService.getTopic(topic.id, project.value.id)
    currentTopicDetail.value = response.topic
    replyingTopic.value = currentTopicDetail.value
    
    // 重置分页并获取评论
    commentsCurrentPage.value = 1
    await fetchComments()
  } catch (error) {
    console.error('获取话题详情失败:', error)
    handleApiError(error, '获取话题详情')
    topicDetailDialogVisible.value = false
  } finally {
    topicDetailLoading.value = false
  }
}

const fetchComments = async () => {
  if (!currentTopicDetail.value || !project.value?.id) {
    return
  }
  
  commentsLoading.value = true
  try {
    const response: any = await topicService.getTopic(
      currentTopicDetail.value.id, 
      project.value.id
    )
    const allComments = response.comments || []
    commentsTotal.value = allComments.length
    
    // 前端分页逻辑
    const startIndex = (commentsCurrentPage.value - 1) * commentsPageSize.value
    const endIndex = startIndex + commentsPageSize.value
    topicComments.value = allComments.slice(startIndex, endIndex)
  } catch (error) {
    console.error('获取评论失败:', error)
  } finally {
    commentsLoading.value = false
  }
}

const handleCommentsPageChange = (page: number) => {
  commentsCurrentPage.value = page
  fetchComments()
}

const showReplyDialog = (topic: Topic) => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  
  replyingTopic.value = topic
  replyForm.value.content = ''
  replyDialogVisible.value = true
}

const submitReply = async () => {
  if (!replyForm.value.content.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  
  if (!replyingTopic.value) return
  
  if (!project.value?.id) {
    ElMessage.error('课题信息获取失败')
    return
  }
  
  submittingReply.value = true
  try {
    const response: any = await topicService.createComment(replyingTopic.value.id, replyForm.value.content, project.value.id)
    showSuccess('回复成功')
    replyForm.value.content = ''
    
    // 如果在话题详情模态框中，更新评论列表
    if (topicDetailDialogVisible.value && response.comment) {
      if (currentTopicDetail.value) {
        currentTopicDetail.value.comments_count = (currentTopicDetail.value.comments_count || 0) + 1
      }
      
      // 计算新评论可能所在的页面
      const newTotal = currentTopicDetail.value?.comments_count || 1
      const lastPage = Math.ceil(newTotal / commentsPageSize.value)
      
      // 如果新评论在最后一页，跳转到最后一页
      if (lastPage > commentsCurrentPage.value) {
        commentsCurrentPage.value = lastPage
      }
      
      // 刷新评论列表
      await fetchComments()
    }
    
    // 更新话题列表中的回复数量
    replyingTopic.value.comments_count = (replyingTopic.value.comments_count || 0) + 1
    
    // 如果不在话题详情模态框中，关闭老的回复对话框
    if (!topicDetailDialogVisible.value) {
      replyDialogVisible.value = false
    }
  } catch (error) {
    console.error('提交回复失败:', error)
    handleApiError(error, '提交回复')
  } finally {
    submittingReply.value = false
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN')
}

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

const formatRelativeTime = (dateString: string) => {
  if (!dateString) return '-'
  
  const date = new Date(dateString)
  const now = new Date()
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000)
  
  if (diffInSeconds < 60) return '刚刚'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} 分钟前`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} 小时前`
  if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 86400)} 天前`
  if (diffInSeconds < 31536000) return `${Math.floor(diffInSeconds / 2592000)} 个月前`
  return `${Math.floor(diffInSeconds / 31536000)} 年前`
}

const getRoleText = (role: string) => {
  const roleMap: Record<string, string> = {
    owner: '管理员',
    maintainer: '教师',
    developer: '研究员',
    reporter: '学生',
    guest: '访客'
  }
  return roleMap[role] || role
}

const getHomeworkStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    closed: '已关闭'
  }
  return statusMap[status] || status
}

// 打开GitLab Web IDE
const openWebIDE = async () => {
  if (!project.value?.gitlab_project_id) {
    ElMessage.error('无法获取GitLab项目信息')
    return
  }
  
  try {
    // 获取GitLab配置
    const config: any = await gitlabService.getConfig()
    const gitlabUrl = config.gitlab_url
    
    // 获取项目详细信息以获取项目路径
    const gitlabProject: any = await gitlabService.getProject(project.value.gitlab_project_id.toString())
    const projectPath = gitlabProject.project?.path_with_namespace || `root/${project.value?.name}`
    
    const branch = 'main' // 默认分支，可以根据需要修改
    const path = currentPath.value || ''
    
    // 构建GitLab Web IDE URL
    // 格式: http://gitlab-url/-/ide/project/namespace/project-name/edit/branch/-/path
    const ideUrl = `${gitlabUrl}/-/ide/project/${projectPath}/edit/${branch}/-/${path}`
    
    // 在新窗口中打开Web IDE
    window.open(ideUrl, '_blank')
  } catch (error) {
    console.error('获取GitLab配置失败:', error)
    ElMessage.error('无法打开Web IDE，请检查GitLab配置')
  }
}

// 编辑文件（在Web IDE中打开特定文件）
const editFile = async () => {
  if (!selectedFile.value || !project.value?.gitlab_project_id) {
    ElMessage.error('请先选择要编辑的文件')
    return
  }
  
  try {
    // 获取GitLab配置
    const config: any = await gitlabService.getConfig()
    const gitlabUrl = config.gitlab_url
    
    // 获取项目详细信息以获取项目路径
    const gitlabProject: any = await gitlabService.getProject(project.value.gitlab_project_id.toString())
    const projectPath = gitlabProject.project?.path_with_namespace || `root/${project.value?.name}`
    
    const branch = 'main'
    const filePath = currentPath.value ? `${currentPath.value}/${selectedFile.value.name}` : selectedFile.value.name
    
    const ideUrl = `${gitlabUrl}/-/ide/project/${projectPath}/edit/${branch}/-/${filePath}`
    window.open(ideUrl, '_blank')
  } catch (error) {
    console.error('获取GitLab配置失败:', error)
    ElMessage.error('无法打开文件编辑器')
  }
}

// 下载文件
const downloadFile = async () => {
  if (!selectedFile.value || !project.value?.gitlab_project_id) {
    ElMessage.error('请先选择要下载的文件')
    return
  }
  
  try {
    // 获取GitLab配置
    const config: any = await gitlabService.getConfig()
    const gitlabUrl = config.gitlab_url
    
    // 获取项目详细信息以获取项目路径
    const gitlabProject: any = await gitlabService.getProject(project.value.gitlab_project_id.toString())
    const projectPath = gitlabProject.project?.path_with_namespace || `root/${project.value?.name}`
    
    const branch = 'main'
    const filePath = currentPath.value ? `${currentPath.value}/${selectedFile.value.name}` : selectedFile.value.name
    
    const downloadUrl = `${gitlabUrl}/${projectPath}/-/raw/${branch}/${filePath}`
    window.open(downloadUrl, '_blank')
  } catch (error) {
    console.error('获取GitLab配置失败:', error)
    ElMessage.error('无法下载文件')
  }
}

// 搜索用户（占位方法）
const searchUsers = () => {
  // 这里可以实现用户搜索功能
  console.log('搜索用户:', memberSearchQuery.value)
}

// 显示添加成员对话框（占位方法）
const showAddMemberDialog = () => {
  ElMessage.info('添加成员功能待实现')
}

// 生命周期
onMounted(() => {
  fetchProject()
})

// 监听路由参数变化
watch(() => route.params.id, () => {
  if (route.params.id) {
    fetchProject()
    fetchFiles()
  }
})
</script>

<style scoped>
.scene-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 20px;
}

.scene-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.title-section h1 {
  margin: 0 0 8px 0;
  color: var(--primary-color);
}

.scene-meta {
  display: flex;
  gap: 16px;
  align-items: center;
  font-size: 14px;
  color: var(--light-text);
}

.scene-description {
  margin: 0;
  color: var(--text-color);
  line-height: 1.6;
}

.tab-container {
  min-height: 600px;
}

/* 文件浏览样式 */
.file-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background-color: var(--background-light);
  border-radius: 8px;
}

.file-breadcrumb .breadcrumb-item {
  cursor: pointer;
  color: var(--primary-color);
}

.file-breadcrumb .breadcrumb-item:hover {
  text-decoration: underline;
}

.file-actions {
  display: flex;
  gap: 8px;
}

.file-explorer {
  display: flex;
  gap: 20px;
  min-height: 400px;
}

.file-list {
  flex: 1;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

/* 文件列表表头样式 */
.file-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background-color: var(--background-light);
  border-bottom: 1px solid var(--border-color);
  font-weight: 600;
  font-size: 14px;
  color: var(--text-color);
}

.file-header .file-name-col {
  flex: 2;
}

.file-header .file-commit-col {
  flex: 2;
}

.file-header .file-update-col {
  flex: 1;
  text-align: right;
}

/* 文件列表项样式 */
.file-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background-color 0.3s;
}

.file-item:hover {
  background-color: var(--background-light);
}

.file-item:last-child {
  border-bottom: none;
}

.file-name-col {
  flex: 2;
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-commit-col {
  flex: 2;
  font-size: 14px;
  color: var(--text-color);
}

.file-update-col {
  flex: 1;
  text-align: right;
  font-size: 12px;
  color: var(--light-text);
}

.file-icon {
  font-size: 16px;
  color: var(--primary-color);
}

.file-name {
  font-weight: 500;
}

.commit-message {
  color: var(--text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.update-time {
  color: var(--light-text);
}

.file-preview {
  flex: 1;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background-color: var(--background-light);
  border-bottom: 1px solid var(--border-color);
}

.preview-header h3 {
  margin: 0;
}

.preview-actions {
  display: flex;
  gap: 8px;
}

.preview-content {
  padding: 16px;
  max-height: 400px;
  overflow: auto;
}

.code-preview {
  margin: 0;
  background-color: var(--background-dark);
  color: var(--text-light);
  padding: 16px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.binary-file-info {
  text-align: center;
  padding: 40px;
  color: var(--light-text);
}

.binary-file-info .el-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

/* 话题列表样式 */
.topics-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.topics-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.topics-list {
  min-height: 300px;
}

.topic-item {
  display: flex;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background-color 0.3s;
}

.topic-item:hover {
  background-color: var(--background-light);
}

.topic-item:last-child {
  border-bottom: none;
}

.topic-status {
  width: 80px;
  margin-right: 16px;
}

.topic-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  display: block;
  font-weight: bold;
  color: var(--primary-color);
}

.stat-label {
  font-size: 12px;
  color: var(--light-text);
}

.topic-content {
  flex: 1;
  cursor: pointer;
  transition: background-color 0.2s ease;
  padding: 8px;
  border-radius: 6px;
}

.topic-content:hover {
  background-color: var(--background-light);
}

.topic-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: var(--text-color);
}

.topic-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--light-text);
}

.topic-author {
  display: flex;
  align-items: center;
  gap: 4px;
}

.topic-author img {
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

.topic-tag {
  margin: 0;
}

.topic-excerpt {
  color: var(--light-text);
  font-size: 14px;
}

.topic-excerpt p {
  margin: 0;
  line-height: 1.4;
}

.topic-last-reply {
  width: 120px;
  text-align: right;
  font-size: 12px;
  color: var(--light-text);
}

/* 作业列表样式 */
.homework-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.homework-list {
  min-height: 300px;
}

.homework-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  margin-bottom: 16px;
  background-color: var(--card-background);
}

.homework-info {
  flex: 1;
}

.homework-info:hover {
  background-color: var(--background-light);
  border-radius: 6px;
  padding: 8px;
  margin: -8px;
  transition: all 0.2s ease;
}

.homework-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: var(--text-color);
}

.homework-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--light-text);
}

.homework-description {
  color: var(--light-text);
  font-size: 14px;
}

.homework-description p {
  margin: 0;
  line-height: 1.4;
}

.homework-actions {
  display: flex;
  gap: 8px;
}

/* 成员管理样式 */
.member-management {
  max-height: 400px;
  overflow-y: auto;
}

.member-search {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.member-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.member-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.member-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.member-avatar img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
}

.member-details h4 {
  margin: 0 0 4px 0;
  font-size: 14px;
}

.member-role {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.owner-role {
  background-color: var(--danger-color);
  color: white;
}

.maintainer-role {
  background-color: var(--warning-color);
  color: white;
}

.developer-role {
  background-color: var(--info-color);
  color: white;
}

.reporter-role {
  background-color: var(--success-color);
  color: white;
}

.guest-role {
  background-color: var(--light-text);
  color: white;
}

/* 作业样式 */
.homework-content {
  padding: 20px;
}

.homework-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.homework-header h3 {
  margin: 0;
  color: var(--text-color);
}

.homework-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.homework-item {
  background-color: var(--card-background);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 20px;
  transition: all 0.3s ease;
}

.homework-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: var(--primary-color);
}

.homework-item .homework-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
  padding: 0;
}

.homework-title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.homework-title {
  margin: 0;
  color: var(--primary-color);
  font-size: 18px;
  font-weight: 600;
}

.homework-content {
  margin-top: 16px;
}

.homework-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--light-text);
}

.meta-item .el-icon {
  font-size: 16px;
  color: var(--primary-color);
}

.homework-description {
  margin: 12px 0;
  padding: 12px;
  background-color: var(--background-light);
  border-radius: 6px;
  border-left: 4px solid var(--primary-color);
}

.homework-description p {
  margin: 0;
  color: var(--text-color);
  line-height: 1.6;
}

.homework-progress {
  margin: 16px 0;
  padding: 12px;
  background-color: var(--background-light);
  border-radius: 6px;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--text-color);
}

.my-submission-status {
  margin: 16px 0;
  padding: 12px;
  background-color: var(--success-light, #f0f9ff);
  border: 1px solid var(--success-color, #10b981);
  border-radius: 6px;
}

.submission-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--success-color, #10b981);
  font-weight: 500;
}

.submission-info .grade {
  margin-left: auto;
  font-weight: 600;
  color: var(--primary-color);
}

.homework-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

/* 表单样式 */
.form-tip {
  margin-left: 8px;
  font-size: 12px;
  color: var(--light-text);
}

.danger-item {
  color: var(--danger-color) !important;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .scene-detail {
    padding: 10px;
  }
  
  .header-content {
    flex-direction: column;
    gap: 16px;
  }
  
  .file-explorer {
    flex-direction: column;
  }
  
  .file-preview {
    margin-top: 20px;
  }
  
  .topic-item {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .topic-status {
    width: auto;
    margin-right: 0;
  }
  
  .topic-stats {
    flex-direction: row;
    justify-content: space-around;
  }
  
  .topic-last-reply {
    width: auto;
    text-align: left;
  }
  
  .homework-item {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  
  .homework-actions {
    justify-content: flex-end;
  }
  
  .homework-meta {
    flex-direction: column;
    gap: 8px;
  }
  
  .meta-item {
    font-size: 12px;
  }
  
  .member-item {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
}

/* 删除确认对话框样式 */
.delete-warning {
  text-align: center;
  padding: 20px;
}

.warning-icon {
  margin-bottom: 20px;
}

.warning-text {
  font-size: 18px;
  color: #303133;
  margin-bottom: 20px;
  line-height: 1.5;
}

.warning-text strong {
  color: #F56C6C;
}

.warning-details {
  text-align: left;
  background: #FEF0F0;
  border: 1px solid #FCDCDC;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 20px;
}

.warning-subtitle {
  font-size: 14px;
  font-weight: 600;
  color: #F56C6C;
  margin: 0 0 12px 0;
}

.warning-list {
  margin: 0 0 12px 0;
  padding-left: 20px;
}

.warning-list li {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 4px;
}

.warning-emphasis {
  font-size: 14px;
  color: #E6A23C;
  margin: 0;
  padding: 8px;
  background: #FDF6EC;
  border-radius: 4px;
  border-left: 4px solid #E6A23C;
}

.confirmation-input {
  text-align: left;
  margin-top: 20px;
}

.input-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
}

/* 话题操作区域样式 */
.topic-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
  display: flex;
  gap: 8px;
}

.topic-actions .el-button {
  border-radius: 16px;
}

.comment-count {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #909399;
  font-size: 14px;
}

/* 回复对话框样式 */
.reply-context {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.reply-context h4 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 16px;
}

.context-excerpt {
  color: #606266;
  font-size: 14px;
  line-height: 1.5;
  margin: 0;
}

/* 话题详情模态框样式 */
.topic-detail {
  margin-bottom: 24px;
}

.topic-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.topic-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.author-name {
  font-weight: 600;
  color: #303133;
  font-size: 14px;
}

.publish-time {
  color: #909399;
  font-size: 12px;
}

.topic-labels {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.topic-label {
  border-radius: 12px;
}

.topic-content {
  margin-bottom: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #409eff;
}

.topic-content p {
  margin: 0;
  color: #303133;
  line-height: 1.6;
  font-size: 14px;
}

.topic-stats {
  display: flex;
  gap: 20px;
  padding: 12px 0;
  border-bottom: 1px solid #e4e7ed;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #909399;
  font-size: 14px;
}

.stat-item .el-icon {
  font-size: 16px;
}

.comments-section {
  margin: 24px 0;
}

.comments-title {
  color: #303133;
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid #e4e7ed;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.comment-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.comment-content {
  flex: 1;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.comment-author {
  font-weight: 600;
  color: #303133;
  font-size: 14px;
}

.comment-time {
  color: #909399;
  font-size: 12px;
}

.comment-body {
  color: #606266;
  line-height: 1.6;
  font-size: 14px;
}

.reply-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}

.reply-title {
  color: #303133;
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
}

.reply-input {
  margin-bottom: 12px;
}

.reply-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.login-prompt {
  text-align: center;
  padding: 40px 20px;
  color: #909399;
}

.login-prompt p {
  margin: 0;
  font-size: 14px;
}

/* 回复分页样式 */
.comments-pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}
</style>

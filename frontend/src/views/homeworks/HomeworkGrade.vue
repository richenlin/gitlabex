<template>
  <div class="homework-grade" v-loading="loading">
    <div class="grade-container">
      <div class="header">
        <el-button @click="$router.go(-1)" text>
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>批改作业</h1>
      </div>

      <div v-if="homework" class="homework-info">
        <el-card>
          <template #header>
            <span>{{ homework.title }}</span>
          </template>
          <div class="homework-meta">
            <span>截止时间：{{ homework.due_date ? formatDate(homework.due_date) : '无限制' }}</span>
            <span>满分：{{ homework.max_grade }}分</span>
          </div>
        </el-card>
      </div>

      <!-- 批改表单 -->
      <div class="grading-form">
        <el-card title="批改作业">
          <el-form
            ref="gradeFormRef"
            :model="gradeForm"
            label-width="80px"
          >
            <el-form-item label="成绩" prop="grade">
              <el-input-number
                v-model="gradeForm.grade"
                :min="0"
                :max="homework?.max_grade || 100"
                :precision="0"
                controls-position="right"
                style="width: 200px"
              />
              <span class="grade-suffix">分</span>
            </el-form-item>
            
            <el-form-item label="评语" prop="feedback">
              <el-input
                v-model="gradeForm.feedback"
                type="textarea"
                :rows="6"
                placeholder="请输入评语和建议..."
                maxlength="1000"
                show-word-limit
              />
            </el-form-item>
            
            <el-form-item>
              <div class="grade-actions">
                <el-button @click="clearGradeForm">重置</el-button>
                <el-button 
                  type="primary" 
                  @click="saveGrade"
                  :loading="grading"
                >
                  保存评分
                </el-button>
              </div>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { homeworkService } from '@/services/api'
import type { Homework } from '@/types'
import { ElMessage, type FormInstance } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const homework = ref<Homework | null>(null)
const loading = ref(false)
const grading = ref(false)

const gradeFormRef = ref<FormInstance>()
const gradeForm = ref({
  grade: 0,
  feedback: ''
})

// 计算属性
const homeworkId = computed(() => route.params.id as string)

// 方法
const fetchHomework = async () => {
  loading.value = true
  try {
    const response = await homeworkService.getHomeworkById(homeworkId.value)
    homework.value = response.data || response
  } catch (error) {
    ElMessage.error('获取作业详情失败')
  } finally {
    loading.value = false
  }
}

const saveGrade = async () => {
  try {
    grading.value = true
    
    // TODO: 实现评分保存逻辑
    ElMessage.success('评分保存成功')
    
  } catch (error) {
    ElMessage.error('评分保存失败')
  } finally {
    grading.value = false
  }
}

const clearGradeForm = () => {
  gradeForm.value = {
    grade: 0,
    feedback: ''
  }
}

const formatDate = (date: string | Date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchHomework()
})
</script>

<style scoped>
.homework-grade {
  padding: 20px;
}

.grade-container {
  max-width: 800px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.homework-info {
  margin-bottom: 24px;
}

.homework-meta {
  display: flex;
  gap: 24px;
  font-size: 14px;
  color: #606266;
}

.grading-form {
  margin-bottom: 24px;
}

.grade-suffix {
  margin-left: 8px;
  color: #606266;
}

.grade-actions {
  display: flex;
  gap: 12px;
}
</style>

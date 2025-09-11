<template>
  <div class="homework-submit" v-loading="loading">
    <div class="submit-container">
      <div class="header">
        <el-button @click="$router.go(-1)" text>
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>提交作业</h1>
      </div>

      <div v-if="homework" class="homework-info">
        <el-card>
          <template #header>
            <div class="homework-header">
              <span>{{ homework.title }}</span>
              <el-tag :type="isOverdue ? 'danger' : 'success'">
                {{ isOverdue ? '已逾期' : '进行中' }}
              </el-tag>
            </div>
          </template>
          
          <div class="homework-meta">
            <div class="meta-item">
              <el-icon><Calendar /></el-icon>
              <span>截止时间：{{ formatDate(homework.due_date) }}</span>
            </div>
            <div class="meta-item">
              <el-icon><Star /></el-icon>
              <span>满分：{{ homework.max_grade }}分</span>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 提交表单 -->
      <div class="submit-form">
        <el-card title="作业提交">
          <el-form
            ref="submitFormRef"
            :model="submitForm"
            label-width="120px"
          >
            <el-form-item label="提交内容" prop="content">
              <el-input
                v-model="submitForm.content"
                type="textarea"
                :rows="8"
                placeholder="请输入作业内容..."
                maxlength="5000"
                show-word-limit
              />
            </el-form-item>

            <el-form-item label="提交说明">
              <el-input
                v-model="submitForm.notes"
                type="textarea"
                :rows="3"
                placeholder="可选：添加提交说明..."
                maxlength="500"
                show-word-limit
              />
            </el-form-item>
          </el-form>

          <div class="submit-actions">
            <el-button @click="$router.go(-1)">取消</el-button>
            <el-button 
              type="primary" 
              @click="submitHomework"
              :loading="submitting"
              :disabled="!canSubmitNow"
            >
              提交作业
            </el-button>
          </div>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Calendar, Star } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const homework = ref<Homework | null>(null)
const loading = ref(false)
const submitting = ref(false)

const submitForm = ref({
  content: '',
  notes: ''
})

// 计算属性
const homeworkId = computed(() => route.params.id as string)

const isOverdue = computed(() => {
  if (!homework.value?.due_date) return false
  return new Date() > new Date(homework.value.due_date)
})

const canSubmitNow = computed(() => {
  return !isOverdue.value && !!submitForm.value.content.trim()
})

// 方法
const fetchHomework = async () => {
  loading.value = true
  try {
    const response = await homeworkService.getHomeworkById(homeworkId.value)
    homework.value = response
  } catch (error) {
    ElMessage.error('获取作业详情失败')
    router.go(-1)
  } finally {
    loading.value = false
  }
}

const submitHomework = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要提交作业吗？提交后无法修改。',
      '确认提交',
      { type: 'warning' }
    )
    
    submitting.value = true
    
    const submitData = {
      homework_id: homeworkId.value,
      content: submitForm.value.content,
      notes: submitForm.value.notes
    }
    
    await homeworkService.submitHomework(submitData)
    
    ElMessage.success('作业提交成功')
    router.push(`/homeworks/${homeworkId.value}`)
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('作业提交失败')
    }
  } finally {
    submitting.value = false
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
.homework-submit {
  padding: 20px;
}

.submit-container {
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

.homework-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.homework-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.submit-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}
</style>

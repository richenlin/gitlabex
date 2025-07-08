<template>
  <div class="not-found-view">
    <el-result icon="warning" title="404" sub-title="抱歉，您访问的页面不存在">
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">
          <el-icon><House /></el-icon>
          返回首页
        </el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
// 404页面不需要额外逻辑
</script>

<style scoped>
.not-found-view {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 50vh;
  padding: 40px;
}
</style> 
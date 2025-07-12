<template>
  <div class="assignment-detail">
    <el-page-header @back="$router.go(-1)" content="作业详情" />
    <el-card style="margin-top: 20px;">
      <h2>作业详情页面</h2>
      <p>作业ID: {{ $route.params.id }}</p>
      <p>此页面正在开发中...</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
// 作业详情页面占位符
</script>

<style scoped>
.assignment-detail {
  padding: 20px;
}
</style> 
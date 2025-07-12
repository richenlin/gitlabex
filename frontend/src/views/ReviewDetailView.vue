<template>
  <div class="review-detail">
    <el-page-header @back="$router.go(-1)" content="评审详情" />
    <el-card style="margin-top: 20px;">
      <h2>评审详情页面</h2>
      <p>评审ID: {{ $route.params.id }}</p>
      <p>此页面正在开发中...</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
// 评审详情页面占位符
</script>

<style scoped>
.review-detail {
  padding: 20px;
}
</style> 
<template>
  <div class="login-container">
    <div class="login-card">
      <h1>GitLabEx 教学平台</h1>
      <p>使用 GitLab 账号登录</p>
      <button @click="loginWithGitLab" class="login-btn">
        使用 GitLab 登录
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const loginWithGitLab = () => {
  // 直接跳转到后端的认证端点
  window.location.href = '/api/auth/gitlab';
};
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f5f5;
}

.login-card {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  text-align: center;
  max-width: 400px;
  width: 100%;
}

.login-card h1 {
  color: #333;
  margin-bottom: 0.5rem;
}

.login-card p {
  color: #666;
  margin-bottom: 2rem;
}

.login-btn {
  background-color: #fc6d26;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  width: 100%;
  transition: background-color 0.2s;
}

.login-btn:hover {
  background-color: #e24329;
}

.login-btn:active {
  background-color: #d12f1b;
}
</style> 
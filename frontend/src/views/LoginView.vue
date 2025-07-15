<template>
  <div class="login-container">
    <div class="container">
      <div class="header">
        <h1>🚀 GitLabEx</h1>
        <p>基于 GitLab + OnlyOffice 的现代化教育协作平台</p>
      </div>

      <div class="features">
        <div class="feature-card">
          <h3>📚 教育优化</h3>
          <p>专为教育场景设计，完美适配课程管理、作业提交、团队协作等教学需求</p>
        </div>
        <div class="feature-card">
          <h3>📝 在线编辑</h3>
          <p>集成OnlyOffice，支持Word、Excel、PowerPoint文档的实时协作编辑</p>
        </div>
        <div class="feature-card">
          <h3>🔗 GitLab集成</h3>
          <p>100%复用GitLab的用户管理、权限控制、项目协作能力</p>
        </div>
      </div>

      <div class="login-section">
        <h2>🔐 登录系统</h2>
        <p>使用您的 GitLab 账号登录以访问教育协作平台</p>
        <div class="login-card">
          <button @click="loginWithGitLab" class="login-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" style="margin-right: 8px;">
              <path d="M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 0 1-.3-.94l1.22-3.78 2.44-7.51A.42.42 0 0 1 4.82 2a.43.43 0 0 1 .58 0 .42.42 0 0 1 .11.18l2.44 7.49h8.1l2.44-7.51A.42.42 0 0 1 18.6 2a.43.43 0 0 1 .58 0 .42.42 0 0 1 .11.18l2.44 7.51L23 13.45a.84.84 0 0 1-.35.94z"/>
            </svg>
            使用 GitLab 登录
          </button>
        </div>
      </div>

      <div class="tech-stack">
        <div class="tech-item">Go + Gin</div>
        <div class="tech-item">Vue 3 + TypeScript</div>
        <div class="tech-item">PostgreSQL</div>
        <div class="tech-item">Redis</div>
        <div class="tech-item">OnlyOffice</div>
        <div class="tech-item">GitLab CE</div>
        <div class="tech-item">Docker</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const loginWithGitLab = () => {
  // 使用环境变量配置的API基础URL
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const authUrl = `${apiBaseUrl}/api/auth/gitlab`;
  
  console.log('Redirecting to:', authUrl);
  
  // 直接跳转到后端的认证端点
  window.location.href = authUrl;
};
</script>

<style scoped>
* { 
  margin: 0; 
  padding: 0; 
  box-sizing: border-box; 
}

.login-container {
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
  color: #333;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.header {
  text-align: center;
  color: white;
  margin-bottom: 3rem;
}

.header h1 {
  font-size: 3rem;
  margin-bottom: 1rem;
  text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
}

.header p {
  font-size: 1.2rem;
  opacity: 0.9;
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin-bottom: 3rem;
}

.feature-card {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  text-align: center;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-card h3 {
  color: #5a67d8;
  margin-bottom: 1rem;
  font-size: 1.5rem;
}

.feature-card p {
  color: #666;
  line-height: 1.6;
}

.login-section {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  margin-bottom: 2rem;
  text-align: center;
}

.login-section h2 {
  color: #5a67d8;
  margin-bottom: 1rem;
}

.login-section p {
  color: #666;
  margin-bottom: 2rem;
  line-height: 1.6;
}

.login-card {
  max-width: 400px;
  margin: 0 auto;
}

.login-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 16px 32px;
  background: #fc6d26;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(252, 109, 38, 0.3);
}

.login-btn:hover {
  background: #e24329;
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(252, 109, 38, 0.4);
}

.login-btn:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(252, 109, 38, 0.3);
}

.tech-stack {
  display: flex;
  justify-content: center;
  gap: 2rem;
  margin-top: 2rem;
  flex-wrap: wrap;
}

.tech-item {
  background: rgba(255,255,255,0.2);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
  backdrop-filter: blur(10px);
}

@media (max-width: 768px) {
  .container {
    padding: 1rem;
  }
  
  .header h1 {
    font-size: 2rem;
  }
  
  .features {
    grid-template-columns: 1fr;
  }
  
  .tech-stack {
    gap: 1rem;
  }
}
</style> 
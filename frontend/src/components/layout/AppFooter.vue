<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="footer-section">
        <h3>协同创新社区</h3>
        <p>基于 GitLab 的教育协作平台</p>
      </div>
      
      <div class="footer-section">
        <h4>快速链接</h4>
        <ul>
          <li><router-link to="/">首页</router-link></li>
          <li><router-link to="/scenes">课题</router-link></li>
          <li><router-link to="/topics">话题</router-link></li>
          <li><router-link to="/documents">文档</router-link></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h4>帮助与支持</h4>
        <ul>
          <li><a href="#" @click.prevent="showHelp">使用指南</a></li>
          <li><a href="#" @click.prevent="showFeedback">反馈建议</a></li>
          <li><a href="#" @click.prevent="showAbout">关于我们</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h4>联系我们</h4>
        <p>邮箱: support@gitlabex.com</p>
        <p>技术支持: 7x24小时在线</p>
      </div>
    </div>
    
    <div class="footer-bottom">
      <p>&copy; 2024 协同创新社区. 保留所有权利.</p>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'

const showHelp = () => {
  ElMessage.info('使用指南功能开发中...')
}

const showFeedback = () => {
  ElMessage.info('反馈建议功能开发中...')
}

const showAbout = () => {
  ElMessage.info('关于我们功能开发中...')
}
</script>

<style scoped>
.app-footer {
  background-color: var(--card-background);
  border-top: 1px solid var(--border-color);
  padding: 40px 0 20px;
  margin-top: auto;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 30px;
}

.footer-section h3,
.footer-section h4 {
  color: var(--primary-color);
  margin-bottom: 16px;
}

.footer-section p {
  color: var(--light-text);
  margin-bottom: 8px;
  line-height: 1.6;
}

.footer-section ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.footer-section ul li {
  margin-bottom: 8px;
}

.footer-section ul li a {
  color: var(--light-text);
  transition: color 0.3s;
}

.footer-section ul li a:hover {
  color: var(--accent-color);
}

.footer-bottom {
  text-align: center;
  padding-top: 20px;
  margin-top: 20px;
  border-top: 1px solid var(--border-color);
  color: var(--lighter-text);
  font-size: 14px;
}

@media (max-width: 768px) {
  .footer-content {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  
  .app-footer {
    padding: 20px 0 10px;
  }
  
  .footer-content {
    padding: 0 16px;
  }
}
</style>
// 主要JavaScript文件
document.addEventListener('DOMContentLoaded', function() {
  console.log('页面已加载');
  
  // 弹窗功能
  const modal = document.getElementById('announcement-modal');
  const closeBtn = document.querySelector('.close-modal');
  const announcementBtn = document.querySelector('.announcement-icon');
  
  if (announcementBtn && modal) {
    announcementBtn.addEventListener('click', function() {
      modal.style.display = 'flex';
      document.body.style.overflow = 'hidden';
    });
    
    closeBtn.addEventListener('click', function() {
      modal.style.display = 'none';
      document.body.style.overflow = '';
    });
    
    window.addEventListener('click', function(event) {
      if (event.target === modal) {
        modal.style.display = 'none';
        document.body.style.overflow = '';
      }
    });
  }
  
  // 初始化所有功能模块
  if (typeof initEmojiPicker === 'function') {
    initEmojiPicker();
  }
  
  if (typeof initFileUpload === 'function') {
    initFileUpload();
  }
  
  if (typeof initTabs === 'function') {
    initTabs();
  }
  
  if (typeof initComments === 'function') {
    initComments();
  }
  
  if (typeof initSceneManagement === 'function') {
    initSceneManagement();
    // 加载模拟数据（如果函数存在）
    if (typeof loadMockData === 'function') {
      loadMockData();
    }
  }
  
  if (typeof initHomeworkSystem === 'function') {
    initHomeworkSystem();
  }
  
  // 显示欢迎公告（如果函数存在）
  if (typeof showAnnouncement === 'function') {
    showAnnouncement('欢迎', '欢迎使用协同创新社区！');
  }
});
// 标签页切换功能
function initTabs() {
  const tabs = document.querySelectorAll('[data-tab]');
  
  tabs.forEach(tab => {
    tab.addEventListener('click', function() {
      // 移除所有标签的活动状态
      tabs.forEach(t => t.classList.remove('active'));
      
      // 添加当前标签的活动状态
      this.classList.add('active');
      
      // 获取目标内容ID
      const targetId = this.getAttribute('data-tab');
      
      // 隐藏所有内容
      const tabContents = document.querySelectorAll('.tab-pane');
      tabContents.forEach(content => {
        content.classList.remove('active');
      });
      
      // 显示目标内容
      const targetContent = document.getElementById(targetId);
      if (targetContent) {
        targetContent.classList.add('active');
      }
    });
  });
  
  // 默认显示第一个标签页
  const firstTab = document.querySelector('[data-tab]');
  if (firstTab) {
    firstTab.click();
  }
}

// 初始化标签页
document.addEventListener('DOMContentLoaded', function() {
  initTabs();
});
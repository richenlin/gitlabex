// 作业功能
function initHomeworkSystem() {
  // 提交作业按钮
  const submitHomeworkButtons = document.querySelectorAll('.submit-homework-button');
  submitHomeworkButtons.forEach(button => {
    button.addEventListener('click', function() {
      // 在实际应用中，这将打开一个模态框或导航到提交页面
      alert('提交作业功能将在这里实现');
    });
  });
  
  // 批阅作业功能
  const gradeButtons = document.querySelectorAll('.grade-homework-button');
  gradeButtons.forEach(button => {
    button.addEventListener('click', function() {
      const homeworkId = this.getAttribute('data-homework-id');
      // 在实际应用中，这将打开一个模态框或导航到批阅页面
      alert(`批阅作业 ID: ${homeworkId}`);
    });
  });
  
  // 作业列表交互
  const homeworkItems = document.querySelectorAll('.homework-item');
  homeworkItems.forEach(item => {
    item.addEventListener('click', function(e) {
      // 防止点击按钮时触发
      if (e.target.tagName !== 'BUTTON') {
        const homeworkTitle = this.querySelector('.homework-title').textContent;
        // 在实际应用中，这将导航到作业详情页面
        alert(`查看作业: ${homeworkTitle}`);
      }
    });
  });
}
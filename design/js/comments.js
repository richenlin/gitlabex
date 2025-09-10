// 评论功能
function initComments() {
  const commentForms = document.querySelectorAll('.comment-form');
  
  commentForms.forEach(form => {
    form.addEventListener('submit', function(e) {
      e.preventDefault();
      
      const textarea = this.querySelector('textarea');
      const commentText = textarea.value.trim();
      
      if (commentText) {
        // 创建新评论
        const commentsContainer = document.querySelector('.comments-list');
        const newComment = document.createElement('div');
        newComment.className = 'comment';
        
        // 获取当前用户信息（在实际应用中，这将来自后端）
        const currentUser = {
          name: '当前用户',
          avatar: 'https://via.placeholder.com/32'
        };
        
        // 格式化当前日期
        const now = new Date();
        const formattedDate = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
        
        newComment.innerHTML = `
          <div class="comment-header">
            <div class="comment-author">
              <div class="avatar">
                <img src="${currentUser.avatar}" alt="${currentUser.name}">
              </div>
              <span class="comment-author-name">${currentUser.name}</span>
            </div>
            <div class="comment-date">${formattedDate}</div>
          </div>
          <div class="comment-body">
            ${commentText}
          </div>
          <div class="comment-actions">
            <span class="comment-action">👍 点赞</span>
            <span class="comment-action">💬 回复</span>
          </div>
        `;
        
        // 添加到评论列表
        if (commentsContainer) {
          commentsContainer.appendChild(newComment);
        }
        
        // 清空文本区域
        textarea.value = '';
        
        // 处理文件上传（在实际应用中，这将涉及到文件上传API）
        const filePreview = this.querySelector('.file-preview');
        if (filePreview) {
          filePreview.innerHTML = '';
        }
      }
    });
  });
}
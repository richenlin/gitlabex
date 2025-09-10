document.addEventListener('DOMContentLoaded', function() {
    // 点赞功能
    const likeButtons = document.querySelectorAll('.like-btn');
    likeButtons.forEach(button => {
        button.addEventListener('click', function() {
            const countElement = this.querySelector('.action-count');
            let count = parseInt(countElement.textContent);
            
            // 检查是否已点赞（这里简单模拟，实际应用中应该有服务器状态）
            if (!this.classList.contains('liked')) {
                count++;
                this.classList.add('liked');
                this.style.color = '#e74c3c'; // 点赞后变红
            } else {
                count--;
                this.classList.remove('liked');
                this.style.color = ''; // 恢复默认颜色
            }
            
            countElement.textContent = count;
        });
    });
    
    // 回复按钮功能
    const replyButtons = document.querySelectorAll('.reply-btn');
    replyButtons.forEach(button => {
        button.addEventListener('click', function() {
            const replyForm = document.querySelector('.reply-form');
            const replyTextarea = replyForm.querySelector('textarea');
            
            // 获取要回复的用户名
            const authorName = this.closest('.reply-item').querySelector('.author-name').textContent;
            
            // 在文本框中添加@用户名
            replyTextarea.value = `@${authorName} `;
            
            // 滚动到回复框并聚焦
            replyForm.scrollIntoView({ behavior: 'smooth' });
            replyTextarea.focus();
        });
    });
    
    // 收藏功能
    const bookmarkButton = document.querySelector('.bookmark-btn');
    if (bookmarkButton) {
        bookmarkButton.addEventListener('click', function() {
            if (!this.classList.contains('bookmarked')) {
                this.classList.add('bookmarked');
                this.querySelector('.action-text').textContent = '已收藏';
                this.style.color = '#f39c12'; // 收藏后变黄
            } else {
                this.classList.remove('bookmarked');
                this.querySelector('.action-text').textContent = '收藏';
                this.style.color = ''; // 恢复默认颜色
            }
        });
    }
    
    // 分享功能
    const shareButton = document.querySelector('.share-btn');
    if (shareButton) {
        shareButton.addEventListener('click', function() {
            // 获取当前页面URL
            const url = window.location.href;
            
            // 创建临时输入框复制链接
            const tempInput = document.createElement('input');
            document.body.appendChild(tempInput);
            tempInput.value = url;
            tempInput.select();
            document.execCommand('copy');
            document.body.removeChild(tempInput);
            
            // 显示提示
            alert('链接已复制到剪贴板');
        });
    }
    
    // 回复表单提交
    const replyForm = document.querySelector('.reply-form');
    if (replyForm) {
        replyForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const replyContent = this.querySelector('textarea').value;
            if (replyContent.trim() === '') {
                alert('回复内容不能为空');
                return;
            }
            
            // 在实际应用中，这里应该发送数据到服务器
            alert('回复已提交！在实际应用中，这将保存到数据库并显示在页面上。');
            
            // 清空表单
            this.querySelector('textarea').value = '';
        });
    }
    
    // 编辑器工具按钮
    const toolButtons = document.querySelectorAll('.tool-btn');
    toolButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            
            const textarea = document.querySelector('.reply-form textarea');
            const title = this.getAttribute('title');
            
            // 根据不同的工具按钮插入不同的内容
            switch (title) {
                case '添加图片':
                    textarea.value += '\n![图片描述](图片URL)\n';
                    break;
                case '添加代码':
                    textarea.value += '\n```\n// 在这里输入代码\n```\n';
                    break;
                case '添加附件':
                    alert('在实际应用中，这里会打开文件选择对话框');
                    break;
                case '添加公式':
                    textarea.value += '\n$$E = mc^2$$\n';
                    break;
            }
            
            textarea.focus();
        });
    });
});
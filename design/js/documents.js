// 文档模块 - 太空主题

// 文档上传配置
const documentUploadConfig = {
  maxFileSize: 20 * 1024 * 1024, // 20MB
  allowedTypes: [
    'application/pdf', // PDF
    'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', // Word
    'application/vnd.ms-excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', // Excel
    'application/vnd.ms-powerpoint', 'application/vnd.openxmlformats-officedocument.presentationml.presentation', // PowerPoint
    'text/plain', // 文本文件
    'image/jpeg', 'image/png', 'image/gif', 'image/webp' // 图片
  ],
  uploadEndpoint: '/api/documents/upload' // 文档上传API端点
};

// 文档预览配置
const documentPreviewConfig = {
  pdfViewerUrl: '/lib/pdfjs/web/viewer.html?file=', // PDF.js查看器URL
  officeViewerUrl: 'https://view.officeapps.live.com/op/embed.aspx?src=' // Office在线查看器URL
};

// 初始化文档模块
document.addEventListener('DOMContentLoaded', function() {
  // 初始化文档上传功能
  initDocumentUpload();
  
  // 初始化文档预览功能
  initDocumentPreview();
  
  // 初始化文档下载功能
  initDocumentDownload();
  
  // 初始化文档分享功能
  initDocumentShare();
  
  // 初始化文档评论功能
  initDocumentComments();
  
  // 加载文档详情（如果当前是文档详情页）
  if (document.querySelector('.document-detail')) {
    loadDocumentDetail();
  }
});

/**
 * 初始化文档上传功能
 */
function initDocumentUpload() {
  const uploadButton = document.getElementById('upload-document-button');
  const uploadModal = document.getElementById('upload-modal');
  const closeModal = uploadModal.querySelector('.close-modal');
  const uploadForm = document.getElementById('document-upload-form');
  
  if (!uploadButton || !uploadModal) return;
  
  // 打开上传模态框
  uploadButton.addEventListener('click', function() {
    uploadModal.style.display = 'block';
  });
  
  // 关闭上传模态框
  closeModal.addEventListener('click', function() {
    uploadModal.style.display = 'none';
  });
  
  // 点击模态框外部关闭
  window.addEventListener('click', function(event) {
    if (event.target === uploadModal) {
      uploadModal.style.display = 'none';
    }
  });
  
  // 处理表单提交
  uploadForm.addEventListener('submit', function(e) {
    e.preventDefault();
    
    const title = document.getElementById('document-title').value;
    const description = document.getElementById('document-description').value;
    const fileInput = document.getElementById('document-file');
    const file = fileInput.files[0];
    
    // 验证文件
    if (!file) {
      showNotification('错误', '请选择要上传的文件', 'error');
      return;
    }
    
    if (!validateDocumentType(file)) {
      showNotification('错误', `不支持的文件类型: ${file.name}`, 'error');
      return;
    }
    
    if (!validateDocumentSize(file)) {
      showNotification('错误', `文件过大: ${file.name}. 最大允许 ${formatFileSize(documentUploadConfig.maxFileSize)}`, 'error');
      return;
    }
    
    // 创建FormData
    const formData = new FormData();
    formData.append('title', title);
    formData.append('description', description);
    formData.append('file', file);
    
    // 获取选中的分类
    const categories = [];
    uploadForm.querySelectorAll('input[name="document-category"]:checked').forEach(checkbox => {
      categories.push(checkbox.value);
    });
    formData.append('categories', JSON.stringify(categories));
    
    // 显示上传进度
    const progressBar = document.createElement('div');
    progressBar.className = 'upload-progress';
    progressBar.innerHTML = '<div class="progress-bar"></div>';
    uploadForm.appendChild(progressBar);
    
    // 禁用表单
    uploadForm.querySelectorAll('input, textarea, button').forEach(element => {
      element.disabled = true;
    });
    
    // 发送上传请求
    const xhr = new XMLHttpRequest();
    xhr.open('POST', documentUploadConfig.uploadEndpoint, true);
    
    // 上传进度
    xhr.upload.addEventListener('progress', function(e) {
      if (e.lengthComputable) {
        const percentComplete = (e.loaded / e.total) * 100;
        progressBar.querySelector('.progress-bar').style.width = percentComplete + '%';
      }
    });
    
    // 上传完成
    xhr.addEventListener('load', function() {
      if (xhr.status >= 200 && xhr.status < 300) {
        const response = JSON.parse(xhr.responseText);
        showNotification('成功', '文档上传成功', 'success');
        
        // 重置表单
        uploadForm.reset();
        uploadModal.style.display = 'none';
        
        // 如果是文档列表页，刷新列表
        if (document.querySelector('.document-list')) {
          setTimeout(() => {
            window.location.reload();
          }, 1500);
        }
      } else {
        showNotification('错误', `文档上传失败: ${xhr.statusText}`, 'error');
      }
      
      // 启用表单
      uploadForm.querySelectorAll('input, textarea, button').forEach(element => {
        element.disabled = false;
      });
      
      // 移除进度条
      if (progressBar.parentNode === uploadForm) {
        uploadForm.removeChild(progressBar);
      }
    });
    
    // 上传错误
    xhr.addEventListener('error', function() {
      showNotification('错误', '文档上传失败: 网络错误', 'error');
      
      // 启用表单
      uploadForm.querySelectorAll('input, textarea, button').forEach(element => {
        element.disabled = false;
      });
      
      // 移除进度条
      if (progressBar.parentNode === uploadForm) {
        uploadForm.removeChild(progressBar);
      }
    });
    
    // 发送请求
    xhr.send(formData);
  });
}

/**
 * 验证文档类型
 * @param {File} file 文件对象
 * @returns {boolean} 是否为允许的文档类型
 */
function validateDocumentType(file) {
  return documentUploadConfig.allowedTypes.includes(file.type) || documentUploadConfig.allowedTypes.length === 0;
}

/**
 * 验证文档大小
 * @param {File} file 文件对象
 * @returns {boolean} 是否在允许的大小范围内
 */
function validateDocumentSize(file) {
  return file.size <= documentUploadConfig.maxFileSize;
}

/**
 * 格式化文件大小
 * @param {number} bytes 字节数
 * @returns {string} 格式化后的文件大小
 */
function formatFileSize(bytes) {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * 初始化文档预览功能
 */
function initDocumentPreview() {
  const previewButtons = document.querySelectorAll('.preview-button');
  
  previewButtons.forEach(button => {
    button.addEventListener('click', function() {
      const previewContainer = this.closest('.preview-container');
      const previewContent = previewContainer.querySelector('.preview-content');
      
      // 获取当前文档的URL（这里需要根据实际项目调整）
      const documentUrl = '/documents/' + getDocumentId() + '/file';
      
      // 根据文档类型显示预览
      const documentType = getDocumentType();
      
      if (documentType === 'pdf') {
        // 使用PDF.js预览PDF
        previewContent.innerHTML = `
          <iframe src="${documentPreviewConfig.pdfViewerUrl + encodeURIComponent(documentUrl)}" 
                  style="width: 100%; height: 500px; border: none;"></iframe>
        `;
      } else if (documentType === 'word' || documentType === 'excel' || documentType === 'powerpoint') {
        // 使用Office在线查看器预览Office文档
        previewContent.innerHTML = `
          <iframe src="${documentPreviewConfig.officeViewerUrl + encodeURIComponent(documentUrl)}" 
                  style="width: 100%; height: 500px; border: none;" 
                  frameborder="0"></iframe>
        `;
      } else if (documentType === 'image') {
        // 直接显示图片
        previewContent.innerHTML = `
          <img src="${documentUrl}" style="max-width: 100%; max-height: 500px;">
        `;
      } else {
        // 不支持预览的文档类型
        previewContent.innerHTML = `
          <div class="preview-not-supported">
            <p>不支持在线预览此类型的文档</p>
            <p>请下载后查看</p>
          </div>
        `;
      }
      
      // 全屏预览
      if (this.textContent.includes('全屏')) {
        previewContainer.requestFullscreen().catch(err => {
          showNotification('提示', `无法进入全屏模式: ${err.message}`, 'info');
        });
      }
    });
  });
}

/**
 * 获取当前文档ID
 * @returns {string} 文档ID
 */
function getDocumentId() {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('id') || '1'; // 默认为1，实际项目中应该从URL获取
}

/**
 * 获取当前文档类型
 * @returns {string} 文档类型 (pdf, word, excel, powerpoint, image, text, other)
 */
function getDocumentType() {
  const documentUrl = window.location.pathname;
  
  if (documentUrl.endsWith('.pdf')) {
    return 'pdf';
  } else if (documentUrl.endsWith('.doc') || documentUrl.endsWith('.docx')) {
    return 'word';
  } else if (documentUrl.endsWith('.xls') || documentUrl.endsWith('.xlsx')) {
    return 'excel';
  } else if (documentUrl.endsWith('.ppt') || documentUrl.endsWith('.pptx')) {
    return 'powerpoint';
  } else if (documentUrl.match(/\.(jpeg|jpg|png|gif|webp)$/i)) {
    return 'image';
  } else if (documentUrl.endsWith('.txt')) {
    return 'text';
  } else {
    return 'other';
  }
}

/**
 * 初始化文档下载功能
 */
function initDocumentDownload() {
  const downloadButtons = document.querySelectorAll('.action-button.primary-button');
  
  downloadButtons.forEach(button => {
    if (button.textContent.includes('下载')) {
      button.addEventListener('click', function() {
        const documentId = getDocumentId();
        const documentUrl = `/api/documents/${documentId}/download`;
        
        // 创建隐藏的下载链接
        const a = document.createElement('a');
        a.href = documentUrl;
        a.download = '';
        a.style.display = 'none';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        
        // 记录下载次数（需要后端支持）
        recordDocumentDownload(documentId);
      });
    }
  });
}

/**
 * 记录文档下载次数
 * @param {string} documentId 文档ID
 */
function recordDocumentDownload(documentId) {
  fetch(`/api/documents/${documentId}/record-download`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    }
  }).then(response => {
    if (!response.ok) {
      console.error('记录下载次数失败');
    }
  }).catch(error => {
    console.error('记录下载次数出错:', error);
  });
}

/**
 * 初始化文档分享功能
 */
function initDocumentShare() {
  const shareButtons = document.querySelectorAll('.action-button:not(.primary-button)');
  
  shareButtons.forEach(button => {
    if (button.textContent.includes('分享')) {
      button.addEventListener('click', function() {
        const documentId = getDocumentId();
        const shareUrl = `${window.location.origin}/document-detail.html?id=${documentId}`;
        
        // 创建分享模态框
        const shareModal = document.createElement('div');
        shareModal.className = 'share-modal';
        shareModal.style.position = 'fixed';
        shareModal.style.top = '0';
        shareModal.style.left = '0';
        shareModal.style.width = '100%';
        shareModal.style.height = '100%';
        shareModal.style.backgroundColor = 'rgba(0, 0, 0, 0.8)';
        shareModal.style.zIndex = '1000';
        shareModal.style.display = 'flex';
        shareModal.style.justifyContent = 'center';
        shareModal.style.alignItems = 'center';
        
        shareModal.innerHTML = `
          <div class="share-content" style="background-color: #1e293b; padding: 30px; border-radius: 8px; max-width: 500px; width: 90%;">
            <h3 style="margin-top: 0; color: #7dd3fc;">分享文档</h3>
            <div class="share-url" style="margin-bottom: 20px;">
              <label style="display: block; margin-bottom: 8px; color: #e2e8f0;">分享链接</label>
              <div style="display: flex;">
                <input type="text" value="${shareUrl}" readonly style="flex-grow: 1; padding: 8px; background-color: #1e293b; border: 1px solid #334155; color: #e2e8f0; border-radius: 4px 0 0 4px;">
                <button class="copy-url-button" style="padding: 8px 15px; background-color: #3b82f6; color: white; border: none; border-radius: 0 4px 4px 0; cursor: pointer;">复制</button>
              </div>
            </div>
            <div class="share-platforms" style="margin-bottom: 20px;">
              <label style="display: block; margin-bottom: 8px; color: #e2e8f0;">分享到</label>
              <div style="display: flex; gap: 10px;">
                <button class="share-wechat" style="padding: 8px 15px; background-color: #1e293b; border: 1px solid #334155; color: #e2e8f0; border-radius: 4px; cursor: pointer;">微信</button>
                <button class="share-qq" style="padding: 8px 15px; background-color: #1e293b; border: 1px solid #334155; color: #e2e8f0; border-radius: 4px; cursor: pointer;">QQ</button>
                <button class="share-weibo" style="padding: 8px 15px; background-color: #1e293b; border: 1px solid #334155; color: #e2e8f0; border-radius: 4px; cursor: pointer;">微博</button>
              </div>
            </div>
            <div class="share-actions" style="display: flex; justify-content: flex-end;">
              <button class="close-share-modal" style="padding: 8px 15px; background-color: #1e293b; border: 1px solid #334155; color: #e2e8f0; border-radius: 4px; cursor: pointer;">关闭</button>
            </div>
          </div>
        `;
        
        // 复制链接功能
        shareModal.querySelector('.copy-url-button').addEventListener('click', function() {
          const urlInput = shareModal.querySelector('input[type="text"]');
          urlInput.select();
          document.execCommand('copy');
          showNotification('提示', '链接已复制到剪贴板', 'info');
        });
        
        // 关闭模态框
        shareModal.querySelector('.close-share-modal').addEventListener('click', function() {
          document.body.removeChild(shareModal);
        });
        
        // 添加到文档
        document.body.appendChild(shareModal);
      });
    }
  });
}

/**
 * 初始化文档评论功能
 */
function initDocumentComments() {
  const commentForm = document.querySelector('.comment-form');
  if (!commentForm) return;
  
  const textarea = commentForm.querySelector('textarea');
  const submitButton = commentForm.querySelector('.submit-button');
  
  // 表情选择器
  const emojiPickerTrigger = commentForm.querySelector('.emoji-picker-trigger');
  if (emojiPickerTrigger) {
    emojiPickerTrigger.addEventListener('click', function() {
      showEmojiPicker(textarea);
    });
  }
  
  // 提交评论
  commentForm.addEventListener('submit', function(e) {
    e.preventDefault();
    
    const commentText = textarea.value.trim();
    if (!commentText) {
      showNotification('错误', '评论内容不能为空', 'error');
      return;
    }
    
    const documentId = getDocumentId();
    
    // 禁用表单
    textarea.disabled = true;
    submitButton.disabled = true;
    submitButton.textContent = '提交中...';
    
    // 发送评论请求
    fetch(`/api/documents/${documentId}/comments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        content: commentText
      })
    }).then(response => {
      if (response.ok) {
        return response.json();
      }
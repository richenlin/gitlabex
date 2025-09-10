// 文件上传配置
const fileUploadConfig = {
  maxFileSize: 10 * 1024 * 1024, // 10MB
  allowedTypes: [
    'image/jpeg', 'image/png', 'image/gif', 'image/webp', // 图片
    'application/pdf', // PDF
    'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', // Word
    'application/vnd.ms-excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', // Excel
    'application/vnd.ms-powerpoint', 'application/vnd.openxmlformats-officedocument.presentationml.presentation', // PowerPoint
    'text/plain', // 文本文件
    'video/mp4', 'video/webm', // 视频
    'audio/mpeg', 'audio/wav', 'audio/ogg' // 音频
  ],
  uploadEndpoint: '/api/upload' // 上传API端点
};

// 存储已选择的文件
const selectedFiles = new Map();

// 文件上传功能
function initFileUpload() {
  const fileInputs = document.querySelectorAll('.file-upload-input');
  
  fileInputs.forEach(input => {
    // 为每个文件输入创建一个唯一ID
    const inputId = `file-input-${Math.random().toString(36).substr(2, 9)}`;
    input.dataset.inputId = inputId;
    selectedFiles.set(inputId, new Map());
    
    // 添加上传按钮
    const uploadButton = document.createElement('button');
    uploadButton.className = 'upload-button';
    uploadButton.textContent = '上传文件';
    uploadButton.style.display = 'none'; // 初始隐藏
    uploadButton.addEventListener('click', function(e) {
      e.preventDefault();
      uploadFiles(inputId);
    });
    
    const fileUploadContainer = input.closest('.file-upload');
    fileUploadContainer.appendChild(uploadButton);
    
    // 添加拖放功能
    setupDragAndDrop(fileUploadContainer, inputId);
    
    input.addEventListener('change', function(e) {
      const files = e.target.files;
      const filePreview = this.closest('.file-upload').querySelector('.file-preview');
      const inputId = this.dataset.inputId;
      
      // 处理选择的文件
      handleSelectedFiles(files, filePreview, inputId);
      
      // 重置input，以便可以再次选择相同的文件
      this.value = '';
    });
      
  });
}

/**
 * 设置拖放功能
 * @param {HTMLElement} container 文件上传容器
 * @param {string} inputId 输入框ID
 */
function setupDragAndDrop(container, inputId) {
  const filePreview = container.querySelector('.file-preview');
  
  // 添加拖放区域提示
  const dropZone = document.createElement('div');
  dropZone.className = 'drop-zone';
  dropZone.innerHTML = '<p>拖放文件到这里</p>';
  dropZone.style.display = 'none';
  dropZone.style.border = '2px dashed #ccc';
  dropZone.style.borderRadius = '4px';
  dropZone.style.padding = '20px';
  dropZone.style.textAlign = 'center';
  dropZone.style.color = '#666';
  dropZone.style.marginTop = '10px';
  
  container.appendChild(dropZone);
  
  // 拖放事件
  container.addEventListener('dragover', function(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.style.display = 'block';
    dropZone.style.borderColor = '#0366d6';
    dropZone.style.backgroundColor = 'rgba(3, 102, 214, 0.05)';
  });
  
  container.addEventListener('dragleave', function(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.style.borderColor = '#ccc';
    dropZone.style.backgroundColor = 'transparent';
  });
  
  container.addEventListener('drop', function(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.style.display = 'none';
    
    const files = e.dataTransfer.files;
    handleSelectedFiles(files, filePreview, inputId);
  });
}

/**
 * 处理选择的文件
 * @param {FileList} files 文件列表
 * @param {HTMLElement} filePreview 文件预览区域
 * @param {string} inputId 输入框ID
 */
function handleSelectedFiles(files, filePreview, inputId) {
  // 获取已选择的文件Map
  const filesMap = selectedFiles.get(inputId);
  
  // 显示上传按钮
  const uploadButton = filePreview.closest('.file-upload').querySelector('.upload-button');
  uploadButton.style.display = 'block';
  
  // 如果是第一次添加文件，清空预览区域
  if (filesMap.size === 0 && filePreview) {
    filePreview.innerHTML = '';
  }
  
  // 显示选择的文件
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fileId = `file-${Math.random().toString(36).substr(2, 9)}`;
    
    // 验证文件类型
    if (!validateFileType(file)) {
      showNotification('错误', `不支持的文件类型: ${file.name}`, 'error');
      continue;
    }
    
    // 验证文件大小
    if (!validateFileSize(file)) {
      showNotification('错误', `文件过大: ${file.name}. 最大允许 ${formatFileSize(fileUploadConfig.maxFileSize)}`, 'error');
      continue;
    }
    
    // 存储文件
    filesMap.set(fileId, file);
    
    const fileItem = document.createElement('div');
    fileItem.className = 'file-item';
    fileItem.dataset.fileId = fileId;
    
    // 文件图标（根据文件类型）
    const fileIcon = getFileIcon(file);
    
    // 文件大小格式化
    const fileSize = formatFileSize(file.size);
    
    fileItem.innerHTML = `
      <span class="file-icon">${fileIcon}</span>
      <div class="file-info">
        <span class="file-name">${file.name}</span>
        <span class="file-size">${fileSize}</span>
      </div>
      <div class="file-actions">
        <span class="file-remove" title="移除">❌</span>
      </div>
      <div class="upload-progress" style="display: none;">
        <div class="progress-bar"></div>
      </div>
    `;
    
    // 添加删除功能
    const removeButton = fileItem.querySelector('.file-remove');
    removeButton.addEventListener('click', function() {
      // 从Map中删除文件
      filesMap.delete(fileId);
      fileItem.remove();
      
      // 如果没有文件了，隐藏上传按钮
      if (filesMap.size === 0) {
        uploadButton.style.display = 'none';
      }
    });
    
    // 如果是图片，添加预览功能
    if (file.type.startsWith('image/')) {
      const previewButton = document.createElement('span');
      previewButton.className = 'file-preview-button';
      previewButton.title = '预览';
      previewButton.textContent = '👁️';
      previewButton.style.marginRight = '8px';
      previewButton.style.cursor = 'pointer';
      
      previewButton.addEventListener('click', function() {
        previewImage(file);
      });
      
      fileItem.querySelector('.file-actions').prepend(previewButton);
    }
    
    filePreview.appendChild(fileItem);
  }
}

/**
 * 获取文件图标
 * @param {File} file 文件对象
 * @returns {string} 文件图标
 */
function getFileIcon(file) {
  let fileIcon = '📄';
  if (file.type.startsWith('image/')) {
    fileIcon = '🖼️';
  } else if (file.type.startsWith('video/')) {
    fileIcon = '🎬';
  } else if (file.type.startsWith('audio/')) {
    fileIcon = '🎵';
  } else if (file.type.includes('pdf')) {
    fileIcon = '📕';
  } else if (file.type.includes('word')) {
    fileIcon = '📘';
  } else if (file.type.includes('excel') || file.type.includes('sheet')) {
    fileIcon = '📗';
  } else if (file.type.includes('powerpoint') || file.type.includes('presentation')) {
    fileIcon = '📙';
  } else if (file.type.includes('text/')) {
    fileIcon = '📝';
  } else if (file.type.includes('zip') || file.type.includes('compressed')) {
    fileIcon = '🗜️';
  }
  return fileIcon;
}

/**
 * 验证文件类型
 * @param {File} file 文件对象
 * @returns {boolean} 是否为允许的文件类型
 */
function validateFileType(file) {
  return fileUploadConfig.allowedTypes.includes(file.type) || fileUploadConfig.allowedTypes.length === 0;
}

/**
 * 验证文件大小
 * @param {File} file 文件对象
 * @returns {boolean} 是否在允许的大小范围内
 */
function validateFileSize(file) {
  return file.size <= fileUploadConfig.maxFileSize;
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
 * 预览图片
 * @param {File} file 图片文件
 */
function previewImage(file) {
  const reader = new FileReader();
  reader.onload = function(e) {
    const modal = document.createElement('div');
    modal.className = 'image-preview-modal';
    modal.style.position = 'fixed';
    modal.style.top = '0';
    modal.style.left = '0';
    modal.style.width = '100%';
    modal.style.height = '100%';
    modal.style.backgroundColor = 'rgba(0, 0, 0, 0.8)';
    modal.style.zIndex = '1000';
    modal.style.display = 'flex';
    modal.style.justifyContent = 'center';
    modal.style.alignItems = 'center';
    
    const img = document.createElement('img');
    img.src = e.target.result;
    img.style.maxWidth = '90%';
    img.style.maxHeight = '90%';
    img.style.objectFit = 'contain';
    
    const closeButton = document.createElement('span');
    closeButton.textContent = '×';
    closeButton.style.position = 'absolute';
    closeButton.style.top = '20px';
    closeButton.style.right = '20px';
    closeButton.style.fontSize = '30px';
    closeButton.style.color = 'white';
    closeButton.style.cursor = 'pointer';
    
    closeButton.addEventListener('click', function() {
      document.body.removeChild(modal);
    });
    
    modal.addEventListener('click', function(e) {
      if (e.target === modal) {
        document.body.removeChild(modal);
      }
    });
    
    modal.appendChild(img);
    modal.appendChild(closeButton);
    document.body.appendChild(modal);
  };
  reader.readAsDataURL(file);
}

/**
 * 上传文件
 * @param {string} inputId 输入框ID
 */
function uploadFiles(inputId) {
  const filesMap = selectedFiles.get(inputId);
  if (filesMap.size === 0) {
    showNotification('提示', '请先选择文件', 'info');
    return;
  }
  
  // 获取文件预览区域
  const input = document.querySelector(`[data-input-id="${inputId}"]`);
  const filePreview = input.closest('.file-upload').querySelector('.file-preview');
  
  // 遍历所有文件并上传
  filesMap.forEach((file, fileId) => {
    const fileItem = filePreview.querySelector(`[data-file-id="${fileId}"]`);
    const progressBar = fileItem.querySelector('.progress-bar');
    const progressContainer = fileItem.querySelector('.upload-progress');
    
    // 显示进度条
    progressContainer.style.display = 'block';
    
    // 创建FormData
    const formData = new FormData();
    formData.append('file', file);
    
    // 创建XHR请求
    const xhr = new XMLHttpRequest();
    xhr.open('POST', fileUploadConfig.uploadEndpoint, true);
    
    // 进度事件
    xhr.upload.addEventListener('progress', function(e) {
      if (e.lengthComputable) {
        const percentComplete = (e.loaded / e.total) * 100;
        progressBar.style.width = percentComplete + '%';
      }
    });
    
    // 完成事件
    xhr.addEventListener('load', function() {
      if (xhr.status >= 200 && xhr.status < 300) {
        // 上传成功
        progressContainer.style.backgroundColor = '#4caf50';
        progressBar.style.width = '100%';
        
        // 添加成功标记
        const successIcon = document.createElement('span');
        successIcon.textContent = '✓';
        successIcon.className = 'upload-success';
        successIcon.style.color = '#4caf50';
        successIcon.style.fontWeight = 'bold';
        fileItem.querySelector('.file-actions').prepend(successIcon);
        
        showNotification('成功', `文件 ${file.name} 上传成功`, 'success');
      } else {
        // 上传失败
        progressContainer.style.backgroundColor = '#f44336';
        showNotification('错误', `文件 ${file.name} 上传失败: ${xhr.statusText}`, 'error');
      }
    });
    
    // 错误事件
    xhr.addEventListener('error', function() {
      progressContainer.style.backgroundColor = '#f44336';
      showNotification('错误', `文件 ${file.name} 上传失败: 网络错误`, 'error');
    });
    
    // 发送请求
    xhr.send(formData);
  });
}

/**
 * 显示通知
 * @param {string} title 通知标题
 * @param {string} message 通知消息
 * @param {string} type 通知类型 (success, error, warning, info)
 */
function showNotification(title, message, type = 'info') {
  // 创建通知元素
  const notification = document.createElement('div');
  notification.className = `notification notification-${type}`;
  notification.style.position = 'fixed';
  notification.style.top = '20px';
  notification.style.right = '20px';
  notification.style.padding = '15px 20px';
  notification.style.borderRadius = '4px';
  notification.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.15)';
  notification.style.zIndex = '1001';
  notification.style.minWidth = '300px';
  notification.style.maxWidth = '400px';
  notification.style.animation = 'slideIn 0.3s ease-out forwards';
  
  // 设置背景颜色
  switch (type) {
    case 'success':
      notification.style.backgroundColor = '#4caf50';
      notification.style.color = 'white';
      break;
    case 'error':
      notification.style.backgroundColor = '#f44336';
      notification.style.color = 'white';
      break;
    case 'warning':
      notification.style.backgroundColor = '#ff9800';
      notification.style.color = 'white';
      break;
    default:
      notification.style.backgroundColor = '#2196f3';
      notification.style.color = 'white';
  }
  
  // 添加内容
  notification.innerHTML = `
    <div style="display: flex; align-items: flex-start;">
      <div style="flex-grow: 1;">
        <div style="font-weight: bold; margin-bottom: 5px;">${title}</div>
        <div>${message}</div>
      </div>
      <span style="cursor: pointer; margin-left: 10px;" class="notification-close">×</span>
    </div>
  `;
  
  // 添加关闭功能
  const closeButton = notification.querySelector('.notification-close');
  closeButton.addEventListener('click', function() {
    closeNotification(notification);
  });
  
  // 添加到文档
  document.body.appendChild(notification);
  
  // 自动关闭
  setTimeout(() => {
    closeNotification(notification);
  }, 5000);
  
  // 添加动画样式
  const style = document.createElement('style');
  style.textContent = `
    @keyframes slideIn {
      from { transform: translateX(100%); opacity: 0; }
      to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
      from { transform: translateX(0); opacity: 1; }
      to { transform: translateX(100%); opacity: 0; }
    }
  `;
  document.head.appendChild(style);
  
  return notification;
}

/**
 * 关闭通知
 * @param {HTMLElement} notification 通知元素
 */
function closeNotification(notification) {
  notification.style.animation = 'slideOut 0.3s ease-in forwards';
  setTimeout(() => {
    if (document.body.contains(notification)) {
      document.body.removeChild(notification);
    }
  }, 300);
}
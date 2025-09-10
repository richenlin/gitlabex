// 公告功能
function showAnnouncement(title, message) {
  const announcementsContainer = document.querySelector('.announcements');
  if (!announcementsContainer) return;
  
  const announcement = document.createElement('div');
  announcement.className = 'announcement';
  
  const now = new Date();
  const formattedDate = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
  
  announcement.innerHTML = `
    <div class="announcement-title">${title}</div>
    <div class="announcement-content">${message}</div>
    <div class="announcement-date">${formattedDate}</div>
  `;
  
  // 添加关闭按钮
  const closeButton = document.createElement('button');
  closeButton.innerHTML = '&times;';
  closeButton.className = 'announcement-close';
  closeButton.style.float = 'right';
  closeButton.style.background = 'none';
  closeButton.style.border = 'none';
  closeButton.style.fontSize = '18px';
  closeButton.style.cursor = 'pointer';
  
  closeButton.addEventListener('click', function() {
    announcement.remove();
  });
  
  announcement.querySelector('.announcement-title').prepend(closeButton);
  announcementsContainer.prepend(announcement);
}
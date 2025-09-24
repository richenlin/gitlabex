/**
 * 表情选择器功能
 */

// 表情数据，按分类组织
const emojiData = {
  '常用': [
    { emoji: '😀', description: '笑脸' },
    { emoji: '😂', description: '笑哭' },
    { emoji: '🙂', description: '微笑' },
    { emoji: '😍', description: '爱心眼' },
    { emoji: '🤔', description: '思考' },
    { emoji: '😎', description: '酷' },
    { emoji: '😢', description: '悲伤' },
    { emoji: '😡', description: '生气' }
  ],
  '手势': [
    { emoji: '👍', description: '赞' },
    { emoji: '👎', description: '踩' },
    { emoji: '👌', description: '好的' },
    { emoji: '✌️', description: '胜利' },
    { emoji: '🤝', description: '握手' },
    { emoji: '👏', description: '鼓掌' }
  ],
  '符号': [
    { emoji: '❤️', description: '爱心' },
    { emoji: '🎉', description: '庆祝' },
    { emoji: '🔥', description: '火热' },
    { emoji: '👀', description: '眼睛' },
    { emoji: '💯', description: '满分' },
    { emoji: '⭐', description: '星星' }
  ]
};

// 当前选中的表情分类
let currentEmojiCategory = '常用';

/**
 * 初始化表情选择器
 */
function initEmojiPicker() {
  console.log('初始化表情选择器');
  
  // 查找所有表情选择器触发按钮
  const emojiTriggers = document.querySelectorAll('.emoji-picker-trigger');
  
  if (emojiTriggers.length === 0) {
    console.log('未找到表情选择器触发按钮');
    return;
  }
  
  // 为每个触发按钮添加点击事件
  emojiTriggers.forEach(trigger => {
    trigger.addEventListener('click', function(event) {
      event.stopPropagation();
      
      // 移除所有已打开的表情选择器
      removeAllEmojiPickers();
      
      // 创建表情选择器
      const emojiPicker = createEmojiPicker();
      
      // 定位表情选择器
      positionEmojiPicker(emojiPicker, trigger);
      
      // 添加表情选择事件
      setupEmojiSelection(emojiPicker, trigger);
    });
  });
  
  // 点击页面其他区域关闭表情选择器
  document.addEventListener('click', function() {
    removeAllEmojiPickers();
  });
}

/**
 * 创建表情选择器DOM
 * @returns {HTMLElement} 表情选择器元素
 */
function createEmojiPicker() {
  // 创建表情选择器容器
  const emojiPicker = document.createElement('div');
  emojiPicker.className = 'emoji-picker';
  emojiPicker.style.position = 'absolute';
  emojiPicker.style.backgroundColor = '#fff';
  emojiPicker.style.border = '1px solid #e1e4e8';
  emojiPicker.style.borderRadius = '4px';
  emojiPicker.style.padding = '10px';
  emojiPicker.style.boxShadow = '0 2px 10px rgba(0, 0, 0, 0.1)';
  emojiPicker.style.zIndex = '1000';
  emojiPicker.style.width = '300px';
  emojiPicker.style.maxHeight = '350px';
  emojiPicker.style.overflow = 'hidden';
  emojiPicker.style.display = 'flex';
  emojiPicker.style.flexDirection = 'column';
  
  // 创建搜索框
  const searchContainer = document.createElement('div');
  searchContainer.style.marginBottom = '10px';
  searchContainer.style.position = 'relative';
  
  const searchInput = document.createElement('input');
  searchInput.type = 'text';
  searchInput.placeholder = '搜索表情...';
  searchInput.style.width = '100%';
  searchInput.style.padding = '8px 10px';
  searchInput.style.border = '1px solid #ddd';
  searchInput.style.borderRadius = '4px';
  searchInput.style.boxSizing = 'border-box';
  
  searchContainer.appendChild(searchInput);
  emojiPicker.appendChild(searchContainer);
  
  // 创建分类选项卡
  const tabsContainer = document.createElement('div');
  tabsContainer.className = 'emoji-tabs';
  tabsContainer.style.display = 'flex';
  tabsContainer.style.borderBottom = '1px solid #eee';
  tabsContainer.style.marginBottom = '10px';
  
  Object.keys(emojiData).forEach(category => {
    const tab = document.createElement('button');
    tab.textContent = category;
    tab.style.padding = '5px 10px';
    tab.style.border = 'none';
    tab.style.background = 'none';
    tab.style.cursor = 'pointer';
    tab.style.borderBottom = category === currentEmojiCategory ? '2px solid #0366d6' : 'none';
    tab.style.color = category === currentEmojiCategory ? '#0366d6' : '#666';
    tab.style.fontWeight = category === currentEmojiCategory ? 'bold' : 'normal';
    
    tab.addEventListener('click', () => {
      // 更新当前分类
      currentEmojiCategory = category;
      
      // 更新标签样式
      tabsContainer.querySelectorAll('button').forEach(btn => {
        btn.style.borderBottom = 'none';
        btn.style.color = '#666';
        btn.style.fontWeight = 'normal';
      });
      tab.style.borderBottom = '2px solid #0366d6';
      tab.style.color = '#0366d6';
      tab.style.fontWeight = 'bold';
      
      // 更新表情显示
      updateEmojiDisplay(emojiContainer, searchInput.value);
    });
    
    tabsContainer.appendChild(tab);
  });
  
  emojiPicker.appendChild(tabsContainer);
  
  // 创建表情容器
  const emojiContainer = document.createElement('div');
  emojiContainer.className = 'emoji-container';
  emojiContainer.style.display = 'grid';
  emojiContainer.style.gridTemplateColumns = 'repeat(8, 1fr)';
  emojiContainer.style.gap = '5px';
  emojiContainer.style.overflowY = 'auto';
  emojiContainer.style.maxHeight = '250px';
  emojiContainer.style.padding = '5px';
  
  // 初始显示当前分类的表情
  updateEmojiDisplay(emojiContainer, '');
  
  // 添加搜索功能
  searchInput.addEventListener('input', function() {
    updateEmojiDisplay(emojiContainer, this.value);
  });
  
  emojiPicker.appendChild(emojiContainer);
  
  // 阻止点击表情选择器内部关闭选择器
  emojiPicker.addEventListener('click', function(event) {
    event.stopPropagation();
  });
  
  document.body.appendChild(emojiPicker);
  return emojiPicker;
}

/**
 * 更新表情显示
 * @param {HTMLElement} container 表情容器
 * @param {string} searchText 搜索文本
 */
function updateEmojiDisplay(container, searchText) {
  // 清空容器
  container.innerHTML = '';
  
  // 获取当前分类的表情或搜索结果
  let emojisToShow = [];
  
  if (searchText) {
    // 搜索所有分类
    Object.values(emojiData).forEach(categoryEmojis => {
      const filtered = categoryEmojis.filter(item => 
        item.description.includes(searchText.toLowerCase())
      );
      emojisToShow = emojisToShow.concat(filtered);
    });
  } else {
    // 显示当前分类
    emojisToShow = emojiData[currentEmojiCategory];
  }
  
  // 添加表情按钮
  emojisToShow.forEach(item => {
    const emojiButton = document.createElement('button');
    emojiButton.className = 'emoji-item';
    emojiButton.textContent = item.emoji;
    emojiButton.title = item.description;
    emojiButton.style.border = 'none';
    emojiButton.style.background = 'none';
    emojiButton.style.fontSize = '20px';
    emojiButton.style.cursor = 'pointer';
    emojiButton.style.padding = '5px';
    emojiButton.style.borderRadius = '4px';
    emojiButton.style.transition = 'background-color 0.2s';
    
    emojiButton.addEventListener('mouseover', function() {
      this.style.backgroundColor = '#f0f0f0';
    });
    
    emojiButton.addEventListener('mouseout', function() {
      this.style.backgroundColor = 'transparent';
    });
    
    container.appendChild(emojiButton);
  });
  
  // 如果没有结果，显示提示
  if (emojisToShow.length === 0) {
    const noResults = document.createElement('div');
    noResults.textContent = '没有找到匹配的表情';
    noResults.style.gridColumn = '1 / -1';
    noResults.style.textAlign = 'center';
    noResults.style.padding = '20px';
    noResults.style.color = '#666';
    container.appendChild(noResults);
  }
}

/**
 * 定位表情选择器
 * @param {HTMLElement} emojiPicker 表情选择器元素
 * @param {HTMLElement} trigger 触发按钮
 */
function positionEmojiPicker(emojiPicker, trigger) {
  const triggerRect = trigger.getBoundingClientRect();
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
  const scrollLeft = window.pageXOffset || document.documentElement.scrollLeft;
  
  // 获取视口尺寸
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;
  
  // 获取表情选择器尺寸
  // 由于表情选择器刚刚添加到DOM，可能需要等待一下才能获取准确尺寸
  setTimeout(() => {
    const pickerRect = emojiPicker.getBoundingClientRect();
    const pickerWidth = pickerRect.width;
    const pickerHeight = pickerRect.height;
    
    // 计算水平位置
    let left = triggerRect.left + scrollLeft;
    if (left + pickerWidth > viewportWidth) {
      // 如果会超出右边界，则向左对齐
      left = Math.max(0, viewportWidth - pickerWidth - 10);
    }
    
    // 计算垂直位置
    let top = triggerRect.bottom + scrollTop;
    if (top + pickerHeight > viewportHeight + scrollTop) {
      // 如果会超出底部边界，则显示在按钮上方
      top = triggerRect.top + scrollTop - pickerHeight;
      
      // 如果上方也放不下，则尽量靠近顶部，但保留一些空间
      if (top < scrollTop) {
        top = scrollTop + 10;
      }
    }
    
    emojiPicker.style.top = top + 'px';
    emojiPicker.style.left = left + 'px';
  }, 0);
  
  // 初始位置设置，后续会通过setTimeout调整
  emojiPicker.style.top = (triggerRect.bottom + scrollTop) + 'px';
  emojiPicker.style.left = (triggerRect.left + scrollLeft) + 'px';
}

/**
 * 设置表情选择事件
 * @param {HTMLElement} emojiPicker 表情选择器元素
 * @param {HTMLElement} trigger 触发按钮
 */
function setupEmojiSelection(emojiPicker, trigger) {
  const emojiButtons = emojiPicker.querySelectorAll('.emoji-item');
  
  emojiButtons.forEach(button => {
    button.addEventListener('click', function() {
      // 查找相关的文本区域
      const commentForm = trigger.closest('.comment-form');
      if (commentForm) {
        const textarea = commentForm.querySelector('textarea');
        if (textarea) {
          // 在光标位置插入表情
          const emoji = this.textContent;
          insertTextAtCursor(textarea, emoji);
        }
      }
      
      // 关闭表情选择器
      removeAllEmojiPickers();
    });
  });
}

/**
 * 在文本区域光标位置插入文本
 * @param {HTMLTextAreaElement} textarea 文本区域
 * @param {string} text 要插入的文本
 */
function insertTextAtCursor(textarea, text) {
  const startPos = textarea.selectionStart;
  const endPos = textarea.selectionEnd;
  const before = textarea.value.substring(0, startPos);
  const after = textarea.value.substring(endPos, textarea.value.length);
  
  textarea.value = before + text + after;
  
  // 将光标位置设置到插入文本之后
  const newPos = startPos + text.length;
  textarea.setSelectionRange(newPos, newPos);
  textarea.focus();
}

/**
 * 移除所有表情选择器
 */
function removeAllEmojiPickers() {
  const existingPickers = document.querySelectorAll('.emoji-picker');
  existingPickers.forEach(picker => {
    picker.remove();
  });
}
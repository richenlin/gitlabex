// 课题管理功能
function initSceneManagement() {
  // 示例课题数据
  const mockScenes = [
    {
      id: 1,
      title: 'LEO卫星轨道管理',
      description: '低地球轨道卫星的轨道预测和碰撞风险评估',
      thumbnail: 'https://via.placeholder.com/300x160?text=LEO轨道',
      author: '张宇航',
      createdAt: '2023-11-10',
      updatedAt: '2023-11-14',
      tags: ['卫星', '轨道预测', '碰撞预警'],
      isPublic: true,
      isShared: false
    },
    {
      id: 2,
      title: 'GEO卫星通信干扰分析',
      description: '地球静止轨道卫星间的通信干扰模式分析',
      thumbnail: 'https://via.placeholder.com/300x160?text=GEO通信',
      author: '李星辰',
      createdAt: '2023-11-08',
      updatedAt: '2023-11-12',
      tags: ['频谱分析', '干扰', '通信'],
      isPublic: true,
      isShared: true
    },
    {
      id: 3,
      title: '太空碎片追踪系统',
      description: '追踪和分析近地轨道上的太空碎片运动轨迹',
      thumbnail: 'https://via.placeholder.com/300x160?text=太空碎片',
      author: '王天宇',
      createdAt: '2023-11-05',
      updatedAt: '2023-11-10',
      tags: ['碎片', '追踪', '太空安全'],
      isPublic: false,
      isShared: true
    },
    {
      id: 4,
      title: '卫星星座部署模拟',
      description: '模拟大规模卫星星座的部署和轨道规划',
      thumbnail: 'https://via.placeholder.com/300x160?text=卫星星座',
      author: '赵航天',
      createdAt: '2023-11-03',
      updatedAt: '2023-11-08',
      tags: ['星座', '部署', '规划'],
      isPublic: true,
      isShared: false
    },
    {
      id: 5,
      title: '轨道机动计算器',
      description: '计算卫星轨道机动所需的Δv和燃料消耗',
      thumbnail: 'https://via.placeholder.com/300x160?text=轨道机动',
      author: '钱轨道',
      createdAt: '2023-11-12',
      updatedAt: '2023-11-15',
      tags: ['机动', 'Δv', '燃料'],
      isPublic: false,
      isShared: false
    },
    {
      id: 6,
      title: '太空交通管制系统',
      description: '模拟和管理太空交通流量和优先级',
      thumbnail: 'https://via.placeholder.com/300x160?text=太空交通',
      author: '孙管制',
      createdAt: '2023-11-07',
      updatedAt: '2023-11-11',
      tags: ['交通', '管制', '优先级'],
      isPublic: true,
      isShared: true
    }
  ];

  // 渲染课题列表
  function renderSceneList(scenes) {
    const sceneListContainer = document.querySelector('.scene-list');
    sceneListContainer.innerHTML = '';

    scenes.forEach(scene => {
      const sceneElement = document.createElement('div');
      sceneElement.className = 'scene-item';
      sceneElement.innerHTML = `
        <div class="scene-thumbnail">
          <img src="${scene.thumbnail}" alt="${scene.title}">
        </div>
        <div class="scene-info">
          <h3 class="scene-title">${scene.title}</h3>
          <div class="scene-meta">
            由 ${scene.author} 创建于 ${scene.createdAt} • 最后更新 ${scene.updatedAt}
          </div>
          <p>${scene.description}</p>
          <div class="scene-tags">
            ${scene.tags.map(tag => `<span class="scene-tag">${tag}</span>`).join('')}
          </div>
        </div>
      `;
      
      // 修改课题标题为链接
      const titleElement = sceneElement.querySelector('.scene-title');
      titleElement.innerHTML = `<a href="scene-detail.html">${scene.title}</a>`;
      
      sceneListContainer.appendChild(sceneElement);
    });
  }

  // 初始化课题列表
  renderSceneList(mockScenes);

  // 创建新课题按钮
  const createSceneButton = document.getElementById('create-scene-button');
  if (createSceneButton) {
    createSceneButton.addEventListener('click', function() {
      // 导航到课题创建页面
      window.location.href = 'scene-edit.html';
    });
  }
  
  // 编辑课题按钮
  document.querySelectorAll('.edit-scene-btn').forEach(button => {
    button.addEventListener('click', function(e) {
      e.preventDefault();
      e.stopPropagation();
      
      // 导航到课题编辑页面
      window.location.href = 'scene-edit.html';
    });
  });
  
  // 管理成员按钮
  document.querySelectorAll('.manage-members-btn').forEach(button => {
    button.addEventListener('click', function(e) {
      e.preventDefault();
      e.stopPropagation();
      
      // 在实际应用中，这里会打开成员管理模态窗口
      alert('成员管理功能将在实际应用中实现');
    });
  });
  
  // 搜索功能
  const searchInput = document.querySelector('.search-box input');
  const searchButton = document.querySelector('.search-button');
  
  function handleSearch() {
    const searchTerm = searchInput.value.toLowerCase();
    if (searchTerm) {
      const filteredScenes = mockScenes.filter(scene => 
        scene.title.toLowerCase().includes(searchTerm) || 
        scene.description.toLowerCase().includes(searchTerm) ||
        scene.tags.some(tag => tag.toLowerCase().includes(searchTerm))
      );
      renderSceneList(filteredScenes);
    } else {
      renderSceneList(mockScenes);
    }
  }
  
  searchButton.addEventListener('click', handleSearch);
  searchInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  });

  // 筛选功能
  const filterCheckboxes = document.querySelectorAll('.filter-options input[type="checkbox"]');
  filterCheckboxes.forEach(checkbox => {
    checkbox.addEventListener('change', () => {
      const showMine = document.querySelector('.filter-options input[type="checkbox"]:nth-child(1)').checked;
      const showShared = document.querySelector('.filter-options input[type="checkbox"]:nth-child(2)').checked;
      const showPublic = document.querySelector('.filter-options input[type="checkbox"]:nth-child(3)').checked;
      
      const filteredScenes = mockScenes.filter(scene => {
        if (showMine && scene.author === '当前用户') return true;
        if (showShared && scene.isShared) return true;
        if (showPublic && scene.isPublic) return true;
        return false;
      });
      
      renderSceneList(filteredScenes.length > 0 ? filteredScenes : mockScenes);
    });
  });

  // 课题文件列表交互
  const fileItems = document.querySelectorAll('.file-item-row');
  fileItems.forEach(item => {
    item.addEventListener('click', function() {
      const fileName = this.querySelector('.file-name').textContent;
      alert(`查看文件: ${fileName}\n在实际应用中，这将导航到文件详情页面`);
    });
  });
  
  // 话题列表交互
  const topicItems = document.querySelectorAll('.topic-item');
  topicItems.forEach(item => {
    item.addEventListener('click', function(e) {
      if (e.target.tagName !== 'A') {
        const topicTitle = this.querySelector('.topic-title').textContent;
        alert(`查看话题: ${topicTitle}\n在实际应用中，这将导航到话题详情页面`);
      }
    });
  });
  
  // 作业批改按钮
  document.querySelectorAll('.teacher-action').forEach(button => {
    button.addEventListener('click', function(e) {
      e.preventDefault();
      e.stopPropagation();
      
      // 导航到作业批改页面
      window.location.href = 'homework-grade.html';
    });
  });
}

// 模拟数据加载函数
function loadMockData() {
  console.log('加载模拟数据...');
  
  // 示例：加载热门话题
  const hotTopicsContainer = document.getElementById('hot-topics');
  if (hotTopicsContainer) {
    const mockTopics = [
      { title: '如何优化轨道预测算法？', author: '张轨道', date: '2023-11-14', comments: 24, likes: 45 },
      { title: '关于LEO卫星碰撞概率计算的讨论', author: '李太空', date: '2023-11-13', comments: 18, likes: 32 },
      { title: '太空交通管制最佳实践', author: '王宇航', date: '2023-11-12', comments: 15, likes: 28 }
    ];
    
    mockTopics.forEach(topic => {
      const topicElement = document.createElement('div');
      topicElement.className = 'topic-item';
      topicElement.innerHTML = `
        <div class="topic-header">
          <h3 class="topic-title">${topic.title}</h3>
          <span class="topic-likes">❤️ ${topic.likes}</span>
        </div>
        <div class="topic-meta">
          由 ${topic.author} 发布于 ${topic.date} • 💬 ${topic.comments} 条评论
        </div>
      `;
      hotTopicsContainer.appendChild(topicElement);
    });
  }
}

// 初始化课题管理功能
document.addEventListener('DOMContentLoaded', () => {
  initSceneManagement();
  loadMockData();
});
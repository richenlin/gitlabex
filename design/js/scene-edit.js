document.addEventListener('DOMContentLoaded', function() {
    // 获取表单元素
    const sceneForm = document.getElementById('scene-form');
    const sceneTypeRadios = document.querySelectorAll('input[name="scene-type"]');
    const membersSection = document.getElementById('members-section');
    
    // 根据课题类型显示/隐藏成员选择区域
    function toggleMembersSection() {
        const selectedType = document.querySelector('input[name="scene-type"]:checked').value;
        if (selectedType === 'private') {
            membersSection.style.display = 'block';
        } else {
            membersSection.style.display = 'none';
        }
    }
    
    // 初始化时执行一次
    toggleMembersSection();
    
    // 监听课题类型变化
    sceneTypeRadios.forEach(radio => {
        radio.addEventListener('change', toggleMembersSection);
    });
    
    // 处理表单提交
    sceneForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        // 收集表单数据
        const formData = {
            title: document.getElementById('scene-title').value,
            description: document.getElementById('scene-description').value,
            type: document.querySelector('input[name="scene-type"]:checked').value,
            tags: document.getElementById('scene-tags').value.split(',').map(tag => tag.trim()).filter(tag => tag)
        };
        
        // 如果是专有课题，收集成员信息
        if (formData.type === 'private') {
            formData.members = [];
            document.querySelectorAll('.selected-member').forEach(member => {
                const memberName = member.querySelector('.member-info span').textContent;
                formData.members.push(memberName);
            });
        }
        
        // 在实际应用中，这里应该发送数据到服务器
        console.log('提交的课题数据:', formData);
        
        // 模拟提交成功，跳转到课题列表页
        alert('课题保存成功！');
        window.location.href = 'scene.html';
    });
    
    // 处理成员移除
    document.querySelectorAll('.remove-member').forEach(button => {
        if (!button.disabled) {
            button.addEventListener('click', function() {
                const memberElement = this.closest('.selected-member');
                memberElement.remove();
            });
        }
    });
});
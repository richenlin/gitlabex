document.addEventListener('DOMContentLoaded', function() {
    // 获取所有标签和内容区域
    const tabs = document.querySelectorAll('.tab');
    const tabContents = document.querySelectorAll('.tab-content');

    // 为每个标签添加点击事件
    tabs.forEach(tab => {
        tab.addEventListener('click', function() {
            const tabId = this.getAttribute('data-tab');
            
            // 移除所有标签的active类
            tabs.forEach(t => {
                t.classList.remove('active');
            });
            
            // 移除所有内容区域的active类
            tabContents.forEach(content => {
                content.classList.remove('active');
            });
            
            // 为当前标签和对应内容添加active类
            this.classList.add('active');
            document.getElementById(`${tabId}-tab`).classList.add('active');
        });
    });
});
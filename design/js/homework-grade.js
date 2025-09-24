document.addEventListener('DOMContentLoaded', function() {
    // 获取评分表单
    const gradingForm = document.getElementById('grading-form');
    
    // 处理表单提交
    if (gradingForm) {
        gradingForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            // 获取表单数据
            const score = document.getElementById('score').value;
            const comments = document.getElementById('comments').value;
            
            // 验证分数
            if (score < 0 || score > 100) {
                alert('请输入0-100之间的分数');
                return;
            }
            
            // 在实际应用中，这里应该发送数据到服务器
            console.log('提交的评分数据:', {
                score: score,
                comments: comments
            });
            
            // 模拟提交成功，返回课题详情页
            alert('评分提交成功！');
            window.location.href = 'scene-detail.html';
        });
    }
    
    // 分数输入限制
    const scoreInput = document.getElementById('score');
    if (scoreInput) {
        scoreInput.addEventListener('input', function() {
            let value = parseInt(this.value);
            if (isNaN(value)) {
                this.value = '';
            } else if (value < 0) {
                this.value = 0;
            } else if (value > 100) {
                this.value = 100;
            }
        });
    }
});
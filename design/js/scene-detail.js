document.addEventListener('DOMContentLoaded', function() {
    // 获取模态窗口元素
    const membersModal = document.getElementById('members-modal');
    const gradeModal = document.getElementById('grade-modal');
    const manageMembersBtn = document.getElementById('manage-members-btn');
    const closeButtons = document.querySelectorAll('.close-modal');
    const teacherActions = document.querySelectorAll('.teacher-action');

    // 打开成员管理模态窗口
    if (manageMembersBtn) {
        manageMembersBtn.addEventListener('click', function() {
            membersModal.style.display = 'block';
        });
    }

    // 打开作业批改模态窗口
    teacherActions.forEach(button => {
        if (button.textContent.includes('批改作业')) {
            button.addEventListener('click', function() {
                gradeModal.style.display = 'block';
            });
        }
    });

    // 关闭模态窗口
    closeButtons.forEach(button => {
        button.addEventListener('click', function() {
            membersModal.style.display = 'none';
            gradeModal.style.display = 'none';
        });
    });

    // 点击模态窗口外部关闭
    window.addEventListener('click', function(event) {
        if (event.target === membersModal) {
            membersModal.style.display = 'none';
        }
        if (event.target === gradeModal) {
            gradeModal.style.display = 'none';
        }
    });

    // 显示/隐藏教师操作按钮（在实际应用中，这应该基于用户角色）
    function showTeacherActions() {
        // 假设当前用户是老师
        const isTeacher = true;
        
        if (isTeacher) {
            document.querySelectorAll('.teacher-action').forEach(el => {
                el.style.display = 'inline-block';
            });
        } else {
            document.querySelectorAll('.teacher-action').forEach(el => {
                el.style.display = 'none';
            });
        }
    }

    // 初始化显示教师操作按钮
    showTeacherActions();
});
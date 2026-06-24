// 异步删除 + toast
function deleteItem(url, btn, msg) {
  if (!confirm(msg || '确认删除？')) return;
  fetch(url, { method: 'POST', redirect: 'follow' })
    .then(function(r) {
      if (r.ok) {
        showToast('删除成功', 'success');
        var tr = btn.closest('tr');
        if (tr) {
          tr.style.transition = 'opacity 0.3s';
          tr.style.opacity = '0';
          setTimeout(function() { tr.remove(); }, 300);
        }
      } else {
        r.text().then(function(t) {
          var m = t.match(/<div class="alert[^"]*">([\s\S]*?)<\/div>/);
          showToast(m ? m[1].trim() : '操作失败', 'error');
        });
      }
    })
    .catch(function() { showToast('网络错误', 'error'); });
}

document.addEventListener('DOMContentLoaded', function() {
  // 暗黑模式切换
  var toggle = document.getElementById('darkToggle');
  if (toggle) {
    updateDarkIcon();
    toggle.addEventListener('click', function() {
      var isDark = document.documentElement.classList.toggle('dark');
      localStorage.setItem('darkMode', isDark);
      updateDarkIcon();
    });
  }
  function updateDarkIcon() {
    var icon = toggle ? toggle.querySelector('i') : null;
    if (!icon) return;
    icon.className = document.documentElement.classList.contains('dark') ? 'fa fa-sun-o' : 'fa fa-moon-o';
  }

  // 侧边栏高亮：精确匹配
  var path = location.pathname;
  var links = document.querySelectorAll('.admin-sidebar nav a');
  for (var i = 0; i < links.length; i++) {
    var a = links[i];
    var href = a.getAttribute('href');
    if (!href) continue;
    // 精确匹配：路径完全相等，或者路径以 href/ 开头
    if (href === path || path === href + '/') {
      a.classList.add('active');
    } else if (href !== '/' && href !== '/admin' && path.indexOf(href) === 0) {
      a.classList.add('active');
    }
  }
});

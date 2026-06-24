// 自定义模态框
function showModal(options) {
  return new Promise(function(resolve) {
    var type = options.type || 'warning';
    var icons = { warning: 'fa-exclamation-triangle', danger: 'fa-trash-o', info: 'fa-info-circle' };
    var overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.innerHTML = '<div class="modal-box">'
      + '<div class="modal-title ' + type + '"><i class="fa ' + (icons[type] || icons.warning) + '"></i>' + (options.title || '确认操作') + '</div>'
      + '<p class="modal-msg">' + (options.msg || '确定要执行此操作吗？') + '</p>'
      + '<div class="modal-actions">'
      + '<button class="btn" id="modal-cancel">取消</button>'
      + '<button class="btn ' + (type === 'danger' ? 'danger' : 'primary') + '" id="modal-confirm">' + (options.confirmText || '确定') + '</button>'
      + '</div></div>';
    document.body.appendChild(overlay);
    requestAnimationFrame(function() { overlay.classList.add('show'); });

    function close(result) {
      overlay.classList.remove('show');
      setTimeout(function() { overlay.remove(); }, 200);
      resolve(result);
    }

    overlay.querySelector('#modal-cancel').addEventListener('click', function() { close(false); });
    overlay.querySelector('#modal-confirm').addEventListener('click', function() { close(true); });
    overlay.addEventListener('click', function(e) {
      if (e.target === overlay) close(false);
    });
  });
}

// 异步删除 + toast
function deleteItem(url, btn, msg) {
  showModal({
    type: 'danger',
    title: '确认删除',
    msg: msg || '删除后无法恢复，确定要删除吗？',
    confirmText: '删除'
  }).then(function(ok) {
    if (!ok) return;
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
  });
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

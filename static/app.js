function togglePwd(btn) {
  var wrap = btn.closest('.input-eye-wrap');
  var input = wrap.querySelector('input');
  var isPwd = input.type === 'password';
  input.type = isPwd ? 'text' : 'password';
  btn.innerHTML = isPwd
    ? '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>'
    : '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>';
}

function bindAuthTransition() {
  var form = document.querySelector('.auth-form');
  if (!form || form.dataset.bound === '1') return;
  form.dataset.bound = '1';
  form.addEventListener('submit', function(e) {
    if (form.dataset.submitting === '1') return;
    e.preventDefault();
    var fields = form.querySelectorAll('[required]');
    for (var i = 0; i < fields.length; i++) {
      var f = fields[i];
      if (!f.value.trim()) {
        var label = f.closest('label');
        var name = (label ? label.textContent.trim().split(/\s+/)[0] : '') || f.name;
        showToast('请填写' + name, 'warning');
        f.focus();
        return;
      }
      if (f.type === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(f.value)) {
        showToast('邮箱格式不正确', 'warning');
        f.focus();
        return;
      }
      if (f.pattern && !new RegExp(f.pattern).test(f.value)) {
        showToast(f.title || '格式不正确', 'warning');
        f.focus();
        return;
      }
    }
    form.dataset.submitting = '1';

    var card = form.closest('.auth-card');
    var btn = form.querySelector('.auth-submit');

    var overlay = document.createElement('div');
    overlay.className = 'auth-overlay';
    overlay.innerHTML = '<div class="auth-overlay-spinner"></div><div class="auth-overlay-text">' + (form.dataset.loadingText || '处理中…') + '</div>';
    document.body.appendChild(overlay);

    if (card) card.classList.add('is-submitting');
    if (btn) {
      btn.classList.add('is-loading');
      btn.disabled = true;
    }

    var fd = new FormData(form);
    var params = [];
    fd.forEach(function(v, k) { params.push(encodeURIComponent(k) + '=' + encodeURIComponent(v)); });
    fetch(form.action, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.join('&'),
      redirect: 'follow'
    }).then(function(r) { return r.text().then(function(html) { return { url: r.url, html: html }; }); })
    .then(function(res) {
      var cur = location.pathname;
      if (res.url.indexOf(cur) !== -1 || res.url.indexOf('/login') !== -1 || res.url.indexOf('/register') !== -1) {
        var doc = new DOMParser().parseFromString(res.html, 'text/html');
        var errEl = doc.querySelector('.alert');
        var errMsg = errEl ? errEl.textContent.trim() : '操作失败，请重试';
        showToast(errMsg, 'error');
        form.dataset.submitting = '';
        if (card) card.classList.remove('is-submitting');
        if (btn) { btn.classList.remove('is-loading'); btn.disabled = false; }
        var overlay = document.querySelector('.auth-overlay');
        if (overlay) overlay.remove();
        return;
      }
      window.location.href = res.url;
    }).catch(function() {
      form.submit();
    });
  });
}

document.addEventListener('DOMContentLoaded', function() {
  bindAuthTransition();

  // 暗黑模式切换
  var toggle = document.getElementById('darkToggle');
  if (toggle) {
    updateDarkIcon();
    toggle.addEventListener('click', function() {
      var isDark = document.documentElement.classList.toggle('dark');
      localStorage.setItem('darkMode', isDark);
      updateDarkIcon();
      // 更新壁纸遮罩
      var wp = localStorage.getItem('wallpaper');
      if (wp) applyWallpaper(wp);
    });
  }
  function updateDarkIcon() {
    var icon = toggle ? toggle.querySelector('i') : null;
    if (!icon) return;
    icon.className = document.documentElement.classList.contains('dark') ? 'fa fa-sun-o' : 'fa fa-moon-o';
  }

  // 首次访问自动加载随机壁纸
  if (!localStorage.getItem('wallpaper')) {
    fetch('https://www.loliapi.com/acg/')
      .then(function(r) { var u = r.url; return r.blob().then(function(){ return u; }); })
      .then(function(url) { localStorage.setItem('wallpaper', url); applyWallpaper(url); })
      .catch(function(){});
  }

  // 换壁纸
  var wpBtn = document.getElementById('wpRefresh');
  if (wpBtn) {
    wpBtn.addEventListener('click', function() {
      wpBtn.querySelector('i').className = 'fa fa-spinner fa-spin';
      fetch('https://www.loliapi.com/acg/')
        .then(function(r) {
          var u = r.url;
          return r.blob().then(function() { return u; });
        })
        .then(function(url) {
          localStorage.setItem('wallpaper', url);
          applyWallpaper(url);
          wpBtn.querySelector('i').className = 'fa fa-picture-o';
        })
        .catch(function() {
          wpBtn.querySelector('i').className = 'fa fa-picture-o';
        });
    });
  }

  // 移动端菜单
  var menuBtn = document.getElementById('menuToggle');
  var navLinks = document.querySelector('.nav-links');
  if (menuBtn && navLinks) {
    menuBtn.addEventListener('click', function() {
      navLinks.classList.toggle('open');
    });
    var links = navLinks.querySelectorAll('a');
    for (var i = 0; i < links.length; i++) {
      links[i].addEventListener('click', function() {
        navLinks.classList.remove('open');
      });
    }
  }

  // 点赞
  var likeBtn = document.getElementById('like-btn');
  if (likeBtn) {
    likeBtn.addEventListener('click', function() {
      var url = likeBtn.dataset.url;
      fetch(url, { method: 'POST' })
        .then(function(r) { return r.json(); })
        .then(function(data) {
          if (data.code === 0) {
            var countEl = document.getElementById('like-count');
            if (countEl && data.data && data.data.like_count != null) {
              countEl.textContent = data.data.like_count;
            }
            likeBtn.innerHTML = '已赞 <span id="like-count">' + (data.data ? data.data.like_count : '') + '</span>';
            likeBtn.disabled = true;
          } else if (data.code === 401) {
            showToast('请先登录', 'warning');
          } else if (data.code === 1) {
            likeBtn.innerHTML = '已赞 <span id="like-count">' + (data.data ? data.data.like_count : '') + '</span>';
            likeBtn.disabled = true;
          }
        })
        .catch(function(e) { console.error(e); });
    });
  }

  // 当前导航高亮
  var path = location.pathname;
  var navItems = document.querySelectorAll('.nav-links > a');
  for (var i = 0; i < navItems.length; i++) {
    var a = navItems[i];
    var href = a.getAttribute('href');
    if (href === '/' && path === '/') {
      a.classList.add('active');
    } else if (href && href !== '/' && path.indexOf(href) === 0) {
      a.classList.add('active');
    }
  }
});

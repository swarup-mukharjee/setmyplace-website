/* Set My Place — redesign behaviour (loader, logo replay, before/after).
   Forms, header, reveal, counters, carousel and FAQ stay in main.js. */
(function () {
  'use strict';
  var root = document.documentElement;

  function initLoader() {
    var loader = document.getElementById('loader');
    if (!loader) return;
    if (root.classList.contains('no-loader')) { loader.remove(); return; }
    try { sessionStorage.setItem('smp-loaded', '1'); } catch (e) {}
    // CSS animation hides it at ~3s; remove it from the DOM afterwards.
    setTimeout(function () { loader.remove(); root.classList.remove('has-loader'); }, 3200);
  }

  function initLogoReplay() {
    var img = document.querySelector('.site-header .brand img');
    if (!img) return;
    var src = img.getAttribute('src');
    var busy = false;
    img.parentNode.addEventListener('mouseenter', function () {
      if (busy) return;
      busy = true;
      img.src = src.split('#')[0] + '#r' + Date.now();
      setTimeout(function () { busy = false; }, 2500);
    });
  }

  function initBeforeAfter() {
    document.querySelectorAll('.ba').forEach(function (ba) {
      var input = ba.querySelector('input[type=range]');
      if (!input) return;
      var set = function () { ba.style.setProperty('--x', input.value + '%'); };
      input.addEventListener('input', set);
      set();
    });
  }

  function start() { initLoader(); initLogoReplay(); initBeforeAfter(); }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', start);
  else start();
})();

/* ============================================================
   Set My Place — shared behaviour
   ============================================================ */
document.addEventListener('DOMContentLoaded', () => {

  /* ---------- sticky header shrink ---------- */
  const header = document.querySelector('.site-header');
  if (header) {
    let lastScrollY = window.scrollY;
    const onScroll = () => {
      const currentScrollY = window.scrollY;
      const mobileNav = document.querySelector('.mobile-nav');
      const scrollingDown = currentScrollY > lastScrollY;
      const shouldHide = currentScrollY > 80 && scrollingDown && !mobileNav?.classList.contains('open');

      header.classList.toggle('scrolled', currentScrollY > 12);
      header.classList.toggle('nav-hidden', shouldHide);
      toTopBtn && toTopBtn.classList.toggle('show', window.scrollY > 500);
      lastScrollY = currentScrollY;
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    onScroll();
  }

  /* ---------- mobile menu ---------- */
  const toggle = document.querySelector('.menu-toggle');
  const mobileNav = document.querySelector('.mobile-nav');
  if (toggle && mobileNav) {
    toggle.addEventListener('click', () => {
      const isOpen = mobileNav.classList.toggle('open');
      toggle.classList.toggle('open', isOpen);
      toggle.setAttribute('aria-expanded', isOpen);
    });
    mobileNav.querySelectorAll('a').forEach(a => {
      a.addEventListener('click', () => {
        mobileNav.classList.remove('open');
        toggle.classList.remove('open');
      });
    });
  }

  /* ---------- scroll reveal ---------- */
  const revealEls = document.querySelectorAll('.reveal, .reveal-scale, .check-item');
  const io = new IntersectionObserver((entries) => {
    entries.forEach((entry, i) => {
      if (entry.isIntersecting) {
        const delay = entry.target.dataset.delay || (i * 70);
        setTimeout(() => entry.target.classList.add('in'), delay);
        io.unobserve(entry.target);
      }
    });
  }, { threshold: 0.15 });
  revealEls.forEach(el => io.observe(el));

  /* ---------- timeline steps sequenced reveal ---------- */
  const steps = document.querySelectorAll('.step');
  const stepIO = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const idx = [...entry.target.parentElement.children].indexOf(entry.target);
        setTimeout(() => entry.target.classList.add('in'), idx * 130);
        stepIO.unobserve(entry.target);
      }
    });
  }, { threshold: 0.25 });
  steps.forEach(s => { s.classList.add('reveal'); stepIO.observe(s); });

  /* ---------- timeline truck follows the hovered booking step ---------- */
  document.querySelectorAll('.timeline-with-truck').forEach(timeline => {
    const truck = timeline.querySelector('.timeline-truck');
    const timelineSteps = timeline.querySelectorAll('.step');
    if (!truck || !timelineSteps.length) return;

    timelineSteps.forEach(step => {
      step.addEventListener('mouseenter', () => {
        const timelineBox = timeline.getBoundingClientRect();
        const stepBox = step.getBoundingClientRect();
        timeline.style.setProperty('--truck-x', `${stepBox.left - timelineBox.left + stepBox.width / 2}px`);
      });
    });
  });

  /* ---------- animated counters ---------- */
  const counters = document.querySelectorAll('[data-count]');
  const countIO = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (!entry.isIntersecting) return;
      const el = entry.target;
      const target = parseFloat(el.dataset.count);
      const suffix = el.dataset.suffix || '';
      const duration = 1200;
      const start = performance.now();
      const startVal = 0;
      const step = (now) => {
        const p = Math.min((now - start) / duration, 1);
        const eased = 1 - Math.pow(1 - p, 3);
        const val = startVal + (target - startVal) * eased;
        el.textContent = (target % 1 === 0 ? Math.round(val) : val.toFixed(1)) + suffix;
        if (p < 1) requestAnimationFrame(step);
      };
      requestAnimationFrame(step);
      countIO.unobserve(el);
    });
  }, { threshold: 0.6 });
  counters.forEach(c => countIO.observe(c));

  /* ---------- FAQ accordion ---------- */
  document.querySelectorAll('.faq-item').forEach(item => {
    const q = item.querySelector('.faq-q');
    const a = item.querySelector('.faq-a');
    q && q.addEventListener('click', () => {
      const isOpen = item.classList.contains('open');
      document.querySelectorAll('.faq-item.open').forEach(o => {
        if (o !== item) {
          o.classList.remove('open');
          o.querySelector('.faq-a').style.maxHeight = null;
        }
      });
      item.classList.toggle('open', !isOpen);
      a.style.maxHeight = !isOpen ? a.scrollHeight + 'px' : null;
    });
  });

  /* ---------- back to top ---------- */
  var toTopBtn = document.querySelector('.to-top');
  if (toTopBtn) {
    toTopBtn.addEventListener('click', () => window.scrollTo({ top: 0, behavior: 'smooth' }));
  }

  /* ---------- hero logo subtle parallax tilt (desktop only) ---------- */
  const stage = document.querySelector('.logo-hero-stage');
  if (stage && window.matchMedia('(hover:hover)').matches) {
    document.addEventListener('mousemove', (e) => {
      const x = (e.clientX / window.innerWidth - 0.5) * 10;
      const y = (e.clientY / window.innerHeight - 0.5) * 10;
      stage.style.transform = `rotate(${x * 0.15}deg) translate(${x * 0.4}px, ${y * 0.4}px)`;
    });
  }

  /* ---------- contact form (posts to the Go backend) ---------- */
  const form = document.querySelector('#contactForm');
  if (form) {
    const statusEl = document.querySelector('#formStatus');
    const sendOverlay = document.querySelector('#sendOverlay');
    const originalStatus = statusEl ? statusEl.textContent : '';

    const showSendOverlay = () => {
      sendOverlay?.classList.remove('is-sent', 'is-confirmed');
      sendOverlay?.classList.add('is-visible');
      sendOverlay?.setAttribute('aria-hidden', 'false');
    };
    const hideSendOverlay = (sent = false) => {
      if (!sendOverlay) return;
      if (sent) {
        sendOverlay.classList.add('is-sent');
        setTimeout(() => sendOverlay.classList.add('is-confirmed'), 900);
        setTimeout(() => {
          sendOverlay.classList.remove('is-visible', 'is-sent');
          sendOverlay.classList.remove('is-confirmed');
          sendOverlay.setAttribute('aria-hidden', 'true');
        }, 2300);
      } else {
        sendOverlay.classList.remove('is-visible');
        sendOverlay.setAttribute('aria-hidden', 'true');
      }
    };

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      const btn = form.querySelector('.form-submit');
      const originalBtn = btn.innerHTML;
      showSendOverlay();
      btn.disabled = true;
      btn.innerHTML = 'Sending…';
      btn.style.opacity = '.7';

      const payload = {
        name: form.name.value,
        phone: form.phone.value,
        service: form.service ? form.service.value : '',
        date: form.date ? form.date.value : '',
        message: form.message ? form.message.value : '',
        company_website: form.company_website ? form.company_website.value : '' // honeypot
      };

      try {
        const res = await fetch('/api/contact', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        const data = await res.json().catch(() => ({}));

        if (res.ok && data.success) {
          btn.innerHTML = 'Message sent ✓';
          hideSendOverlay(true);
          if (statusEl) { statusEl.textContent = "Thanks — we've got your details and will reach out shortly."; }
          form.reset();
        } else {
          hideSendOverlay();
          btn.innerHTML = 'Try again';
          if (statusEl) { statusEl.textContent = data.error || 'Something went wrong. Please call or WhatsApp us instead.'; }
        }
      } catch (err) {
        hideSendOverlay();
        btn.innerHTML = 'Try again';
        if (statusEl) { statusEl.textContent = "Couldn't reach the server. Please call or WhatsApp us instead."; }
      } finally {
        btn.disabled = false;
        btn.style.opacity = '1';
        setTimeout(() => {
          btn.innerHTML = originalBtn;
          if (statusEl) statusEl.textContent = originalStatus;
        }, 4000);
      }
    });
  }
});

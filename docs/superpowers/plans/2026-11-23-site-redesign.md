# Set My Place Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the six public pages of setmyplace.com in the approved "Warm Home + Glow" look, keeping all content, SEO tags and forms.

**Architecture:** There is one new stylesheet, `assets/site.css`, and one new script, `assets/site.js`. Public pages switch from `style.css` to `site.css`, but they **keep loading the unchanged `main.js`**. Pages reuse the class names `main.js` already drives (`.site-header`, `.menu-toggle`, `.mobile-nav`, `.reveal`, `.step`, `[data-count]`, `[data-testimonial-carousel]`, `.faq-item`, `.to-top`, `#contactForm`, `#partnerRegistrationForm`, `#sendOverlay`), so form submission code is never copied or touched. `site.js` adds only the new behaviour: loader, header logo replay, before/after slider, and a `prefers-reduced-motion` guard. `style.css` stays for the admin page.

**Tech Stack:** Static HTML, CSS, vanilla JS. Pillow is used once to convert the logo. The preview runs on `python -m http.server`, serving the `setmyplace-website/` folder.

**Spec:** `docs/superpowers/specs/2026-11-23-site-redesign-design.md`. The reference mockup is `setmyplace-website/.superpowers/brainstorm/1749-1795408504/content/hybrid-home-glow-full.html`.

**Builder's call (deviation from spec §7.1):** the spec said to copy the form logic into `site.js`. Keeping `main.js` loaded unchanged achieves the same goal ("same selectors, endpoints, payloads") with zero copy risk.

## Global Constraints

- Tokens: cream `#FBF8F3`, cream-2 `#F4EDE1`, line `#EADFCC`, navy `#0F2545`, navy-d `#0A1A33`, muted `#56637D`, orange `#F2790C` → `#FFA23C`.
- Fonts: Space Grotesk for headings, Inter for body text, both from the existing Google Fonts link.
- All existing page copy, `<title>`, meta, OG/Twitter, canonical tags and JSON-LD are kept verbatim.
- Form field `name`s and IDs are unchanged, byte for byte.
- The footer is identical to the live site's footer markup.
- The loader shows once per session (`sessionStorage` key `smp-loaded`), lasts at most ~2.7s, falls back to CSS after 3s, and becomes a fade under reduced motion.
- No horizontal scroll at 360px. No JS libraries.
- Asset links use cache-busting `?v=20261123`.

## Review Focus

1. **The loader never clears** (JS error, or `load` never fires on a slow network). The page must still become visible within 3s. Test: block `site.js` and the loader must hide by CSS.
2. **The contact "Other service" field** must reveal on selecting Other and still post `service` as the typed value. Test: in the browser, choose Other and check `#otherServiceRow` is not hidden.
3. **Partner form payload** must have the same keys as before. Test: collect `[name]` fields on the old and new page and compare the name sets; they must be equal.
4. **Mobile menu** must open and close and stay above the hero. Test: at 375px, click `.menu-toggle` and check `.mobile-nav.open` is visible.
5. **The logo on a non-white header** must have no white box. Test: screenshot the header after scrolling, when it is frosted.

---

### Task 1: Transparent animated logo

**Files:** Create `assets/logo-anim.webp`. Keep `assets/logo.png` as a static fallback.

- [ ] Convert `assets/logo.gif`: crop to the content box, turn near-white pixels (all channels > 235) fully transparent with a soft alpha ramp from 215–235, resize to 136px tall, and save as an animated WebP (loop 1, same frame durations, quality 80).
- [ ] Verify: the file is under 400KB, `Image.open(...).n_frames > 1`, and the corner pixel alpha is 0.

### Task 2: Shared design system and behaviour

**Files:** Create `assets/site.css` and `assets/site.js`.

- `site.css` sections: tokens and base, buttons (`.btn-primary`, `.btn-ghost`, `.btn-outline-navy`, `.btn-light`), loader (`#loader`), header (`.site-header`, `.navbar`, `.brand`, `.main-nav`, `.nav-cta`, `.menu-toggle`, `.mobile-nav`, `.scrolled`, `.nav-hidden`), home hero, marquee, photo service cards, stats band, before/after (`.ba`), photo steps (`.timeline`, `.step` with `.step-img`), why bento, testimonials carousel, partner block, inner page hero (`.page-hero` with photo), service rows with photos, FAQ, badge cards, benefit grid, forms (copied from `style.css` §Contact and restyled), CTA photo band, footer (identical to live), floating WhatsApp `.wa-float`, `.to-top`, send overlay (copied from `style.css`), reveal states, responsive rules at 900px and 560px, and reduced motion.
- `site.js` exports nothing. It runs `initLoader()`, `initLogoReplay()` and `initBeforeAfter()` on DOMContentLoaded.
- [ ] Verify: both files parse (no console errors on the home page).

### Task 3: Home page (`frontend/index.html`)

The section order and copy follow spec §7.5, using the reference mockup's markup, re-classed to the Task 2 names and pointing at `../assets/`. The stats keep `data-count` (600+, 5, 30+, "1 Call"). The testimonials keep the `data-testimonial-carousel` structure with all nine reviews.
- [ ] Verify in the browser: the loader runs, the logo is visible, the slider drags, counters count, and the carousel moves.

### Task 4: Services, How it works, Why us

Each page gets a photo `.page-hero` (eyebrow, H1, lead, breadcrumb), then its existing sections in the new components:
- Services: six alternating photo rows, the FAQ and the CTA.
- How it works: four customer photo steps, the six-step partner timeline and the CTA.
- Why us: the six-commitment bento, four guarantee cards, three testimonials and the CTA.
- [ ] Verify: each page renders, the FAQ toggles, and there is no console error.

### Task 5: Partner and Contact

These get the same page hero. The form markup is moved over **unchanged**; only the wrappers change.
- [ ] Verify: the Review Focus 2 and 3 checks pass. Submit against a local mock (`fetch` stub) and confirm the payload keys.

### Task 6: Full check

- [ ] Screenshot every page at desktop and 375px. Run Review Focus 1, 4 and 5. Check that the admin page still loads `style.css` and is unchanged.
- [ ] Commit on branch `redesign`.

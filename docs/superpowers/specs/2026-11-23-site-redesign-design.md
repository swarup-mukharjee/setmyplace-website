# Set My Place — Website Redesign ("Warm Home + Glow")

Date: 2026-11-23
Status: draft, awaiting approval

## 1. Intent and why

The owner feels the current setmyplace.com design looks dated. They want the
site to look, move, and feel like a modern agency site in the spirit of
omniqen.com: the overall look, the motion, the page layout, and the polished
feel. The **loader (preloader)** is a specific priority for them.

Set My Place is a home-setup service in Lucknow. It picks up where movers leave
off: shifting, unpacking, cleaning, installation and repair, house help, and
instant help. The site serves two audiences: **families moving home**, who
should book, and **moving companies and vendors**, who should register as
partners.

The business content, pages, and forms stay. Only the design is replaced.

## 2. Goals

- A fresh visual identity: the **"Warm Home + Glow" hybrid**. The owner saw
  three mockups and picked a mix. B's warm, light, photo-rich, homely look is
  the base, because A felt "too techy" for a home-services company and they
  want photos so visitors can feel what the service is like. On top of that
  come A's loader, orange glow, and motion. The reference mockup is
  `.superpowers/brainstorm/.../content/hybrid-home-glow-full.html`.
- A signature **full-screen loader**: a navy screen with the Set My Place icon,
  orange rings, a big 0–100 counter and the caption "Setting your place",
  ending with a curtain wipe that reveals the page.
- The **animated GIF logo** in the header bar. It plays once after the loader
  and again on hover.
- Real photography everywhere it helps: hero collage, service cards, steps,
  why-us, partner block, CTA band.
- A **before/after slider** of the *same* room (cluttered with boxes vs.
  finished).
- An omniqen-style structure and motion: animated hero, count-up stats, photo
  service cards, numbered photo steps, why-us bento, testimonials carousel,
  partner block, CTA band, mega footer, scroll reveal throughout.
- All six public pages redesigned consistently: `index`, `services`,
  `how-it-works`, `why-us`, `partner`, and `contact`.
- Forms keep working exactly as now, posting to `/api/contact` and
  `/api/partner-registration`.
- Mobile-first responsive, from 360px phones to wide desktops.

## 3. Non-goals

- No framework migration (no Next.js or Astro) and no build step.
- No backend, API, database, nginx, or Docker changes.
- No copywriting overhaul. Existing text is reused, with light trims only
  where a new layout needs shorter lines.
- The admin page (`setmyplaceadminsuper.html`) is not redesigned.
- No new pages (Work, Industries, Pricing, Insights) just because omniqen has
  them. They do not fit this business.

## 4. Anti-goals (failure even if it "works")

- Copying omniqen's text, graphics, logos, or brand. We borrow the style
  language only.
- A loader that annoys people. It must not block for more than ~2.5s, must
  appear only once per browser session, and must fall back gracefully.
- Broken forms or lost submissions.
- Breaking the admin page by changing the stylesheet it depends on.
- Losing SEO: titles, meta descriptions, JSON-LD, the sitemap, the Google
  verification file, and canonical URLs must survive.
- Heavy pages: no large JS libraries, and the site must stay fast on Indian
  mobile networks.

## 5. Constraints (from the codebase)

- Static HTML in `setmyplace-website/frontend/*.html`, sharing
  `setmyplace-website/assets/style.css` and `assets/main.js`. Pages reference
  assets as `../assets/...`, which resolves to `/assets/...` under nginx and
  the Go server.
- **`setmyplaceadminsuper.html` also loads `assets/style.css`.** Rewriting that
  file would break admin styling.
- `main.js` holds the form logic: `#contactForm` with `#service`,
  `#otherServiceRow`, `#otherService`, `#formStatus`, `#sendOverlay`, and
  `.form-submit`, plus `#partnerRegistrationForm` with `#partnerFormStatus`,
  which collects every `[name]` field. Field `name`s and IDs must not change.
- Production is served by nginx (`nginx/default.conf`) from `frontend/` and
  `assets/`. The Go backend handles `/api/*`.
- External links to keep: WhatsApp `wa.me/917007045262`, the WhatsApp channel
  link, and phone/email from the current pages.
- Google Fonts are already in use (Space Grotesk and Inter).

## 6. Approaches considered

| | Approach | Trade-off | Decision |
|---|---|---|---|
| A | New design on the same plain HTML/CSS/JS stack | Full layout and motion freedom; zero infra change; header and footer are duplicated across 6 pages | **Chosen**: only option that delivers the full look without touching what works |
| B | Migrate to Astro or Next.js | Shared components, closer to omniqen's tech; adds a build step and changes deployment and serving, for the same visual result | Rejected: cost and risk with no visual gain |
| C | Restyle CSS only, keep current HTML | Fast, but cannot produce the new layouts; would still feel like the old site | Rejected: fails the main goal |

Visual direction options were shown as live mockups (A Midnight glow, B Warm
light, C Bold split). The owner liked A's loader and motion, but said it was
too techy for a services company and that they want photos. The result is the
**A+B hybrid**: B's warm, photo-led base with A's loader, glow and motion.

## 7. Design decisions

### 7.1 File strategy (builder's call, decided)
- New files: **`assets/site.css`** (design system and components) and
  **`assets/site.js`** (loader, nav, motion, counters, carousel, FAQ, forms).
- Public pages switch to `site.css` and `site.js`. `style.css` and `main.js`
  stay untouched for the admin page.
- Form submission logic is **copied** from `main.js` into `site.js` unchanged
  in behaviour: same selectors, endpoints, and payloads.

### 7.2 Design tokens
- Base cream `#FBF8F3`, alternate sections `#F4EDE1`, hairlines `#EADFCC`.
- Ink and navy `#0F2545`, deep navy `#0A1A33` (loader, stats band, footer),
  muted text `#56637D`.
- Brand orange `#F2790C` → `#FFA23C`, used for buttons, kickers, glows and
  accents.
- White cards with a 1px hairline border, radius 22–32px, soft navy shadow on
  hover. Pills and buttons 999px. Max width 1200px.
- Type: Space Grotesk 600/700 headings (fluid clamp sizes), Inter body.
- Light site with navy accent bands. Dark is used only for the loader, the
  stats band, the CTA photo overlay and the footer.

### 7.3 Loader
- Full-screen navy overlay, first in `<body>`: `icon.png` with orange rings, a
  large 0→100 counter, the caption "Setting your place", and a curtain-wipe
  exit.
- The counter eases toward 90% while loading and finishes on `window.load`.
  Minimum ~900ms, maximum ~2.7s.
- After exit, the hero text rises in with a stagger and the header GIF replays.
- Shown **once per session** (`sessionStorage`); later pages get a ~300ms fade.
- `prefers-reduced-motion`: quick fade only. If JS fails, a CSS fallback hides
  it after 3s.

### 7.4 Shared chrome
- **Header**: a cream bar that becomes frosted and shadowed on scroll. On the
  left, the animated logo (`logo.gif`, cropped and converted to an animated
  WebP or MP4 under ~300KB, with `logo.png` as a fallback). In the middle, the
  six nav links. On the right, an orange "Book now" pill. On mobile, a
  slide-in menu.
- **Footer: kept as on the current live site** (owner's request). Same navy
  background, `logowhite.png`, the "We Set It. You Love It." blurb, all nine
  social links (WhatsApp, WhatsApp channel, call, email, Instagram, LinkedIn,
  Facebook, Threads, YouTube), the Company / Services / Contact columns, and
  the © bottom line. The markup and styles are carried over as-is into
  `site.css`; there is no mega footer and no big outlined wordmark.
- Floating WhatsApp button (pulsing) and back-to-top button.

### 7.5 Home page (in order)
1. **Hero**: the "Serving Lucknow & nearby areas" pill; H1 "You move their
   belongings. *We make their new house, home.*" with the second part in
   orange; the lead line; "Book a setup" and "Partner with us"; an avatar row
   with "600+ orders completed across Lucknow". On the right, a photo collage
   with an orange glow behind it.
2. **Service marquee**: a navy strip scrolling the service names.
3. **Services**: six photo cards (image, number, title, blurb, arrow).
4. **Stats band**: navy, with count-up figures (600+, 5, 30+, 1 call).
5. **Before / after**: a drag slider on the same room.
6. **How booking works**: "One call, one booking, zero stress." Four photo
   steps using the existing copy. Step 1 shows hands on a phone.
7. **Why us**: "Trust is the whole business model." A bento with one big photo
   tile (End-to-end support) and four icon cards.
8. **Testimonials**: all nine real reviews (Indian names and Lucknow areas) in
   a scroll-snap carousel with prev/next buttons. Initial avatars, no stock
   faces.
9. **Partner block**: photo plus the pitch, three chips, and "Explore the
   partner program".
10. **CTA band**: a full-width photo with a navy overlay, "Ready to make it
    feel like home?", call and WhatsApp buttons.

### 7.6 Inner pages
Each gets a shared **page hero**: eyebrow, H1 from the existing page, lead,
breadcrumb, and glow. The existing sections follow, restyled into the new
components:
- **services**: six detailed service blocks alternating left and right, the
  FAQ accordion, and a CTA.
- **how-it-works**: customer booking steps and partner referral steps as
  numbered timelines, then a CTA.
- **why-us**: the six commitments as a bento grid, four guarantees as cards,
  testimonials, and a CTA.
- **partner**: the partner pitch, a six-step "lead to payout" timeline,
  benefits cards, and the **6-section registration form** restyled as a
  stepped white form card with section headers 1–6. Fields are unchanged.
- **contact**: contact info cards (phone, WhatsApp, email, area) beside the
  contact form in a white card. The "other service" reveal and send overlay
  keep working.

### 7.7 Motion
- Scroll reveal (fade and rise) with stagger via IntersectionObserver.
- Count-up for stats, marquee, photo zoom on card hover, hover lift on cards,
  arrow-nudge buttons, header GIF replay.
- Everything is disabled or reduced under `prefers-reduced-motion`.

### 7.8 Quality bar
- Lighthouse mobile: Performance ≥ 85, Accessibility ≥ 90, SEO ≥ 95.
- Colour contrast WCAG AA (muted text ≥ 4.5:1 on cream).
- Keyboard navigable: visible focus rings, menu and FAQ operable by keyboard.
- No horizontal scroll at 360px.

## 8. Delivery phases

1. **Foundation**: `site.css` tokens and components, `site.js` core, loader,
   header and footer, with a component check page.
2. **Home page**: full rebuild of `index.html`.
3. **Inner content pages**: services, how-it-works, why-us.
4. **Form pages**: contact and partner, with the forms tested end to end
   against the Go backend.
5. **Polish**: responsive pass, accessibility, performance, SEO tags check,
   cache-busting query strings, and final cross-browser review.

**Rollout (designer's call):** all phases are built on one branch
(`redesign`) and ship together in a single release. Production never mixes
old-design and new-design pages, so no bridging styles are needed. The owner
reviews on a local preview after each phase.

**Form testing (designer's call):** testing runs only against a local Go
backend with a local throwaway Postgres (`backend/docker-compose` or the
`db` service), never production. Test submissions use obviously fake data.
Payloads are compared field by field against the old `main.js` output.

**Copy rule (designer's call):** existing page copy wins everywhere. The
mockups shortened some lines only to fit the preview. Stats stay exactly as
today (600+ orders, 5 core services, 30+ additional services, "1 Call"). The
existing "5 core services" vs six service cards mismatch is carried over
as-is; flag it to the owner, don't change it.

## 9. Left to the builder

- Exact icon set. Reuse the existing inline SVGs; add new ones in the same
  stroke style.
- Exact breakpoints, spacing scale, and animation timings within the spirit
  above.
- Photos: the mockup uses Unsplash placeholders. Swap in the owner's own job
  photos when they send them. Until then, use downloaded, self-hosted,
  compressed copies of the stock photos (WebP, lazy-loaded), not hot-links.
- Small copy trims where a layout needs them. Meaning must not change.

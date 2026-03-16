# Premium Design System Reference

Reference for building premium Astro apps with a cohesive design system.
Skills should consult this when generating UI code, theming, or styling decisions.

---

## 1. Tailwind CSS v4 @theme Setup

Tailwind v4 uses CSS-first config via the `@theme` directive. Map design tokens from
BM25 search results (brand colors, fonts, spacing) directly into `@theme` variables.

```css
@import "tailwindcss";

@theme {
  --color-primary: oklch(0.7 0.15 200);
  --color-primary-foreground: oklch(0.98 0.005 200);
  --color-secondary: oklch(0.65 0.08 260);
  --color-secondary-foreground: oklch(0.98 0.005 260);
  --color-destructive: oklch(0.55 0.2 25);
  --color-muted: oklch(0.93 0.01 260);
  --color-muted-foreground: oklch(0.45 0.02 260);
  --color-accent: oklch(0.94 0.02 260);
  --color-accent-foreground: oklch(0.2 0.02 260);
  --color-background: oklch(0.99 0.002 260);
  --color-foreground: oklch(0.14 0.02 260);
  --color-card: oklch(0.99 0.002 260);
  --color-card-foreground: oklch(0.14 0.02 260);
  --color-border: oklch(0.88 0.01 260);
  --color-ring: oklch(0.7 0.15 200);
  --font-display: "Inter Variable", sans-serif;
  --font-body: "Inter Variable", sans-serif;
  --font-mono: "JetBrains Mono Variable", monospace;
  --radius-sm: 0.25rem;
  --radius-default: 0.5rem;
  --radius-md: 0.625rem;
  --radius-lg: 0.75rem;
  --radius-xl: 1rem;
  --radius-full: 9999px;
  --shadow-xs: 0 1px 2px oklch(0 0 0 / 0.05);
  --shadow-card: 0 1px 3px oklch(0 0 0 / 0.1), 0 1px 2px oklch(0 0 0 / 0.06);
  --shadow-md: 0 4px 6px oklch(0 0 0 / 0.07), 0 2px 4px oklch(0 0 0 / 0.06);
  --shadow-lg: 0 10px 15px oklch(0 0 0 / 0.1), 0 4px 6px oklch(0 0 0 / 0.05);
  --shadow-xl: 0 20px 25px oklch(0 0 0 / 0.1), 0 8px 10px oklch(0 0 0 / 0.04);
}
```

Never use raw hex/rgb in component code -- always reference token names.

---

## 2. Component Architecture: Astro Components vs React Islands

Components split into two categories based on interactivity needs.

### Astro Components (Static, Zero JS)

Default choice for all UI. Rendered at build time, ships zero client-side JavaScript.
Use `.astro` files for layouts, cards, headers, footers, content blocks, navigation
(CSS-only dropdowns), and any element that does not require client-side state.

### React Islands (Interactive)

Use `.tsx` files with Astro `client:*` directives only when the component genuinely
requires client-side interactivity (form validation, dynamic filtering, modals with
state, live data, animations).

### Client Directive Guide

| Directive        | Hydrates When          | Use For                                      |
|------------------|------------------------|----------------------------------------------|
| `client:load`    | Immediately on load    | Critical above-fold interactive (auth forms, hero CTAs with state) |
| `client:idle`    | Browser is idle        | Low-priority interactive (newsletter signup, theme toggle) |
| `client:visible` | Element scrolls into view | Below-fold interactive (comment sections, carousels, charts) |
| `client:media`   | Media query matches    | Mobile-only or desktop-only interactive widgets |
| `client:only`    | Immediately, no SSR    | Components that cannot server-render (canvas, WebGL) |

Rule: default to no directive (static Astro component). Add a directive only when
you can name the specific interaction that requires JavaScript.

---

## 3. Color Token Reference

Three-tier architecture: **Primitive** (raw oklch values, never used directly) ->
**Semantic** (`--color-primary`, intent-based) -> **Component** (`--color-card`, scoped).

Dark/light mode switching via `:root` and `.dark`:

```css
:root {
  --color-background: oklch(0.99 0.002 260);
  --color-foreground: oklch(0.14 0.02 260);
  --color-card: oklch(0.99 0.002 260);
  --color-border: oklch(0.88 0.01 260);
}
.dark {
  --color-background: oklch(0.14 0.02 260);
  --color-foreground: oklch(0.93 0.01 260);
  --color-card: oklch(0.18 0.02 260);
  --color-border: oklch(0.3 0.02 260);
}
```

Use a small inline `<script>` in the Astro layout `<head>` to read the user's
preference from `localStorage` or `prefers-color-scheme` and set the `.dark` class
before first paint (avoids flash of wrong theme without shipping a framework).

---

## 4. Typography Scale

| Class      | Size     | Line Height | Weight | Use              |
|------------|----------|-------------|--------|------------------|
| `text-xs`  | 0.75rem  | 1rem        | 400    | Captions, labels |
| `text-sm`  | 0.875rem | 1.25rem     | 400    | Secondary text   |
| `text-base`| 1rem     | 1.5rem      | 400    | Body text        |
| `text-lg`  | 1.125rem | 1.75rem     | 500    | Subheadings      |
| `text-xl`  | 1.25rem  | 1.75rem     | 600    | Section titles   |
| `text-2xl` | 1.5rem   | 2rem        | 600    | Page subtitles   |
| `text-3xl` | 1.875rem | 2.25rem     | 700    | Page titles      |
| `text-4xl` | 2.25rem  | 2.5rem      | 700    | Hero headings    |

Self-host fonts by placing variable font files in `public/fonts/` and declaring
`@font-face` rules in your global CSS. Use `font-display: swap` to prevent FOIT.

---

## 5. Spacing, Radius, Shadow, and Animation Scales

**Spacing**: Tailwind default scale in 0.25rem units (0 through 96).

**Radius**: sm (0.25rem), default (0.5rem), md (0.625rem), lg (0.75rem), xl (1rem), full (9999px)

**Shadow**: xs, card, md, lg, xl (values defined in @theme above)

**Animation durations**: fast (150ms), default (200ms), slow (300ms), slower (500ms)

**Easing**: `--ease-default: cubic-bezier(0.4, 0, 0.2, 1)`, `--ease-in: cubic-bezier(0.4, 0, 1, 1)`, `--ease-out: cubic-bezier(0, 0, 0.2, 1)`, `--ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1)`

---

## 6. Theme Presets by App Category

**SaaS**: Blue primary (`oklch(0.6 0.15 240)`), cool grays, Inter, 0.5rem radius. Clean and professional.

**E-commerce**: Warm neutrals, amber accents (`oklch(0.75 0.15 80)`), green CTAs, 0.75rem radius. Inviting, trust-building.

**Creative**: High contrast, bold accent (`oklch(0.65 0.25 330)`), dark backgrounds, display font (Space Grotesk), mixed radii. Expressive, editorial.

**Enterprise**: Conservative blue/slate, system fonts or Inter, 0.375rem radius, WCAG AA minimum. Data-dense, predictable.

---

## 7. Iconography Rules

**No emojis anywhere in the UI.** Hard rule, no exceptions.

**Standard UI icons** -- use the better-icons MCP:
- Choose ONE collection per project and use it exclusively
- Recommended: Lucide for SaaS/enterprise, Phosphor for creative/e-commerce
- Never mix collections; never fall back to generic Material Icons or SF Symbols

**Custom brand icons** -- use Amazon Nova Canvas:
- Generate in a single batch for visual consistency
- Export as SVG, optimize with SVGO
- Store in `public/icons/` or `src/components/icons/`

**Sizing**: 16px (inline), 20px (default), 24px (prominent), 32px (feature)

---

## 8. Page Transitions (Astro View Transitions API)

Use Astro's built-in View Transitions for page-level animation. Add `<ViewTransitions />`
to your base layout:

```astro
---
import { ViewTransitions } from 'astro:transitions';
---
<html>
  <head>
    <ViewTransitions />
  </head>
  <body>
    <slot />
  </body>
</html>
```

Customize per-element transitions with `transition:animate`:

```astro
<main transition:animate="slide">
  <slot />
</main>
<header transition:persist>
  <!-- header stays across navigations -->
</header>
```

Built-in animation presets: `fade`, `slide`, `none`. Custom animations via
`transition:animate={customAnimation}` using the Animation type.

Do NOT use Motion/Framer Motion for page-level transitions. Motion is only
appropriate inside React islands for micro-interactions (button hover, modal enter/exit).

### Motion Inside React Islands

```tsx
// src/components/InteractiveCard.tsx (used with client:visible)
import { motion } from "motion/react";

export function InteractiveCard({ children }) {
  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      transition={{ type: "spring", stiffness: 400, damping: 20 }}
    >
      {children}
    </motion.div>
  );
}
```

---

## 9. Responsive Design Patterns

**Fluid typography**: `font-size: clamp(1.875rem, 1.25rem + 2.5vw, 3.75rem);`

**Container queries**:
```html
<div class="@container">
  <div class="@lg:grid-cols-3 @md:grid-cols-2 grid grid-cols-1 gap-4"></div>
</div>
```

**Breakpoints** (Tailwind defaults): sm (640px), md (768px), lg (1024px), xl (1280px), 2xl (1536px). Design mobile-first with `min-width`. Test at 320px, 375px, 768px, 1024px, 1440px.

---

## 10. Zero-JS Budget Rule

Static pages (landing, marketing, blog, docs) must ship zero client-side JavaScript.
Use Astro components exclusively on these pages. If interactivity is needed on a
static page, isolate it in a single React island with `client:visible` or `client:idle`
and verify the rest of the page ships no JS.

Run `astro build` and check the build output -- pages with zero JS will show no
client-side bundle entries. Any unexpected JS is a budget violation.

---

## 11. Premium Quality Checklist

- [ ] Custom color palette in `@theme` (not default Tailwind colors)
- [ ] Custom typography with self-hosted variable fonts and `font-display: swap`
- [ ] Custom icons from one consistent collection (not emoji, not defaults)
- [ ] Smooth page transitions with Astro View Transitions API
- [ ] Micro-interactions on interactive elements (inside React islands only)
- [ ] Consistent spacing scale (no arbitrary values)
- [ ] Dark and light mode with proper semantic token switching
- [ ] Responsive across all breakpoints (320px through 1440px+)
- [ ] Shadows and elevation for visual hierarchy
- [ ] Loading states with skeleton shimmer (not spinners alone)
- [ ] Accessible contrast ratios (WCAG AA minimum, AAA preferred)
- [ ] No raw color values in component code (all via tokens)
- [ ] Static pages ship zero client-side JavaScript

If any item is unchecked, the design is not yet premium-ready.

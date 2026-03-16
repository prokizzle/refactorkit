# AWS Bedrock Nova Asset Generation Reference (Web / Next.js)

## Prerequisites

- nova-canvas MCP server configured with AWS_PROFILE=bedrock
- AWS account with Bedrock model access enabled (Amazon Nova Canvas model)
- Next.js project with `next/image` configured

## Illustration Categories

Define prompt templates for each category. Always replace `[theme-color-palette]`
with the app's actual brand colors and `[theme-accent-color]` with the primary accent.

### Hero Section Illustrations

- Style: "modern flat illustration, [theme-color-palette], wide aspect ratio 16:9, clean lines"
- Dimensions: 1920x1080 (desktop), 1024x576 (tablet), 640x360 (mobile)
- Subjects: landing page hero, product showcase, call-to-action background
- Generate responsive variants for each breakpoint

### Onboarding Illustrations

- Style: "minimal flat illustration, soft gradients, [theme-color-palette]"
- Dimensions: 800x600
- Subjects: welcome, feature highlight, getting started, account setup

### Empty State Illustrations

- Style: "friendly minimal illustration, muted colors, [theme-color-palette], centered"
- Dimensions: 600x600
- Subjects: no items yet, no search results, offline state, first-time use

### Error State Illustrations

- Style: "empathetic minimal illustration, warm tones, [theme-color-palette]"
- Dimensions: 600x600
- Subjects: 404 not found, 500 server error, connection lost, session expired

### Feature Header Illustrations

- Style: "abstract geometric, [theme-color-palette], subtle depth"
- Dimensions: 1200x400
- Subjects: dashboard, settings, profile, activity, analytics

## Custom Icon Set

Replaces default icon libraries with brand-consistent Nova-generated icons.

- Style: "minimal vector icon, single color, [theme-accent-color], transparent background, no text, no gradients"
- Generate ALL icons in one batch session for visual consistency
- Export as SVG for web (vector-native, infinitely scalable)
- Subjects:
  - Navigation icons: home, search, menu, back, close, notifications
  - Feature icons: dashboard, analytics, messages, calendar, files, users
  - Settings icons: profile, preferences, security, billing, integrations
  - Action icons: add, edit, delete, share, download, upload, copy, save
- Generate 15-30 icons covering all app sections
- Store in `public/images/icons/` as individual SVG files

## OG Image Generation

For social sharing previews (Twitter, Facebook, LinkedIn, Slack).

- Dimensions: 1200x630
- Style: "clean branded graphic, [theme-color-palette], bold typography area"
- Include brand colors and leave space for product name overlay
- Generate per-page variants if needed (home, blog post, feature pages)
- Store in `public/images/og/`
- Reference in Next.js metadata:
  ```tsx
  export const metadata = {
    openGraph: {
      images: [{ url: '/images/og/home.webp', width: 1200, height: 630 }],
    },
  };
  ```

## Favicon and App Icons

Generate a master icon at 512x512, then scale down:

- `favicon.ico` -- 32x32 (place in `app/` directory for Next.js App Router)
- `apple-touch-icon.png` -- 180x180 (place in `public/`)
- PWA manifest icons:
  - `icon-192.png` -- 192x192
  - `icon-512.png` -- 512x512
- Style: "simple recognizable app icon, [theme-accent-color], flat design, no text"
- Reference in `manifest.json`:
  ```json
  {
    "icons": [
      { "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" },
      { "src": "/icon-512.png", "sizes": "512x512", "type": "image/png" }
    ]
  }
  ```

## Web-Specific Optimizations

### Format Selection

- WebP for photographs and complex illustrations (smaller file size, broad support)
- SVG for icons and simple illustrations (vector, no quality loss at any size)
- PNG only as fallback where transparency is needed and SVG is not viable

### Responsive Variants

Generate multiple sizes for hero and feature images:
- Desktop: 1920px wide
- Tablet: 1024px wide
- Mobile: 640px wide

### Next.js Image Component Integration

Always use the `<Image>` component for raster assets:

```tsx
import Image from 'next/image';

<Image
  src="/images/illustrations/hero-desktop.webp"
  alt="Descriptive alt text"
  width={1920}
  height={1080}
  priority  // for above-the-fold images
  sizes="(max-width: 640px) 640px, (max-width: 1024px) 1024px, 1920px"
/>
```

For SVG icons, import directly or use an inline approach for styling control.

### Asset Organization

```
public/
  images/
    illustrations/    # Hero, onboarding, empty states, error states, feature headers
    icons/            # Custom SVG icon set
    og/               # Open Graph images for social sharing
  icon-192.png        # PWA icon
  icon-512.png        # PWA icon
  apple-touch-icon.png
app/
  favicon.ico
```

## Style Guide for Consistency

- Always include the theme color palette in every prompt
- Use consistent style keywords across all prompts in a project
- Include "no text, no words, no letters" to avoid baked-in text in illustrations
- Maintain a single visual style (flat, geometric, or gradient) -- do not mix
- Document chosen style keywords in the project README or design tokens file
- Re-generate the full icon set if the style changes (partial updates break consistency)

## No Emojis Rule

Never use emojis anywhere in the application -- not in UI text, not in code comments,
not in commit messages. For all iconography needs, use custom icons generated via
Nova Canvas and the better-icons MCP server. This ensures a cohesive, brand-consistent
visual language across the entire product.

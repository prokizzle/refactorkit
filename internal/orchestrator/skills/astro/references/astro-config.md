# Astro 5 Project Configuration

Reference for configuring an Astro 5 premium web application with Vercel deployment,
React islands, Tailwind CSS, and Content Collections.

---

## astro.config.mjs

```javascript
import { defineConfig } from 'astro/config'
import vercel from '@astrojs/vercel'
import react from '@astrojs/react'
import tailwind from '@astrojs/tailwind'
import mdx from '@astrojs/mdx'
import sitemap from '@astrojs/sitemap'

export default defineConfig({
  output: 'hybrid',  // static by default, opt-in SSR per page
  adapter: vercel(),
  integrations: [react(), tailwind(), mdx(), sitemap()],
  site: 'https://example.com',
})
```

---

## Output Modes

Astro supports three output modes. Choose based on your app's needs.

### `static`

All pages pre-rendered at build time. Fastest performance, lowest hosting cost.
No server required -- output is plain HTML/CSS/JS files.

### `hybrid`

Static by default. Individual pages opt into SSR by adding:

```typescript
// src/pages/dashboard.astro
export const prerender = false
```

Recommendation: use `hybrid` for most premium apps. Static marketing pages,
blog, and docs get full CDN caching. Dashboard, auth callbacks, and API routes
use SSR only where needed.

### `server`

All pages rendered on every request. Most flexible but highest cost.
Use only when the majority of pages require dynamic data.

---

## Required Integrations

Install all at once:

```bash
npx astro add vercel react tailwind mdx sitemap
```

Or individually:

| Integration        | Install Command                    | Purpose                                    |
|--------------------|------------------------------------|--------------------------------------------|
| @astrojs/vercel    | `npx astro add vercel`             | Vercel adapter for deployment              |
| @astrojs/react     | `npx astro add react`              | React islands for interactive components   |
| @astrojs/tailwind  | `npx astro add tailwind`           | Tailwind CSS v4 integration                |
| @astrojs/mdx       | `npx astro add mdx`               | MDX for rich content pages                 |
| @astrojs/sitemap   | `npx astro add sitemap`           | Auto-generated sitemap.xml                 |
| @astrojs/db        | `npx astro add db`                 | Optional -- built-in SQLite for simple data|

---

## Content Collections (Astro 5)

### Schema Definition

```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content'

const blog = defineCollection({
  type: 'content',
  schema: z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.date(),
    author: z.string(),
    image: z.string().optional(),
    tags: z.array(z.string()).default([]),
    draft: z.boolean().default(false),
  }),
})

export const collections = { blog }
```

### Querying Collections

```typescript
import { getCollection, getEntry } from 'astro:content'

// Get all non-draft blog posts, sorted by date
const posts = (await getCollection('blog', ({ data }) => !data.draft))
  .sort((a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf())

// Get a single entry by slug
const post = await getEntry('blog', 'my-post')
```

### Content Collections vs CMS

- **Content Collections**: developer-managed content stored as Markdown/MDX in the repo.
  Best for blogs, docs, and changelogs owned by the engineering team.
- **CMS (headless)**: editor-managed content via a UI (Sanity, Contentful, Storyblok).
  Best when non-technical stakeholders need to publish without code changes.

---

## Project Structure

```
src/
  components/           # Astro components (static, zero JS)
    Header.astro
    Footer.astro
  islands/              # React components (interactive, hydrated)
    AuthButton.tsx
    PricingCalculator.tsx
    Dashboard.tsx
  content/              # Content Collections (markdown/MDX)
    config.ts
    blog/
  layouts/              # Page layouts
    BaseLayout.astro
  pages/                # File-based routing
    index.astro
    blog/[...slug].astro
    api/                # API routes
  actions/              # Astro Actions (Astro 5)
    index.ts
  styles/
    global.css          # Tailwind v4 @theme
```

Key conventions:

- `components/` holds `.astro` files that ship zero JavaScript to the client.
- `islands/` holds React (`.tsx`) components that require hydration. Always use a
  `client:*` directive when rendering these in Astro templates (see Common Issues).
- `pages/api/` contains API endpoints. These always run server-side.
- `actions/` uses Astro 5 Actions for type-safe server functions callable from the client.

---

## Monorepo Variant (Turborepo)

For app-heavy projects, use the same Turborepo structure as the premium-nextjs plugin
but replace `apps/web` with an Astro app:

```
apps/
  web/                  # Astro app (this config)
  admin/                # Optional second Astro or React app
packages/
  api/                  # Shared API client / tRPC router
  db/                   # Drizzle schema and migrations
  billing/              # Stripe integration
  ui/                   # Shared React component library
  config-ts/            # Shared tsconfig
  config-tailwind/      # Shared Tailwind preset
turbo.json
```

Astro apps import from `packages/*` the same way Next.js apps do. No special
configuration is needed beyond standard TypeScript path aliases.

---

## Vercel Deployment

The `@astrojs/vercel` adapter handles all build configuration automatically.

```bash
# Link project to Vercel
vercel link

# Pull environment variables to .env
vercel env pull

# Deploy (or push to main for automatic deploys)
vercel deploy --prod
```

No additional `vercel.json` configuration is required for standard Astro projects.
The adapter configures serverless functions for SSR routes and static assets
for pre-rendered pages.

---

## Common Issues

### "React component not interactive"

You forgot the `client:load` directive. Astro components ship zero JS by default.
To hydrate a React island:

```astro
---
import PricingCalculator from '../islands/PricingCalculator'
---
<PricingCalculator client:load />
```

Other directives: `client:idle` (hydrate on idle), `client:visible` (hydrate on scroll),
`client:media="(min-width: 768px)"` (hydrate on media query match).

### "Build fails with Node API"

A page uses Node-only APIs (fs, crypto, etc.) but is being statically rendered.
Set `output: 'hybrid'` in `astro.config.mjs` and mark the page as SSR:

```typescript
// At the top of the .astro page
export const prerender = false
```

### "Content collection not found"

Astro generates types for Content Collections. After adding or renaming a collection,
regenerate types:

```bash
npx astro sync
```

This updates the `.astro/` directory with current collection types.

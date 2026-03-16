# CMS Integration with Content Collections and Vercel Content Link

Reference for content management in a premium Astro application: Content Collections
for developer-managed content, headless CMS for editor-managed content.

---

## 1. Content Source Strategy

### Content Collections (Primary -- Developer-Managed Content)

Astro Content Collections are the default and preferred content source. Use them
for any content that developers control: documentation, blog posts, changelogs,
feature pages, landing page copy, FAQs.

Content Collections require no external service, no API calls, and produce fully
static pages with zero client-side JavaScript.

```typescript
// src/content.config.ts
import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const blog = defineCollection({
  loader: glob({ pattern: '**/*.{md,mdx}', base: './src/content/blog' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    publishedAt: z.coerce.date(),
    author: z.string(),
    tags: z.array(z.string()).default([]),
    draft: z.boolean().default(false),
  }),
});

const docs = defineCollection({
  loader: glob({ pattern: '**/*.{md,mdx}', base: './src/content/docs' }),
  schema: z.object({
    title: z.string(),
    section: z.string(),
    order: z.number(),
  }),
});

export const collections = { blog, docs };
```

Query content in pages:

```astro
---
import { getCollection } from 'astro:content';

const posts = await getCollection('blog', ({ data }) => !data.draft);
const sortedPosts = posts.sort((a, b) =>
  b.data.publishedAt.getTime() - a.data.publishedAt.getTime()
);
---
<ul>
  {sortedPosts.map((post) => (
    <li><a href={`/blog/${post.id}`}>{post.data.title}</a></li>
  ))}
</ul>
```

### CMS (Secondary -- Editor-Managed Content)

Use a headless CMS only when non-technical stakeholders need to edit content
without developer involvement. Typical cases: marketing landing pages updated
by a marketing team, product announcements by a product manager, help center
articles by support staff.

If developers are the only content editors, Content Collections are simpler,
faster, and have zero operational overhead.

---

## 2. CMS Decision Matrix

| CMS | Content Link | Self-Hosted | Free Tier | MCP | Lock-in | Best For |
|-----|-------------|-------------|-----------|-----|---------|----------|
| Payload CMS | Yes (via @payloadcms/plugin-csm) | Yes (MIT) | Unlimited | Yes | None | Full ownership, separate API server |
| Sanity | Yes (co-creator of Content Source Maps) | No | 200K API/mo, 20 users | Yes | Medium | Best managed editing experience |
| TinaCMS | Yes (@tinacms/vercel-previews) | Yes | 2 users free | No | Low | Git-backed, markdown/MDX heavy sites |

Note: Content Link requires Vercel Pro or Enterprise plan.

---

## 3. How Content Link Works

Content Source Maps embed invisible metadata into CMS API responses using stega encoding.
Stega encoding hides edit-context data (document ID, field path, CMS URL) inside
zero-width Unicode characters appended to every string value returned by the CMS API.

When a page is loaded with the Vercel toolbar active (draft/preview mode), the toolbar
parses these invisible markers and renders clickable pencil icons next to each editable
field. Clicking an icon opens the corresponding field in the CMS editor.

- No frontend code changes needed -- the toolbar reads stega-encoded strings directly.
- Content Source Maps are only active in draft/preview mode; production responses stay clean.
- The CMS client library must opt into Content Source Maps (usually a single config flag).

---

## 4. Draft Mode in Astro

Astro does not have a built-in `draftMode()` like Next.js. Implement draft mode
using Astro middleware and cookies.

### Middleware Approach

```typescript
// src/middleware.ts
import { defineMiddleware } from 'astro:middleware';

export const onRequest = defineMiddleware(async (context, next) => {
  const url = new URL(context.request.url);

  // Enable draft mode
  if (url.pathname === '/api/draft') {
    const secret = url.searchParams.get('secret');
    const slug = url.searchParams.get('slug') ?? '/';
    if (secret !== import.meta.env.DRAFT_MODE_SECRET) {
      return new Response('Invalid secret', { status: 401 });
    }
    context.cookies.set('draft-mode', 'true', {
      path: '/',
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      maxAge: 60 * 60, // 1 hour
    });
    return context.redirect(slug);
  }

  // Disable draft mode
  if (url.pathname === '/api/draft/disable') {
    context.cookies.delete('draft-mode', { path: '/' });
    return new Response('Draft mode disabled');
  }

  // Make draft status available to all pages
  context.locals.isDraft = context.cookies.get('draft-mode')?.value === 'true';
  return next();
});
```

Check `Astro.locals.isDraft` in pages to conditionally fetch draft content:

```astro
---
const isDraft = Astro.locals.isDraft;
const page = isDraft
  ? await fetchPageDraft(Astro.params.slug)
  : await fetchPagePublished(Astro.params.slug);
---
<PageContent data={page} />
```

SSR must be enabled for draft mode to work (draft pages cannot be statically generated).

---

## 5. Payload CMS Setup

Payload CMS runs as a separate server (not embedded in Astro like it is in Next.js).
Deploy as a separate Vercel project or on Railway/Render, sharing the same Neon database.

### Fetching from Payload in Astro

```typescript
// src/lib/payload.ts
const PAYLOAD_URL = import.meta.env.PAYLOAD_URL;

export async function getPage(slug: string, draft = false) {
  const params = new URLSearchParams({ 'where[slug][equals]': slug });
  if (draft) params.set('draft', 'true');
  const res = await fetch(`${PAYLOAD_URL}/api/pages?${params}`);
  const data = await res.json();
  return data.docs[0];
}
```

---

## 6. Sanity Setup

### GROQ Queries in Astro Pages

```typescript
// src/lib/sanity.ts
import { createClient } from '@sanity/client';

const client = createClient({
  projectId: import.meta.env.PUBLIC_SANITY_PROJECT_ID,
  dataset: 'production',
  apiVersion: '2024-01-01',
  useCdn: false,
  stega: { enabled: true, studioUrl: '/studio' },
});

export async function getPage(slug: string) {
  return client.fetch(
    `*[_type == "page" && slug.current == $slug][0]{ title, body }`,
    { slug }
  );
}
```

### Visual Editing

Install `@sanity/visual-editing` and add the component in your layout (only in
draft mode) so the Vercel toolbar can connect to Content Source Maps.

---

## 7. Webhook Revalidation

When content is published, a webhook triggers a rebuild or on-demand revalidation.

### Static Sites (Default)

For fully static Astro sites, CMS webhooks trigger a Vercel deploy hook to rebuild:

```
POST https://api.vercel.com/v1/integrations/deploy/prj_xxxx/hook_xxxx
```

### SSR / Hybrid Mode

For SSR pages, set short `Cache-Control` headers or use Vercel's ISR-like behavior
with `Astro.response.headers.set('Cache-Control', 's-maxage=60, stale-while-revalidate')`.

### Webhook Configuration

- **Payload CMS**: Add an `afterChange` hook to each collection that calls the deploy hook.
- **Sanity**: Create a GROQ-powered webhook in Manage -> API -> Webhooks.
- **TinaCMS**: Use the `onPut` webhook in `.tina/config.ts`.

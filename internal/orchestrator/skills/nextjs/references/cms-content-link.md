# CMS Integration with Vercel Content Link

Reference for integrating a headless CMS with Content Link in a premium Next.js application.

## CMS Decision Matrix

| CMS | Content Link | Self-Hosted | Free Tier | MCP | Lock-in | Best For |
|-----|-------------|-------------|-----------|-----|---------|----------|
| Payload CMS | Yes (via @payloadcms/plugin-csm) | Yes (MIT) | Unlimited | Yes | None | Full ownership, lives in your Next.js app |
| Sanity | Yes (co-creator of Content Source Maps) | No | 200K API/mo, 20 users | Yes | Medium | Best managed editing experience |
| TinaCMS | Yes (@tinacms/vercel-previews) | Yes | 2 users free | No | Low | Git-backed, markdown/MDX heavy sites |

Note: Content Link requires Vercel Pro or Enterprise plan.

## How Content Link Works

Content Source Maps embed invisible metadata into CMS API responses using stega encoding.
Stega encoding hides edit-context data (document ID, field path, CMS URL) inside
zero-width Unicode characters appended to every string value returned by the CMS API.

When a page is loaded with the Vercel toolbar active (draft/preview mode), the toolbar
parses these invisible markers and renders clickable pencil icons next to each editable
field. Clicking an icon opens the corresponding field in the CMS editor.

- No frontend code changes needed -- the toolbar reads stega-encoded strings directly.
- Content Source Maps are only active in draft/preview mode; production responses stay clean.
- The CMS client library must opt into Content Source Maps (usually a single config flag).

## Draft Mode Setup for App Router

Enable draft mode via a route handler:

```typescript
// app/api/draft/route.ts
import { draftMode } from 'next/headers'
import { redirect } from 'next/navigation'

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const secret = searchParams.get('secret')
  const slug = searchParams.get('slug') ?? '/'

  if (secret !== process.env.DRAFT_MODE_SECRET) {
    return new Response('Invalid secret', { status: 401 })
  }

  (await draftMode()).enable()
  redirect(slug)
}
```

Disable draft mode when the editor is done:

```typescript
// app/api/draft/disable/route.ts
import { draftMode } from 'next/headers'

export async function GET() {
  (await draftMode()).disable()
  return new Response('Draft mode disabled')
}
```

Check `isEnabled` in Server Components to conditionally fetch draft content:

```typescript
// app/[slug]/page.tsx
import { draftMode } from 'next/headers'

export default async function Page({ params }: { params: { slug: string } }) {
  const { isEnabled } = await draftMode()

  const page = isEnabled
    ? await fetchPageDraft(params.slug)   // fetch draft with stega encoding
    : await fetchPagePublished(params.slug) // fetch published, cached

  return <PageContent data={page} />
}
```

## Payload CMS Setup (Default for Owned Infrastructure)

Payload CMS lives inside your Next.js application -- same deployment, same database (Neon).
This is the recommended default when you want full data ownership.

### Installation

```bash
npx create-payload-app@latest   # select "blank" template
# Move the generated files into apps/web if using a monorepo
```

### Schema Definition

Define collections in `payload.config.ts`:

```typescript
import { buildConfig } from 'payload/config'
import { postgresAdapter } from '@payloadcms/db-postgres'
import { lexicalEditor } from '@payloadcms/richtext-lexical'
import { csmPlugin } from '@payloadcms/plugin-csm'

export default buildConfig({
  editor: lexicalEditor(),
  db: postgresAdapter({ pool: { connectionString: process.env.DATABASE_URL } }),
  collections: [
    {
      slug: 'pages',
      fields: [
        { name: 'title', type: 'text', required: true },
        { name: 'slug', type: 'text', required: true, unique: true },
        { name: 'content', type: 'richText' },
        { name: 'status', type: 'select', options: ['draft', 'published'] },
      ],
    },
    {
      slug: 'posts',
      fields: [
        { name: 'title', type: 'text', required: true },
        { name: 'slug', type: 'text', required: true, unique: true },
        { name: 'author', type: 'relationship', relationTo: 'users' },
        { name: 'body', type: 'richText' },
        { name: 'publishedAt', type: 'date' },
      ],
    },
  ],
  plugins: [csmPlugin()],
})
```

### Multi-Tenant Support

Add a `tenantId` field to every collection that requires tenant isolation:

```typescript
{
  name: 'tenantId',
  type: 'text',
  required: true,
  index: true,
  access: { read: () => true, update: () => false },
  defaultValue: ({ req }) => req.user?.tenantId,
}
```

Apply a global access control hook that filters by `tenantId` so tenants never see each other's data.

### Content Source Maps

The `@payloadcms/plugin-csm` plugin (added above) enables stega encoding in draft responses
automatically. No additional frontend changes are needed -- the Vercel toolbar picks up the
encoded metadata.

## Sanity Setup (Managed Alternative)

Use Sanity when you prefer a managed backend with a polished Studio editing experience.

### GROQ Queries in Server Components

```typescript
import { createClient } from '@sanity/client'

const client = createClient({
  projectId: process.env.NEXT_PUBLIC_SANITY_PROJECT_ID!,
  dataset: 'production',
  apiVersion: '2024-01-01',
  useCdn: false,
  stega: { enabled: true, studioUrl: '/studio' },
})

export async function getPage(slug: string) {
  return client.fetch(
    `*[_type == "page" && slug.current == $slug][0]{ title, body }`,
    { slug },
    { next: { tags: [`page:${slug}`] } }
  )
}
```

### Content Link Integration

Install the visual editing package:

```bash
npm install @sanity/visual-editing
```

Add the `VisualEditing` component to your root layout so the Vercel toolbar can
connect to Content Source Maps:

```typescript
// app/layout.tsx
import { VisualEditing } from 'next-sanity'
import { draftMode } from 'next/headers'

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const { isEnabled } = await draftMode()
  return (
    <html>
      <body>
        {children}
        {isEnabled && <VisualEditing />}
      </body>
    </html>
  )
}
```

### Sanity Studio

Sanity Studio can be embedded as a route (`app/studio/[[...tool]]/page.tsx`) or hosted
separately at `<project>.sanity.studio`.

## On-Demand ISR Webhook Pattern

When content is published, a webhook triggers on-demand revalidation so the site reflects
changes within seconds without a full rebuild.

```typescript
// app/api/revalidate/route.ts
import { revalidateTag } from 'next/cache'

export async function POST(request: Request) {
  const secret = request.headers.get('x-webhook-secret')
  if (secret !== process.env.REVALIDATION_SECRET) {
    return new Response('Unauthorized', { status: 401 })
  }

  const { tag } = await request.json()
  revalidateTag(tag)
  return Response.json({ revalidated: true })
}
```

Configure the webhook in your CMS:

- **Payload CMS**: Add an `afterChange` hook to each collection that calls the revalidate endpoint.
- **Sanity**: Create a GROQ-powered webhook in Manage -> API -> Webhooks targeting `/api/revalidate`.
- **TinaCMS**: Use the `onPut` webhook in `.tina/config.ts`.

Tag fetch calls consistently (e.g., `page:<slug>`, `post:<slug>`) so revalidation is granular.

# Architecture Decision Frameworks

Reference document for evaluating key architectural choices in Astro projects
deployed on Vercel with Neon Postgres.

---

## 1. Content-Heavy vs App-Heavy Decision Tree

Evaluate the project's content-to-interactivity ratio before choosing architecture.

### Content-Heavy (>60% Static Pages)

- Single Astro app with Content Collections as the primary content source
- Minimal React islands for isolated interactive elements
- Static site generation (SSG) for the majority of pages
- Example: marketing sites, blogs, documentation, portfolios, e-commerce catalogs
- Deploy as a single Vercel project

### App-Heavy (>60% Interactive Pages)

- Turborepo monorepo with an Astro app for marketing/content pages and a separate
  React SPA or Next.js app for the interactive dashboard
- Heavy use of React islands with `client:load` on dashboard pages
- SSR mode enabled for pages requiring authentication or dynamic data
- Example: SaaS dashboards, admin panels, real-time collaboration tools

### Hybrid (Roughly Even Split)

- Single Astro app with SSR enabled for dynamic sections
- Astro pages for marketing, blog, docs (static, zero JS)
- React islands for dashboard and app sections (interactive)
- Content Collections for developer-managed content
- CMS integration for editor-managed content
- This is the most common pattern for SaaS with a marketing site

---

## 2. Islands Architecture Rules

Astro's islands architecture is the core performance advantage. These rules
prevent degradation into a fully hydrated SPA.

### Default: Astro Component (Zero JS)

Every component starts as an `.astro` file. Only promote to a React island when
you can name the specific user interaction that requires JavaScript.

### Directive Selection

| Need | Directive | Rationale |
|------|-----------|-----------|
| Critical above-fold interactive (auth, hero CTA with state) | `client:load` | Must be interactive immediately |
| Low-priority interactive (theme toggle, newsletter) | `client:idle` | Can wait for browser idle |
| Below-fold interactive (comments, charts, carousels) | `client:visible` | Lazy hydrate on scroll into view |
| Responsive-only interactive (mobile menu) | `client:media` | Only hydrate when media query matches |
| Cannot server-render (canvas, WebGL) | `client:only="react"` | Skip SSR entirely |

### Anti-Patterns

- Never use `client:load` on every island -- defeats the purpose of Astro
- Never wrap the entire page in a single React island -- use Astro layout with targeted islands
- Never pass large serialized props to islands -- keep island boundaries narrow
- Never use a React context provider as a layout wrapper -- use Astro's built-in slot system for composition and nanostores for shared state across islands

---

## 3. Multi-Tenant Architecture Decision Tree

Use sequential-thinking to walk through these criteria before choosing a strategy.

### Option A: Row-Level Security (RLS in Postgres)

- Best for: most SaaS applications, startups, moderate tenant counts
- How it works: shared schema, shared tables, `tenant_id` column on every row, Postgres RLS policies enforce isolation
- Neon-native: RLS works out of the box, no extra infrastructure
- Pros: simple migrations (one schema), efficient resource usage, easy to query across tenants for admin
- Cons: noisy-neighbor risk under heavy load, requires discipline to always include tenant context

### Option B: Schema-per-Tenant

- Best for: regulated industries needing stronger isolation without separate databases
- Pros: strong logical isolation, per-tenant backup/restore possible, no RLS complexity
- Cons: migration complexity (must run against every schema), connection pooling harder

### Option C: Separate Databases

- Best for: enterprise customers with strict compliance (HIPAA, SOC2 with data residency)
- Pros: maximum isolation, independent scaling, per-tenant branching
- Cons: highest cost, cross-tenant queries impossible, deployment complexity multiplied

### Decision Criteria

1. **Compliance requirements** -- if data residency or strict isolation is mandated, use separate databases
2. **Data sensitivity** -- if logical isolation suffices, schema-per-tenant
3. **Tenant count** -- if >100 tenants, schema-per-tenant becomes unwieldy; prefer RLS
4. **Team size** -- if small team, prefer RLS (least operational overhead)
5. **Default choice** -- RLS with shared schema unless a specific requirement rules it out

---

## 4. Search Infrastructure Decision Tree

Plan search early. Retrofitting search into a mature codebase creates significant debt.

### Option A: pgvector in Neon

- Best for: semantic/vector search, AI-powered search, small-to-medium datasets
- Cost: free (Neon extension, no extra service)
- Choose when: you need semantic search, budget is constrained, dataset is under 1M rows

### Option B: Algolia

- Best for: instant search UIs, e-commerce, content-heavy apps needing typo tolerance
- Cost: free tier available, ~$1/1K search requests at scale
- Choose when: search UX is a core feature, budget allows, team prefers managed services

### Option C: Typesense

- Best for: teams wanting Algolia-like features without the cost
- Cost: self-hosted is free (open source), Typesense Cloud available
- Choose when: need instant search on a budget, comfortable with Docker deployment

### Option D: Meilisearch

- Best for: developer experience, rapid prototyping, smaller datasets
- Choose when: DX is priority, dataset under 10M documents, team values simplicity

### Decision Criteria

1. **Vector/semantic search needed?** -- use pgvector
2. **Search is core product feature?** -- Algolia or Typesense
3. **Budget constrained?** -- pgvector (free) or Typesense/Meilisearch (self-hosted free)
4. **Default choice** -- start with pgvector, add Typesense when full-text search UX becomes critical

---

## 5. Monolith vs Microservices

### Stay Monolith When

- Single team (or small team <5 engineers)
- Codebase under 50K LOC
- Heavy shared state between features
- Rapid iteration phase (pre-product-market fit)
- Vercel advantage: monolith deploys as a single project, simple CI/CD

### Consider Services When

- Multiple teams need independent deploy cadences
- A component has fundamentally different scaling needs (e.g., media processing)
- A component needs a different runtime (e.g., Python ML service)
- Clear bounded contexts with minimal cross-service data needs

### Vercel-Specific Considerations

- Monolith: one Vercel project, one `vercel.json`, shared preview deployments
- Microservices: separate Vercel projects per service, need explicit service-to-service auth
- Middle ground: Turborepo monorepo with shared packages but single deployable Astro app

---

## 6. Microfrontend Decision Tree (Vercel Multi-Zones)

### When to Use Multi-Zones

- Multiple teams owning distinct product areas
- Build times exceeding 10 minutes due to app size
- Different frameworks per zone (e.g., marketing site in Astro, app in Next.js)

### When to Avoid Multi-Zones

- Single team owning the entire frontend
- Frequent cross-zone navigation (each zone transition is a full page load)
- Small-to-medium app with manageable build times

---

## 7. God Module Detection Checklist

| Signal | Threshold | Action |
|--------|-----------|--------|
| File length | >300 lines | Split by responsibility into separate files |
| Unrelated imports | Imports from 3+ unrelated domains | Wrong abstraction boundary; extract cohesive modules |
| Multiple change reasons | Changed in PRs for different features | Violates single responsibility; split along feature lines |
| Shared mutable state | Module-level `let`/`var` accessed by multiple functions | Extract to a dedicated service or store |
| High fan-in | >10 other files import from it | Likely a "utils" dumping ground; break into focused modules |
| Mixed abstraction levels | HTTP handling + business logic + DB queries in one file | Layer into controller/service/repository |

### Common Offenders in Astro Projects

- `src/lib/utils.ts` -- break into `src/lib/formatting.ts`, `src/lib/validation.ts`, `src/lib/dates.ts`
- `src/components/Dashboard.astro` -- extract sub-components, move data fetching to page-level
- `src/pages/api/[...catch-all].ts` -- use proper route segments instead of a catch-all

---

## 8. Orthogonality Checklist for Turborepo Packages

Each package should have exactly one reason to change.

| Package | Changes when... | Should NOT change when... |
|---------|----------------|--------------------------|
| `packages/api` | Business logic or API contracts change | UI redesign, billing provider swap |
| `packages/db` | Schema changes, new tables, migrations | Business logic changes, UI changes |
| `packages/ui` | Design system updates, component API changes | Business logic, schema changes |
| `packages/billing` | Billing provider changes | Schema changes, UI changes |
| `packages/email` | Email templates or provider changes | Business logic, billing changes |
| `packages/auth` | Auth provider or strategy changes | Business logic, billing changes |
| `packages/config` | Shared config changes | Any runtime behavior changes |

### Dependency Direction Rule

```
apps/ --> packages/api --> packages/db
  |            |
  +---> packages/ui
  +---> packages/billing --> packages/db
  +---> packages/email
```

Dependencies flow downward. Never import from `apps/` into `packages/`. Never create circular
dependencies between packages. Use `turbo.json` pipeline dependencies to enforce build order.

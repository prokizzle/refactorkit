# Architecture Decision Frameworks

Reference document for evaluating key architectural choices in Next.js SaaS projects
deployed on Vercel with Neon Postgres.

---

## 1. Multi-Tenant Architecture Decision Tree

Use sequential-thinking to walk through these criteria before choosing a strategy.

### Option A: Row-Level Security (RLS in Postgres)

- Best for: most SaaS applications, startups, moderate tenant counts
- How it works: shared schema, shared tables, `tenant_id` column on every row, Postgres RLS policies enforce isolation
- Neon-native: RLS works out of the box, no extra infrastructure
- Pros: simple migrations (one schema), efficient resource usage, easy to query across tenants for admin
- Cons: noisy-neighbor risk under heavy load, requires discipline to always include tenant context
- Setup: enable RLS on tables, create policies using `current_setting('app.tenant_id')`, set tenant context per request in middleware

### Option B: Schema-per-Tenant

- Best for: regulated industries needing stronger isolation without separate databases
- How it works: one database, separate Postgres schema per tenant (`tenant_123.users`)
- Pros: strong logical isolation, per-tenant backup/restore possible, no RLS complexity
- Cons: migration complexity (must run against every schema), connection pooling harder, schema count limits
- Watch out: Neon branching works at database level, not schema level

### Option C: Separate Databases

- Best for: enterprise customers with strict compliance (HIPAA, SOC2 with data residency)
- How it works: one Neon project per tenant (or per region)
- Pros: maximum isolation, independent scaling, per-tenant branching
- Cons: highest cost, cross-tenant queries impossible, deployment complexity multiplied
- Neon-specific: use Neon API to automate project creation per tenant

### Decision Criteria

Evaluate in this order:

1. **Compliance requirements** -- if data residency or strict isolation is mandated, use separate databases
2. **Data sensitivity** -- if tenants must not share any infrastructure, use separate databases; if logical isolation suffices, schema-per-tenant
3. **Tenant count** -- if >100 tenants, schema-per-tenant becomes unwieldy; prefer RLS
4. **Team size** -- if small team, prefer RLS (least operational overhead)
5. **Default choice** -- RLS with shared schema unless a specific requirement rules it out

---

## 2. Search Infrastructure Decision Tree

Plan search early. Retrofitting search into a mature codebase creates significant debt.

### Option A: pgvector in Neon

- Best for: semantic/vector search, AI-powered search, small-to-medium datasets
- Cost: free (Neon extension, no extra service)
- Setup: `CREATE EXTENSION vector;`, add `vector(1536)` columns, use cosine similarity queries
- Pros: no external service, transactional consistency with your data, supports hybrid keyword+vector search
- Cons: no built-in typo tolerance, no search UI components, query performance degrades past ~1M vectors without partitioning
- Choose when: you need semantic search, budget is constrained, dataset is under 1M rows

### Option B: Algolia

- Best for: instant search UIs, e-commerce, content-heavy apps needing typo tolerance
- Cost: free tier available, ~$1/1K search requests at scale
- Setup: sync data via webhook or background job, use InstantSearch.js React components
- Pros: managed service, excellent typo tolerance, faceted search, analytics, pre-built UI widgets
- Cons: vendor lock-in, data sync lag, cost scales with usage
- Choose when: search UX is a core feature, budget allows, team prefers managed services

### Option C: Typesense

- Best for: teams wanting Algolia-like features without the cost
- Cost: self-hosted is free (open source), Typesense Cloud available
- Setup: Docker container, sync data via API, compatible with InstantSearch.js adapters
- Pros: very low cost, instant search, typo tolerance, geo search, easy to operate
- Cons: must self-host (or use cloud), smaller ecosystem than Algolia
- Choose when: need instant search on a budget, comfortable with Docker deployment

### Option D: Meilisearch

- Best for: developer experience, rapid prototyping, smaller datasets
- Cost: self-hosted is free (open source), Meilisearch Cloud available
- Setup: single binary or Docker, REST API, official JS SDK
- Pros: excellent DX, Rust-based (fast), simple configuration, good documentation
- Cons: less mature than Algolia/Typesense for large-scale production, fewer advanced features
- Choose when: DX is priority, dataset under 10M documents, team values simplicity

### Decision Criteria

1. **Vector/semantic search needed?** -- use pgvector (can combine with a text search engine)
2. **Search is core product feature?** -- Algolia (managed, polished) or Typesense (cost-effective)
3. **Budget constrained?** -- pgvector (free) or Typesense/Meilisearch (self-hosted free)
4. **Self-hosted preference?** -- Typesense or Meilisearch
5. **Default choice** -- start with pgvector, add Typesense when full-text search UX becomes critical

---

## 3. Monolith vs Microservices

### Stay Monolith When

- Single team (or small team <5 engineers)
- Codebase under 50K LOC
- Heavy shared state between features (e.g., user context, billing state used everywhere)
- Rapid iteration phase (pre-product-market fit)
- Vercel advantage: monolith deploys as a single project, simple CI/CD, shared environment variables

### Consider Services When

- Multiple teams need independent deploy cadences
- A component has fundamentally different scaling needs (e.g., media processing)
- A component needs a different runtime (e.g., Python ML service)
- Clear bounded contexts with minimal cross-service data needs

### Vercel-Specific Considerations

- Monolith: one Vercel project, one `vercel.json`, shared preview deployments
- Microservices: separate Vercel projects per service, need explicit service-to-service auth, no shared previews
- Middle ground: Turborepo monorepo with shared packages but single deployable app (recommended default)

---

## 4. Microfrontend Decision Tree (Vercel Multi-Zones)

### When to Use Multi-Zones

- Multiple teams owning distinct product areas (e.g., dashboard team, marketing site team, admin team)
- Build times exceeding 10 minutes due to app size
- Need to deploy product areas independently without risking other areas
- Different frameworks per zone (e.g., marketing site in Astro, app in Next.js)

### When to Avoid Multi-Zones

- Single team owning the entire frontend
- Frequent cross-zone navigation (each zone transition is a hard navigation / full page load)
- Deeply shared state across zones (no shared React context between zones)
- Small-to-medium app with manageable build times

### Setup

```
// microfrontends.json at repo root
{
  "applications": [
    { "name": "main", "slug": "main-app" },
    { "name": "admin", "slug": "admin-app", "basePath": "/admin" },
    { "name": "docs", "slug": "docs-app", "basePath": "/docs" }
  ]
}
```

- Install `@vercel/microfrontends` in each zone
- Each zone gets its own `basePath` in `next.config.js`
- Route rewrites in the main zone proxy to sub-zones

### Key Constraints

- Cross-zone navigation triggers full page reload (not SPA transition)
- No shared React state/context between zones
- Each zone is a separate Vercel project with its own deployment
- Authentication must be handled at the edge (middleware) or via shared cookies

---

## 5. God Module Detection Checklist

Add these checks to your review process. A module is a "god module" if any apply:

| Signal | Threshold | Action |
|--------|-----------|--------|
| File length | >300 lines | Split by responsibility into separate files |
| Unrelated imports | Imports from 3+ unrelated domains | Wrong abstraction boundary; extract cohesive modules |
| Multiple change reasons | Changed in PRs for different features | Violates single responsibility; split along feature lines |
| Shared mutable state | Module-level `let`/`var` accessed by multiple functions | Extract to a dedicated service or store |
| High fan-in | >10 other files import from it | Likely a "utils" dumping ground; break into focused modules |
| Mixed abstraction levels | HTTP handling + business logic + DB queries in one file | Layer into controller/service/repository |

### Common Offenders in Next.js Projects

- `lib/utils.ts` -- break into `lib/formatting.ts`, `lib/validation.ts`, `lib/dates.ts`
- `components/Dashboard.tsx` -- extract sub-components, move data fetching to server components
- `app/api/[...catch-all]/route.ts` -- use proper route segments instead of a catch-all

---

## 6. Orthogonality Checklist for Turborepo Packages

Each package should have exactly one reason to change. Verify your package boundaries:

| Package | Changes when... | Should NOT change when... |
|---------|----------------|--------------------------|
| `packages/api` | Business logic or API contracts change | UI redesign, billing provider swap |
| `packages/db` | Schema changes, new tables, migrations | Business logic changes, UI changes |
| `packages/ui` | Design system updates, component API changes | Business logic, schema changes |
| `packages/billing` | Billing provider changes (Stripe to Lemon Squeezy) | Schema changes, UI changes |
| `packages/email` | Email templates or provider changes | Business logic, billing changes |
| `packages/notifications` | Notification channels added/changed (email, push, SMS) | Schema changes, UI changes |
| `packages/auth` | Auth provider or strategy changes | Business logic, billing changes |
| `packages/config` | Shared config (ESLint, TypeScript, Tailwind) changes | Any runtime behavior changes |

### Violation Signals

- Changing a feature requires edits in 3+ packages -- boundaries may be wrong
- A package imports from a package at the same layer -- consider merging or extracting shared code
- A package re-exports most of another package -- likely unnecessary indirection

### Dependency Direction Rule

```
apps/ --> packages/api --> packages/db
  |            |
  +---> packages/ui
  +---> packages/billing --> packages/db
  +---> packages/email
  +---> packages/notifications --> packages/email
```

Dependencies flow downward. Never import from `apps/` into `packages/`. Never create circular
dependencies between packages. Use `turbo.json` pipeline dependencies to enforce build order.

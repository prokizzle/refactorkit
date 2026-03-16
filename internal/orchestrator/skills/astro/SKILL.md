---
name: premium-astro-setup
description: Use when setting up a premium Astro 5 web app from scratch or upgrading an existing one. Triggers on 'premium Astro setup', 'premium Astro app', 'make my Astro app production-ready', 'Astro islands', 'content-first web app', 'Astro Content Collections'
---

# Premium Astro Setup

Master orchestrator for transforming any Astro 5 project into a premium, production-grade web application. Content-first islands architecture. Walks through discovery, infrastructure, architecture, design system, growth, monetization, observability, and launch readiness.

Philosophy: "There is no later. AI later is 24 hours."

## Prerequisites

Before starting, verify:
- Astro CLI installed (`astro --version`)
- Node.js 18+ installed (`node --version`)
- Vercel CLI installed (`vercel --version`)
- AWS CLI configured for Bedrock (`aws bedrock list-foundation-models`)
- Run `setup.sh` from this plugin for BM25 dependencies
- **marketing-skills plugin enabled:** Required for Phase 3.5 conversion optimization skills
- **font-tools plugin enabled:** Required for Phase 3 premium typography

Discover additional skills:
```
/find-skills clerk sentry responsive
```
Look for: `clerk-*`, `sentry-setup-*`, responsive design skills.

Verify MCP connections: `sequential-thinking`, `context7`, `better-icons`.

## Orchestration Flow

Follow these phases in order. Use sequential-thinking MCP for all major decisions.

```dot
digraph phases {
    rankdir=TB;
    "Phase 0: Discovery" -> "Phase 0.5: Product Vision";
    "Phase 0.5: Product Vision" -> "Phase 1: Ecosystem & Infra";
    "Phase 1: Ecosystem & Infra" -> "Phase 2: Architecture";
    "Phase 2: Architecture" -> "Phase 3: Design System";
    "Phase 3: Design System" -> "Phase 3.5: Growth & Conversion";
    "Phase 3.5: Growth & Conversion" -> "Phase 4: Monetization";
    "Phase 4: Monetization" -> "Phase 5: Observability";
    "Phase 5: Observability" -> "Phase 6: Launch Readiness";
}
```

---

## Phase 0 -- Discovery & Research

### 0.1 Environment Setup
Run `/find-skills` to discover available skills. Verify all MCP connections listed in Prerequisites.

### 0.2 Mode Fork
Ask: "Starting fresh or upgrading an existing project?"

### 0.3 Version Verification
Both modes: use firecrawl + context7 MCPs to verify latest stable versions of Astro, React, Tailwind, shadcn/ui. Generate a version manifest pinning all dependencies.

### 0.4 Existing Project Audit (upgrade mode only)
Analyze `astro.config.mjs`: output mode (`static`, `server`, `hybrid`), integrations, content collections schema, adapter configuration. Produce tech debt audit and migration plan using sequential-thinking.

---

## Phase 0.5 -- Product Marketing Context + Brainstorming + Nova Assets

Invoke `product-marketing-context` to establish positioning, audience, and brand voice.
Invoke `superpowers:brainstorming` for product vision exploration.
Use sequential-thinking MCP to align product vision with technical decisions.

Generate Nova assets early (logo, hero image, OG images) -- see `references/nova-assets.md`.
Establish iconography direction: better-icons MCP for standard UI icons, Nova for custom brand artwork.

---

## Phase 1 -- Ecosystem Decision + Infrastructure + CMS

### 1.1 Ecosystem Decision
Evaluate: Composable (Clerk + Neon + Upstash) vs Supabase vs Firebase. See `references/scaling-cost-matrix.md`.

### 1.2 Content Strategy Decision
Evaluate: Content Collections only vs Content Collections + headless CMS vs full CMS. See `references/cms-content-link.md`.

### 1.3 Project Structure
Single Astro app vs Turborepo monorepo. See `references/astro-config.md`. Key directories:
- `src/content/` -- Content Collections with typed schemas
- `src/components/` -- Astro (static) + React (islands) components
- `src/actions/` -- Astro Actions for server mutations
- `src/pages/api/` -- tRPC or REST API routes
- `src/lib/db/` -- Drizzle ORM + migrations
- `src/lib/billing/` -- payment abstraction
- `src/lib/email/` -- transactional email templates

### 1.4 Astro Integrations Setup
Configure integrations in `astro.config.mjs`. See `references/astro-config.md`:
- `@astrojs/react` for interactive islands
- `@astrojs/tailwind` for styling
- `@astrojs/vercel` adapter for deployment
- `@astrojs/sitemap` for SEO
- `@clerk/astro` for authentication

### 1.5 Vercel Connection
Run `vercel link` and `vercel env pull` to connect deployment. Verify Vercel deployment succeeds.

### 1.6 Write CLAUDE.md Agent Guardrails
Add build commands, service layer rules, and forbidden patterns. See `references/service-layer-pattern.md`.

---

## Phase 2 -- Architecture + Islands + Event Bus

### 2.1 Content-Heavy vs App-Heavy Evaluation
Determine output mode: `static` for content sites, `server` for dynamic apps, `hybrid` for mixed. See `references/architecture-decisions.md`.

### 2.2 Multi-Tenant Strategy
Row-level isolation vs schema-per-tenant vs database-per-tenant. See `references/architecture-decisions.md`. Use `references/drizzle-schema.md` for schema patterns.

### 2.3 Service Layer Pattern
All business logic in services, never in Astro Actions or API routes directly. Astro Actions + tRPC for the API surface. See `references/service-layer-pattern.md`.

### 2.4 Islands Architecture Rules
Define hydration boundaries. See `references/architecture-decisions.md`:
- Default to Astro components (zero JS) for all static content
- Use `client:load` only for above-fold interactive elements
- Use `client:visible` for below-fold interactive elements
- Use `client:idle` for non-critical interactivity
- Never hydrate components that do not require client-side state

### 2.5 Event Bus (Inngest)
Set up Inngest for background jobs, scheduled tasks, and event-driven workflows.

### 2.6 Realtime + Push
WebSocket strategy (Ably/Pusher/Supabase Realtime) + web push notifications. See `references/realtime-push.md`.

### 2.7 Search Infrastructure
Evaluate: Algolia vs Typesense vs Meilisearch vs pg_trgm. See `references/architecture-decisions.md`.

### 2.8 Cache Strategy
CDN cache headers via Vercel adapter, `Cache-Control` on static assets, `vercel purgeCache` for on-demand invalidation. No `"use cache"` directive -- Astro uses file-based caching and CDN headers.

### 2.9 View Transitions Setup
Add `<ViewTransitions />` to the base layout for SPA-like page transitions:
```astro
---
import { ViewTransitions } from 'astro:transitions';
---
<html>
  <head><ViewTransitions /></head>
  <body><slot /></body>
</html>
```

### 2.10 API Documentation
tRPC Panel for internal APIs, OpenAPI spec generation for public APIs.

---

## Phase 3 -- Premium Design System + Nova Assets

### 3.1 BM25 Design Search
Run BM25 design search with the product description to find relevant design inspiration and patterns.

### 3.2 Design Tokens
Generate shadcn/ui + Tailwind v4 theme tokens (colors, spacing, typography, radius). See `references/premium-design-system.md`.

### 3.3 Asset Generation
See `references/nova-assets.md` for prompt templates.
- **UI icons:** better-icons MCP -- choose ONE icon collection for consistency
- **Brand assets:** Nova for custom icons, illustrations, and marketing visuals
- **Fonts:** Invoke `google-font-downloader` skill for premium typography

### 3.4 Component Split
Astro components for static content, React islands for interactive elements. See `references/premium-design-system.md` for token integration and composability rules:
- `.astro` files for layout, navigation, content sections (zero JS shipped)
- React components only inside `client:*` islands for forms, dashboards, interactive widgets
- shadcn/ui components used exclusively within React islands

### 3.5 View Transitions + Motion
View Transitions API for page-level transitions (handled by Astro). Motion library only inside React islands for component-level animation. Never ship animation JS to static pages.

### 3.6 Humanize All Copy
Invoke `humanizer` on all AI-generated copy -- onboarding, empty states, error messages, CTAs, marketing pages. Run AFTER writing copy, BEFORE finalizing design.

---

## Phase 3.5 -- Growth & Conversion + Email Marketing

Invoke marketing-skills in sequence:
`product-marketing-context` (from 0.5) -> `onboarding-cro` -> `signup-flow-cro` -> `paywall-upgrade-cro` -> `pricing-strategy` -> `churn-prevention` -> `referral-program` -> `ab-test-setup` -> `launch-strategy`

All copy through humanizer. All visuals Nova-generated.

Email marketing setup: Resend + React Email for transactional and drip campaigns. See `references/email-marketing.md`.

User data syncing via Clerk webhooks to CRM and email lists.

---

## Phase 4 -- Monetization + Billing

### 4.1 Billing Setup
See `references/billing-abstraction.md`. Implement billing abstraction:
- Stripe adapter (default)
- Clerk Billing adapter (optional)
- Custom adapter interface for future providers
- Astro Actions for checkout session creation and webhook handling

### 4.2 Transactional Email
Resend + React Email templates for receipts, trial expiry, dunning, welcome sequences. See `references/email-marketing.md`.

---

## Phase 5 -- Observability + Audit

See `references/observability-stack.md`.

Invoke: `sentry-setup-logging`, `sentry-setup-tracing`, `sentry-setup-metrics` using `@sentry/astro`.

Additional observability:
- PostHog for product analytics and feature flags
- Vercel Analytics + Speed Insights
- Astro-specific audits: verify zero JS on static pages, audit island hydration directives, confirm no unnecessary `client:load` usage

---

## Phase 6 -- Launch Readiness

Final checklist before shipping:
- **Security:** env vars audited, CORS configured, rate limiting via Upstash
- **SEO:** metadata in layouts, dynamic OG images, sitemap.xml via `@astrojs/sitemap`, robots.txt
- **Accessibility:** WCAG 2.1 AA compliance verified
- **Performance:** Core Web Vitals passing (LCP < 2.5s, INP < 200ms, CLS < 0.1)
- **Static pages:** zero JavaScript shipped on content-only pages
- **View Transitions:** working across all page navigations, no flicker
- **Content Collections:** all schemas typed, `getCollection()` calls validated
- **Islands:** every `client:*` directive justified, no over-hydration
- **API:** OpenAPI spec published, tRPC panel accessible in dev
- **CI/CD:** preview deployments per PR, staging environment configured
- **CMS:** content preview and draft mode verified (see `references/cms-content-link.md`)
- **Push:** web push notifications tested end-to-end
- **Realtime:** cache invalidation and live updates verified
- **Launch:** invoke `launch-strategy` skill for phased rollout plan

---

## Quick Reference

| Layer | Tool | Package | Key API |
|---|---|---|---|
| Framework | Astro 5 | `astro` | Content Collections, View Transitions, `Astro.glob()` |
| Islands | React | `@astrojs/react` | `client:load`, `client:visible`, `client:idle` |
| Styling | Tailwind v4 | `tailwindcss` | `@theme`, CSS variables, `cn()` utility |
| Components | shadcn/ui | `@shadcn/ui` | `Button`, `Card`, `Dialog` (React islands only) |
| Auth | Clerk | `@clerk/astro` | `<SignIn />`, middleware, `locals.auth()` |
| Database | Drizzle ORM | `drizzle-orm` | `db.select()`, `db.insert()`, schema definitions |
| API | tRPC + Astro Actions | `@trpc/server` | `router`, `procedure`, `defineAction()` |
| Billing | Stripe | `stripe` | `stripe.checkout.sessions`, `stripe.subscriptions` |
| Email | Resend | `resend` | `resend.emails.send()`, React Email templates |
| Events | Inngest | `inngest` | `inngest.createFunction()`, `inngest.send()` |
| Transitions | View Transitions | `astro:transitions` | `<ViewTransitions />`, `transition:name` |
| Animation | Motion | `motion` | `motion.div`, `AnimatePresence` (React islands only) |
| Analytics | PostHog | `posthog-js` | `posthog.capture()`, feature flags |
| Monitoring | Sentry | `@sentry/astro` | `Sentry.captureException()`, tracing |
| Hosting | Vercel | `@astrojs/vercel` | `vercel deploy`, preview URLs, env vars |

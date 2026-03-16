# Observability Stack Reference -- Premium Astro on Vercel

This reference covers the full observability setup for a premium Astro
application deployed on Vercel: error tracking, product analytics, performance
monitoring, and audit tooling.

---

## 1. Sentry Setup (@sentry/astro)

### Auto-Setup

```bash
npx astro add @sentry/astro
```

This adds the Sentry integration to `astro.config.mjs` and creates a
`sentry.client.config.ts` file for browser-side initialization.

### Manual Config in astro.config.mjs

```typescript
import { defineConfig } from 'astro/config';
import sentry from '@sentry/astro';

export default defineConfig({
  integrations: [
    sentry({
      dsn: import.meta.env.PUBLIC_SENTRY_DSN,
      sourceMapsUploadOptions: {
        org: import.meta.env.SENTRY_ORG,
        project: import.meta.env.SENTRY_PROJECT,
        authToken: import.meta.env.SENTRY_AUTH_TOKEN,
      },
    }),
  ],
});
```

### Client-Side Config

```typescript
// sentry.client.config.ts
import * as Sentry from '@sentry/astro';

Sentry.init({
  dsn: import.meta.env.PUBLIC_SENTRY_DSN,
  tracesSampleRate: 1.0,
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,
  integrations: [
    Sentry.replayIntegration({
      maskAllText: false,
      blockAllMedia: false,
    }),
  ],
});
```

### Source Maps

Source maps upload automatically during `astro build` when `SENTRY_AUTH_TOKEN` is
set. On Vercel, add the Sentry integration or set the token in environment variables.

### Skills to Invoke

After initial setup, invoke these skills for deeper integration:

- `sentry-setup-logging` -- structured logging with breadcrumbs
- `sentry-setup-tracing` -- distributed tracing across API endpoints
- `sentry-setup-metrics` -- custom metrics and dashboards

---

## 2. PostHog Setup (Product Analytics + Feature Flags)

### Client-Side (Inline Script in Layout)

Add PostHog in your base Astro layout. Since Astro pages are static by default,
use an inline script rather than a React provider:

```astro
---
// src/layouts/Base.astro
---
<html>
  <head>
    <script define:vars={{ posthogKey: import.meta.env.PUBLIC_POSTHOG_KEY }}>
      !function(t,e){var o,n,p,r;e.__SV||(window.posthog=e,e._i=[],e.init=function(i,s,a){function g(t,e){var o=e.split(".");2==o.length&&(t=t[o[0]],e=o[1]),t[e]=function(){t.push([e].concat(Array.prototype.slice.call(arguments,0)))}}(p=t.createElement("script")).type="text/javascript",p.async=!0,p.src=s.api_host+"/static/array.js",(r=t.getElementsByTagName("script")[0]).parentNode.insertBefore(p,r);var u=e;for(void 0!==a?u=e[a]=[]:a="posthog",u.people=u.people||[],u.toString=function(t){var e="posthog";return"posthog"!==a&&(e+="."+a),t||(e+=" (stub)"),e},u.people.toString=function(){return u.toString(1)+".people (stub)"},o="init capture register register_once register_for_session unregister opt_out_capturing has_opted_out_capturing opt_in_capturing reset isFeatureEnabled onFeatureFlags getFeatureFlag getFeatureFlagPayload reloadFeatureFlags group updateEarlyAccessFeatureEnrollment getEarlyAccessFeatures getActiveMatchingSurveys getSurveys onSessionId".split(" "),n=0;n<o.length;n++)g(u,o[n]);e._i.push([i,s,a])},e.__SV=1)}(document,window.posthog||[]);
      posthog.init(posthogKey, {
        api_host: 'https://us.i.posthog.com',
        person_profiles: 'identified_only',
      });
    </script>
  </head>
  <body><slot /></body>
</html>
```

For React islands that need PostHog, import `posthog-js` directly and call
`posthog.capture()` within the island component.

### Server-Side Tracking

```typescript
// src/lib/posthog-server.ts
import { PostHog } from 'posthog-node';

const posthogServer = new PostHog(import.meta.env.POSTHOG_API_KEY, {
  host: import.meta.env.POSTHOG_HOST ?? 'https://us.i.posthog.com',
  flushAt: 1,
  flushInterval: 0,
});

export { posthogServer };
```

Use in API endpoints or SSR pages:

```typescript
import { posthogServer } from '../lib/posthog-server';

posthogServer.capture({
  distinctId: userId,
  event: 'checkout_completed',
  properties: { plan: 'pro', amount: 2999 },
});
```

### Feature Flags

```typescript
// Client-side
const showNewCheckout = posthog.isFeatureEnabled('new-checkout');

// Server-side
const flagValue = await posthogServer.getFeatureFlag('new-checkout', userId);
```

---

## 3. Vercel Analytics + Speed Insights v2

### Setup

```astro
---
// src/layouts/Base.astro
import { Analytics } from '@vercel/analytics/astro';
import { SpeedInsights } from '@vercel/speed-insights/astro';
---
<html>
  <body>
    <slot />
    <Analytics />
    <SpeedInsights />
  </body>
</html>
```

- Zero-config: works automatically on all Vercel deployments
- Free for all Vercel teams (Pro and Hobby)
- Core Web Vitals monitoring included out of the box
- Page view and custom event tracking via `track()` from `@vercel/analytics`

---

## 4. Firebase Option (If Firebase Ecosystem Chosen)

If the project uses Firebase for auth, database, or hosting, use the Firebase
observability stack instead of or alongside the above:

- **Firebase Analytics**: automatic screen tracking, custom events
- **Crashlytics**: crash reporting (primarily mobile, limited web support)
- **Remote Config**: server-driven feature flags and A/B testing

Invoke the following skills for Firebase-specific setup:

- `firebase-analytics-setup`
- `firebase-remote-config`
- `firebase-crashlytics`

---

## 5. Performance Audit Checklist

### Lighthouse CI

```bash
npm install -D @lhci/cli
```

Add to your CI pipeline (e.g., GitHub Actions):

```yaml
- name: Lighthouse CI
  run: |
    npx lhci autorun --config=lighthouserc.json
```

### Bundle Analysis

Use `astro build` output to inspect per-page bundle sizes. Astro reports which
pages include client-side JavaScript and the size of each island. For deeper
analysis, use `source-map-explorer` on the built output in `dist/`.

### Drizzle Query Logging

Enable query logging in development to detect slow queries:

```typescript
import { drizzle } from 'drizzle-orm/neon-http';

const db = drizzle(sql, {
  logger: import.meta.env.DEV,
});
```

Monitor queries exceeding 100ms and add database indexes accordingly.

### Core Web Vitals Targets

| Metric | Target    | Description                    |
| ------ | --------- | ------------------------------ |
| LCP    | < 2.5s    | Largest Contentful Paint       |
| INP    | < 200ms   | Interaction to Next Paint      |
| CLS    | < 0.1     | Cumulative Layout Shift        |

### Astro-Specific Checks

- [ ] Verify static pages ship zero client-side JavaScript (check `astro build` output)
- [ ] Check island hydration overhead -- each `client:load` island adds to first-load JS
- [ ] Prefer `client:visible` and `client:idle` over `client:load` to reduce hydration cost
- [ ] Audit total island count per page -- more than 5 islands suggests refactoring

### Bundle Size Budget

- Per-island JS: < 30KB gzipped
- Total first-load JS (all islands on a page): < 100KB gzipped
- Static pages: 0KB client-side JS
- Enforce with `performance.budgets` in Lighthouse CI config

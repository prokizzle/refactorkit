# Observability Stack Reference -- Premium Next.js on Vercel

This reference covers the full observability setup for a premium Next.js
application deployed on Vercel: error tracking, product analytics, performance
monitoring, and audit tooling.

---

## 1. Sentry Setup (@sentry/nextjs v10)

### Auto-Setup

```bash
npx @sentry/wizard@latest -i nextjs
```

This wizard creates the following files:

- `sentry.client.config.ts` -- browser-side SDK init
- `sentry.server.config.ts` -- Node.js runtime SDK init
- `sentry.edge.config.ts` -- Edge runtime SDK init

### Global Error Boundary (App Router)

Create `app/global-error.tsx` so unhandled errors are captured:

```typescript
'use client';

import * as Sentry from '@sentry/nextjs';
import { useEffect } from 'react';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    Sentry.captureException(error);
  }, [error]);

  return (
    <html>
      <body>
        <h2>Something went wrong</h2>
        <button onClick={() => reset()}>Try again</button>
      </body>
    </html>
  );
}
```

### next.config.ts Wrapper

```typescript
import { withSentryConfig } from '@sentry/nextjs';

const nextConfig = {
  // your existing config
};

export default withSentryConfig(nextConfig, {
  org: process.env.SENTRY_ORG,
  project: process.env.SENTRY_PROJECT,
  authToken: process.env.SENTRY_AUTH_TOKEN,
  silent: !process.env.CI,
  widenClientFileUpload: true,
  tunnelRoute: '/monitoring',
  disableLogger: true,
  automaticVercelMonitors: true,
});
```

### Source Maps

Source maps upload automatically during `next build` when `SENTRY_AUTH_TOKEN` is
set. On Vercel, add the Sentry integration or set the token in environment
variables.

### Session Replay

```typescript
// sentry.client.config.ts
import * as Sentry from '@sentry/nextjs';

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
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

### Skills to Invoke

After initial setup, invoke these skills for deeper integration:

- `sentry-setup-logging` -- structured logging with breadcrumbs
- `sentry-setup-tracing` -- distributed tracing across API routes and DB
- `sentry-setup-metrics` -- custom metrics and dashboards

---

## 2. PostHog Setup (Product Analytics + Feature Flags)

### Client-Side Provider

```typescript
// app/providers.tsx
'use client';

import posthog from 'posthog-js';
import { PostHogProvider as PHProvider } from 'posthog-js/react';
import { useEffect } from 'react';

export function PostHogProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    posthog.init(process.env.NEXT_PUBLIC_POSTHOG_KEY!, {
      api_host: process.env.NEXT_PUBLIC_POSTHOG_HOST ?? 'https://us.i.posthog.com',
      person_profiles: 'identified_only',
      capture_pageview: false, // handled manually for Next.js SPA navigation
      session_recording: {
        recordCrossOriginIframes: true,
      },
    });
  }, []);

  return <PHProvider client={posthog}>{children}</PHProvider>;
}
```

### Server-Side Tracking

```typescript
// lib/posthog-server.ts
import { PostHog } from 'posthog-node';

const posthogServer = new PostHog(process.env.POSTHOG_API_KEY!, {
  host: process.env.POSTHOG_HOST ?? 'https://us.i.posthog.com',
  flushAt: 1,
  flushInterval: 0,
});

export { posthogServer };
```

Use in Server Components and Server Actions:

```typescript
import { posthogServer } from '@/lib/posthog-server';

posthogServer.capture({
  distinctId: userId,
  event: 'checkout_completed',
  properties: { plan: 'pro', amount: 2999 },
});
```

### Feature Flags (A/B Testing)

```typescript
const showNewCheckout = posthog.isFeatureEnabled('new-checkout');
```

Server-side flag evaluation:

```typescript
const flagValue = await posthogServer.getFeatureFlag('new-checkout', userId);
```

### Hosting and Pricing

- **Cloud free tier**: 1M events/mo, 5K session recordings/mo
- **Self-hosted option**: Docker Compose deployment, approximately $5/mo on Railway

---

## 3. Vercel Analytics + Speed Insights v2

### Setup

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/next';
import { SpeedInsights } from '@vercel/speed-insights/next';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
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

These skills handle the different initialization flow required by Firebase.

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

```bash
npm install -D @next/bundle-analyzer
```

Configure per Turborepo package in `next.config.ts`:

```typescript
import withBundleAnalyzer from '@next/bundle-analyzer';

const withAnalyzer = withBundleAnalyzer({
  enabled: process.env.ANALYZE === 'true',
});
```

Run with `ANALYZE=true next build` to generate bundle reports.

### Drizzle Query Logging

Enable query logging in development to detect slow queries:

```typescript
import { drizzle } from 'drizzle-orm/neon-http';

const db = drizzle(sql, {
  logger: process.env.NODE_ENV === 'development',
});
```

Monitor queries exceeding 100ms and add database indexes accordingly.

### Core Web Vitals Targets

| Metric | Target    | Description                    |
| ------ | --------- | ------------------------------ |
| LCP    | < 2.5s    | Largest Contentful Paint       |
| INP    | < 200ms   | Interaction to Next Paint      |
| CLS    | < 0.1     | Cumulative Layout Shift        |

### Bundle Size Budget

- Main chunk: < 100KB gzipped
- Total first-load JS: < 200KB gzipped
- Enforce with `performance.budgets` in Lighthouse CI config

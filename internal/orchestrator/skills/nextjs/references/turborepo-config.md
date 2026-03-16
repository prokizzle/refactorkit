# Turborepo Configuration Reference

Complete templates for a Turborepo-based Next.js monorepo with pnpm workspaces.

---

## turbo.json

```json
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": ["**/.env.*local"],
  "globalEnv": ["NODE_ENV", "VERCEL_URL", "NEXT_PUBLIC_*"],
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["$TURBO_DEFAULT$", ".env*"],
      "outputs": [".next/**", "!.next/cache/**", "dist/**"]
    },
    "dev": { "cache": false, "persistent": true },
    "lint": { "dependsOn": ["^build"] },
    "test": {
      "dependsOn": ["^build"],
      "inputs": ["$TURBO_DEFAULT$", "**/*.test.ts", "**/*.test.tsx"]
    },
    "type-check": { "dependsOn": ["^build"] }
  }
}
```

Filter patterns:

```bash
pnpm turbo build --filter='apps/*'        # Only apps
pnpm turbo type-check --filter='packages/*' # Only packages
pnpm turbo build --filter=web...           # Single package + deps
```

---

## Root package.json

```json
{
  "name": "my-nextjs-app",
  "private": true,
  "packageManager": "pnpm@9.15.4",
  "scripts": {
    "build": "turbo run build",
    "dev": "turbo run dev",
    "lint": "turbo run lint",
    "test": "turbo run test",
    "type-check": "turbo run type-check",
    "format": "biome format --write .",
    "check": "biome check --write ."
  },
  "devDependencies": {
    "turbo": "^2.5.0",
    "typescript": "^5.8.2",
    "@biomejs/biome": "^1.9.4"
  }
}
```

`pnpm-workspace.yaml` at root:

```yaml
packages:
  - "apps/*"
  - "packages/*"
```

---

## Per-Package Configs

### apps/web/package.json

```json
{
  "name": "web",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "build": "next build",
    "dev": "next dev --turbopack",
    "lint": "biome check ./src",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "next": "16.1.6",
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "@clerk/nextjs": "^6.12.0",
    "@sentry/nextjs": "^9.5.0",
    "posthog-js": "^1.222.0",
    "@vercel/analytics": "^1.5.0",
    "@vercel/speed-insights": "^1.2.0",
    "motion": "^12.4.7",
    "nuqs": "^2.4.1",
    "zustand": "^5.0.3",
    "rooks": "^7.14.1",
    "tailwindcss": "^4.0.14",
    "@repo/api": "workspace:*",
    "@repo/db": "workspace:*",
    "@repo/ui": "workspace:*",
    "@repo/billing": "workspace:*",
    "@repo/email": "workspace:*",
    "@repo/notifications": "workspace:*"
  }
}
```

### packages/api/package.json

```json
{
  "name": "@repo/api",
  "version": "0.1.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "scripts": { "type-check": "tsc --noEmit" },
  "dependencies": {
    "@trpc/server": "^11.0.0",
    "@trpc/client": "^11.0.0",
    "@trpc/tanstack-react-query": "^11.0.0",
    "next-safe-action": "^7.10.5",
    "zod": "^3.24.2",
    "inngest": "^3.31.0",
    "trpc-to-openapi": "^2.0.0",
    "@repo/db": "workspace:*"
  }
}
```

### packages/db/package.json

```json
{
  "name": "@repo/db",
  "version": "0.1.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "scripts": {
    "db:push": "drizzle-kit push",
    "db:generate": "drizzle-kit generate",
    "db:migrate": "drizzle-kit migrate",
    "db:studio": "drizzle-kit studio",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "drizzle-orm": "^0.38.4",
    "@neondatabase/serverless": "^0.10.4"
  },
  "devDependencies": { "drizzle-kit": "^0.30.4" }
}
```

### packages/ui/package.json

shadcn components are copied into `packages/ui/src/components/` via `npx shadcn@latest add` -- they are source files, not an npm dependency.

```json
{
  "name": "@repo/ui",
  "version": "0.1.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "scripts": { "type-check": "tsc --noEmit" },
  "dependencies": {
    "tailwindcss": "^4.0.14",
    "class-variance-authority": "^0.7.1",
    "clsx": "^2.1.1",
    "tailwind-merge": "^3.0.2"
  }
}
```

### packages/billing, email, notifications

All follow the same structure (`main`, `types` pointing to `./src/index.ts`, `private: true`).

```json
// packages/billing/package.json
{ "name": "@repo/billing", "dependencies": { "stripe": "^17.7.0" } }

// packages/email/package.json
{ "name": "@repo/email", "dependencies": { "resend": "^4.1.2", "@react-email/components": "^0.0.34" } }

// packages/notifications/package.json  (@novu/node is optional)
{ "name": "@repo/notifications", "dependencies": { "web-push": "^3.7.0", "@novu/node": "^2.1.0" } }
```

---

## biome.json

Replaces ESLint and Prettier with a single tool.

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.4/schema.json",
  "organizeImports": { "enabled": true },
  "formatter": {
    "enabled": true,
    "indentStyle": "tab",
    "lineWidth": 100
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "correctness": {
        "noUnusedImports": "warn",
        "noUnusedVariables": "warn"
      },
      "style": { "noNonNullAssertion": "off" }
    }
  },
  "files": {
    "include": ["**/*.ts", "**/*.tsx", "**/*.json"],
    "ignore": ["node_modules", ".next", "dist", "drizzle"]
  }
}
```

---

## Vercel Deployment

Vercel auto-detects Turborepo monorepos. No `vercel.json` is required for standard setups.

1. Import the repository in the Vercel dashboard.
2. Set the root directory to `apps/web`.
3. Vercel detects `turbo.json` and enables Remote Caching automatically.

### Environment Variable Workflow

Vercel is the source of truth for env vars. Never commit `.env` files.

```bash
vercel link                                    # Link local project
vercel env pull .env.local                     # Pull all env vars
vercel env pull .env.local --environment=preview  # Pull for specific env
```

Reference the shared `.env.local` at repo root via `globalDependencies` in `turbo.json`.

### Build Settings (Vercel Dashboard)

| Setting          | Value                                          |
| ---------------- | ---------------------------------------------- |
| Framework Preset | Next.js                                        |
| Build Command    | `cd ../.. && pnpm turbo build --filter=web...` |
| Output Directory | `.next`                                        |
| Install Command  | `pnpm install`                                 |
| Root Directory   | `apps/web`                                     |

---

## Common Issues

| Problem | Cause | Fix |
| ------- | ----- | --- |
| Build OOM on Vercel | Default memory limit too low | Set `NODE_OPTIONS=--max-old-space-size=8192` in Vercel env vars |
| Env vars undefined at runtime | Not prefixed or not pulled | Client-side vars need `NEXT_PUBLIC_` prefix. Run `vercel env pull`. |
| Package resolution errors | Missing workspace protocol | Use `"@repo/pkg": "workspace:*"` for all internal deps |
| `Module not found` for internal pkg | Package not built first | Verify `dependsOn: ["^build"]` in `turbo.json` |
| Turbo cache misses on CI | Env var differences | Add volatile vars to `globalEnv` in `turbo.json` |
| `pnpm install` fails in CI | Lockfile mismatch | Use `pnpm install --frozen-lockfile` in CI |

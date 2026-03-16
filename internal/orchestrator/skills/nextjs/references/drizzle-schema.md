# Drizzle ORM Reference -- Neon Serverless with Multi-Tenant Schema

## drizzle.config.ts

```typescript
import type { Config } from "drizzle-kit";

export default {
  schema: "./src/db/schema.ts",
  out: "./drizzle/migrations",
  dialect: "postgresql",
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
} satisfies Config;
```

Database client setup using the Neon serverless driver:

```typescript
// src/db/index.ts
import { neon } from "@neondatabase/serverless";
import { drizzle } from "drizzle-orm/neon-http";
import * as schema from "./schema";

const sql = neon(process.env.DATABASE_URL!);
export const db = drizzle(sql, { schema });
```

For connection pooling (long-lived servers or heavy workloads):

```typescript
import { Pool } from "@neondatabase/serverless";
import { drizzle } from "drizzle-orm/neon-serverless";
import * as schema from "./schema";

const pool = new Pool({ connectionString: process.env.DATABASE_URL! });
export const db = drizzle(pool, { schema });
```

---

## Multi-Tenant Schema

Every table includes a `tenantId` column to enforce tenant isolation.

```typescript
// src/db/schema.ts
import {
  pgTable,
  uuid,
  text,
  varchar,
  timestamp,
  pgEnum,
  index,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";

export const roleEnum = pgEnum("user_role", ["owner", "admin", "member"]);
export const planEnum = pgEnum("org_plan", ["free", "pro", "enterprise"]);
export const subscriptionStatusEnum = pgEnum("subscription_status", [
  "active",
  "past_due",
  "canceled",
  "trialing",
  "incomplete",
]);

// --- Organizations (maps to Clerk Organizations) ---
export const organizations = pgTable("organizations", {
  id: uuid("id").defaultRandom().primaryKey(),
  name: varchar("name", { length: 255 }).notNull(),
  slug: varchar("slug", { length: 128 }).notNull().unique(),
  plan: planEnum("plan").default("free").notNull(),
  createdAt: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).defaultNow().notNull(),
});

// --- Users ---
export const users = pgTable(
  "users",
  {
    id: uuid("id").defaultRandom().primaryKey(),
    tenantId: uuid("tenant_id")
      .notNull()
      .references(() => organizations.id, { onDelete: "cascade" }),
    clerkUserId: varchar("clerk_user_id", { length: 255 }).notNull().unique(),
    email: varchar("email", { length: 320 }).notNull(),
    name: varchar("name", { length: 255 }),
    role: roleEnum("role").default("member").notNull(),
    createdAt: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
    updatedAt: timestamp("updated_at", { withTimezone: true }).defaultNow().notNull(),
  },
  (table) => ({
    tenantIdx: index("users_tenant_id_idx").on(table.tenantId),
    clerkIdx: index("users_clerk_user_id_idx").on(table.clerkUserId),
  })
);

// --- Subscriptions ---
export const subscriptions = pgTable(
  "subscriptions",
  {
    id: uuid("id").defaultRandom().primaryKey(),
    tenantId: uuid("tenant_id")
      .notNull()
      .references(() => organizations.id, { onDelete: "cascade" }),
    stripeSubscriptionId: varchar("stripe_subscription_id", { length: 255 })
      .notNull()
      .unique(),
    status: subscriptionStatusEnum("status").default("incomplete").notNull(),
    planId: varchar("plan_id", { length: 128 }).notNull(),
    createdAt: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
    updatedAt: timestamp("updated_at", { withTimezone: true }).defaultNow().notNull(),
  },
  (table) => ({
    tenantIdx: index("subscriptions_tenant_id_idx").on(table.tenantId),
    stripeIdx: index("subscriptions_stripe_id_idx").on(table.stripeSubscriptionId),
  })
);
```

---

## Relationships

```typescript
export const organizationsRelations = relations(organizations, ({ many }) => ({
  users: many(users),
  subscriptions: many(subscriptions),
}));

export const usersRelations = relations(users, ({ one }) => ({
  organization: one(organizations, {
    fields: [users.tenantId],
    references: [organizations.id],
  }),
}));

export const subscriptionsRelations = relations(subscriptions, ({ one }) => ({
  organization: one(organizations, {
    fields: [subscriptions.tenantId],
    references: [organizations.id],
  }),
}));
```

For many-to-many, use a junction table:

```typescript
export const userProjects = pgTable("user_projects", {
  userId: uuid("user_id").notNull().references(() => users.id, { onDelete: "cascade" }),
  projectId: uuid("project_id").notNull().references(() => projects.id, { onDelete: "cascade" }),
  tenantId: uuid("tenant_id").notNull(),
}, (table) => ({
  pk: index("user_projects_pk").on(table.userId, table.projectId),
  tenantIdx: index("user_projects_tenant_idx").on(table.tenantId),
}));
```

---

## Row-Level Security (RLS) for Neon

Apply RLS policies so the database itself enforces tenant isolation:

```sql
-- Enable RLS on each table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;

-- Policy: rows visible only when app.current_tenant matches tenant_id
CREATE POLICY tenant_isolation_users ON users
  USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation_subscriptions ON subscriptions
  USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Force RLS even for table owners (important for serverless)
ALTER TABLE users FORCE ROW LEVEL SECURITY;
ALTER TABLE subscriptions FORCE ROW LEVEL SECURITY;
```

Set tenant context per request in your middleware or data-access layer:

```typescript
import { sql } from "drizzle-orm";

export async function withTenant<T>(
  tenantId: string,
  callback: () => Promise<T>
): Promise<T> {
  await db.execute(sql`SET app.current_tenant = ${tenantId}`);
  try {
    return await callback();
  } finally {
    await db.execute(sql`RESET app.current_tenant`);
  }
}
```

---

## Migration Commands

```bash
# Generate migration SQL from schema changes
npx drizzle-kit generate

# Apply pending migrations to the database
npx drizzle-kit migrate

# Open visual database browser (local dev)
npx drizzle-kit studio
```

---

## Query Patterns

```typescript
import { eq, and } from "drizzle-orm";
import { db } from "./db";
import { users, organizations, subscriptions } from "./db/schema";

// SELECT -- fetch user by Clerk ID
const user = await db.query.users.findFirst({
  where: eq(users.clerkUserId, "user_abc123"),
  with: { organization: true },
});

// SELECT with multi-condition filter
const tenantUsers = await db
  .select()
  .from(users)
  .where(and(eq(users.tenantId, tenantId), eq(users.role, "admin")));

// INSERT
const [newUser] = await db
  .insert(users)
  .values({
    tenantId,
    clerkUserId: "user_xyz",
    email: "dev@example.com",
    name: "Dev User",
    role: "member",
  })
  .returning();

// UPDATE
await db
  .update(users)
  .set({ role: "admin", updatedAt: new Date() })
  .where(eq(users.id, userId));

// DELETE
await db.delete(users).where(eq(users.id, userId));
```

### Transactions

```typescript
const result = await db.transaction(async (tx) => {
  const [org] = await tx
    .insert(organizations)
    .values({ name: "Acme", slug: "acme", plan: "pro" })
    .returning();

  const [user] = await tx
    .insert(users)
    .values({
      tenantId: org.id,
      clerkUserId: "user_owner",
      email: "owner@acme.com",
      name: "Owner",
      role: "owner",
    })
    .returning();

  return { org, user };
});
```

---

## Supabase Adapter Variant

Swap the Neon driver for Supabase when using the Supabase ecosystem:

```typescript
// src/db/index.ts (Supabase variant)
import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import * as schema from "./schema";

const connectionString = process.env.DATABASE_URL!;
const client = postgres(connectionString, { prepare: false });
export const db = drizzle(client, { schema });
```

The `drizzle.config.ts` and schema files remain unchanged. Only the client initialization differs. Set `prepare: false` when connecting through Supabase's connection pooler (PgBouncer in transaction mode).

---

## Common Issues

**Migration conflicts** -- When multiple developers generate migrations concurrently, conflicting SQL files appear in `drizzle/migrations`. Delete the conflicting file, merge schema changes, and run `npx drizzle-kit generate` again from the merged schema.

**Type mismatches** -- Drizzle infers TypeScript types from schema definitions. If you cast `uuid` columns to `text` or vice versa in raw SQL, the type system cannot catch errors. Always use the column references from the schema object.

**Connection pooling with Neon serverless** -- The `neon()` HTTP driver creates a new connection per query, which is ideal for serverless functions. For long-lived processes (dev servers, cron jobs), use the `Pool` driver instead to avoid connection overhead. Never mix both drivers in the same runtime.

**Index creation** -- Always add indexes on `tenantId` columns. Queries without an index on the tenant column will trigger full table scans, defeating the purpose of RLS and degrading performance at scale. Define indexes inline in the table definition as shown above.

**Neon cold starts** -- First query after idle may take 300-500ms due to compute spin-up. Use Neon's `autoscaling` or keep-alive pings in production to mitigate latency spikes.

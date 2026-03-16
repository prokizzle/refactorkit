# Service Layer Pattern: Dual API (tRPC + Astro Actions)

## Directory Structure

```
packages/api/src/
├── services/           # Framework-agnostic business logic
│   └── users.ts        # Accepts tenantId, returns typed results
├── trpc/               # tRPC router -- mobile/external API contract
│   ├── router.ts
│   └── routes/users.ts
├── schemas/            # Zod schemas shared across tRPC and Astro Actions
│   └── users.ts
└── events/
    └── schemas.ts      # Typed Inngest event schemas
```

Services hold all business logic. tRPC routes and Astro Actions are thin wrappers.
## Shared Zod Schemas

```typescript
// packages/api/src/schemas/users.ts
import { z } from "zod";

export const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(255),
  role: z.enum(["member", "admin"]).default("member"),
});
export const updateUserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1).max(255).optional(),
  role: z.enum(["member", "admin"]).optional(),
});
export const getUserSchema = z.object({ id: z.string().uuid() });
export type CreateUserInput = z.infer<typeof createUserSchema>;
export type UpdateUserInput = z.infer<typeof updateUserSchema>;
```

## Service Layer

```typescript
// packages/api/src/services/users.ts
import { eq, and } from "drizzle-orm";
import { db } from "@/db";
import { users } from "@/db/schema";
import type { CreateUserInput, UpdateUserInput } from "../schemas/users";

export async function createUser(tenantId: string, input: CreateUserInput) {
  const [user] = await db.insert(users).values({ ...input, tenantId }).returning();
  return user;
}
export async function getUserById(tenantId: string, id: string) {
  const user = await db.query.users.findFirst({
    where: and(eq(users.id, id), eq(users.tenantId, tenantId)),
  });
  if (!user) throw new Error("User not found");
  return user;
}
export async function updateUser(tenantId: string, input: UpdateUserInput) {
  const [user] = await db.update(users)
    .set({ ...input, updatedAt: new Date() })
    .where(and(eq(users.id, input.id), eq(users.tenantId, tenantId)))
    .returning();
  if (!user) throw new Error("User not found");
  return user;
}
export async function listUsers(tenantId: string) {
  return db.query.users.findMany({
    where: eq(users.tenantId, tenantId),
    orderBy: (users, { desc }) => [desc(users.createdAt)],
  });
}
```

## Inngest Event Schemas

```typescript
// packages/api/src/events/schemas.ts
import { Inngest, EventSchemas } from "inngest";
export type Events = {
  "user.created": { data: { tenantId: string; userId: string; email: string } };
  "user.updated": { data: { tenantId: string; userId: string; changes: string[] } };
};
export const inngest = new Inngest({
  id: "my-app",
  schemas: new EventSchemas().fromRecord<Events>(),
});
```

## tRPC v11 Setup

The tRPC router and route definitions are identical to the Next.js service-layer
pattern. The only Astro-specific part is the API route handler.

```typescript
// packages/api/src/trpc/router.ts
import { initTRPC, TRPCError } from "@trpc/server";
import superjson from "superjson";
import { userRouter } from "./routes/users";

export type Context = { userId: string; tenantId: string };
const t = initTRPC.context<Context>().create({ transformer: superjson });
const isAuthed = t.middleware(async ({ ctx, next }) => {
  if (!ctx.userId) throw new TRPCError({ code: "UNAUTHORIZED" });
  return next({ ctx });
});
export const protectedProcedure = t.procedure.use(isAuthed);
export const appRouter = t.router({ user: userRouter });
export type AppRouter = typeof appRouter;
```

tRPC routes wrap services the same way -- see the Next.js reference for the full
`userRouter` with `list`, `getById`, `create`, and `update` procedures.

### tRPC API Route for Astro

```typescript
// src/pages/api/trpc/[...trpc].ts
import type { APIRoute } from "astro";
import { fetchRequestHandler } from "@trpc/server/adapters/fetch";
import { appRouter } from "@repo/api/trpc/router";

export const ALL: APIRoute = async ({ request }) => {
  return fetchRequestHandler({
    endpoint: "/api/trpc",
    req: request,
    router: appRouter,
    createContext: () => ({}),
  });
};
```

### tRPC Client

```typescript
// src/lib/trpc.ts
import { createTRPCClient, httpBatchLink } from "@trpc/client";
import superjson from "superjson";
import type { AppRouter } from "@repo/api/trpc/router";
export const trpc = createTRPCClient<AppRouter>({
  links: [httpBatchLink({ url: "/api/trpc", transformer: superjson })],
});
```

## Astro Actions (wrapping the same service)

Astro Actions replace Next.js Server Actions as the web-only fast-path.

```typescript
// src/actions/index.ts
import { defineAction } from "astro:actions";
import { z } from "astro:schema";
import { createUser, updateUser } from "@repo/api/services/users";
import { inngest } from "@repo/api/events/schemas";

export const server = {
  createUser: defineAction({
    accept: "form",
    input: z.object({
      name: z.string(),
      email: z.string().email(),
    }),
    handler: async (input, context) => {
      const tenantId = context.locals.tenantId;
      const user = await createUser(tenantId, input);
      await inngest.send({
        name: "user.created",
        data: { tenantId, userId: user.id, email: user.email },
      });
      return user;
    },
  }),

  updateUser: defineAction({
    accept: "json",
    input: z.object({
      id: z.string().uuid(),
      name: z.string().min(1).max(255).optional(),
      role: z.enum(["member", "admin"]).optional(),
    }),
    handler: async (input, context) => {
      const tenantId = context.locals.tenantId;
      const user = await updateUser(tenantId, input);
      await inngest.send({
        name: "user.updated",
        data: { tenantId, userId: user.id, changes: Object.keys(input).filter((k) => k !== "id") },
      });
      return user;
    },
  }),
};
```

### Calling from an Astro Component

```astro
---
// src/pages/users/new.astro
import { actions } from "astro:actions";
if (Astro.request.method === "POST") {
  const formData = await Astro.request.formData();
  const { data, error } = await Astro.callAction(actions.createUser, formData);
}
---
<form method="POST">
  <input name="name" required />
  <input name="email" type="email" required />
  <button type="submit">Create User</button>
</form>
```

### Calling from a React Island

```tsx
// src/components/CreateUserForm.tsx
import { actions } from "astro:actions";

export default function CreateUserForm() {
  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const { data, error } = await actions.createUser(new FormData(e.currentTarget));
    if (error) return;
  }
  return (
    <form onSubmit={handleSubmit}>
      <input name="name" required />
      <input name="email" type="email" required />
      <button type="submit">Create User</button>
    </form>
  );
}
```

## OpenAPI Spec Generation from tRPC

Identical to the Next.js pattern. Use `trpc-to-openapi` with `.meta()` on tRPC
routes, then serve the generated document at `src/pages/api/openapi.json.ts`.
## CLAUDE.md Agent Rules Template

```markdown
## Service Layer Rules
- Business logic goes in `packages/api/src/services/`. Astro Actions and tRPC routes are thin wrappers.
- Never put logic directly in an Astro Action without a corresponding tRPC route.
- When adding a feature: update service -> update tRPC route -> update Astro Action -> update API docs.
- Every mutation must emit an Inngest event for downstream processing.
- Zod schemas live in `packages/api/src/schemas/` and are shared across tRPC and Astro Actions.
- Services accept `tenantId` as the first parameter -- never read auth context inside a service.
- tRPC routes handle auth via middleware context. Astro Actions read auth from `context.locals`.
```

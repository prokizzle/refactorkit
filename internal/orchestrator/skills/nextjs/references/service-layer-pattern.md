# Service Layer Pattern: Dual API (tRPC + Server Actions)

## Directory Structure

```
packages/api/src/
├── services/           # Framework-agnostic business logic
│   └── users.ts        # Accepts tenantId, returns typed results
├── trpc/               # tRPC router -- mobile/external API contract
│   ├── router.ts
│   └── routes/users.ts
├── actions/            # Server Actions -- web-only fast-path
│   └── user-actions.ts
├── schemas/            # Zod schemas shared across tRPC and actions
│   └── users.ts
└── events/
    └── schemas.ts      # Typed Inngest event schemas
```

Services hold all business logic. tRPC routes and Server Actions are thin wrappers.

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
  const [user] = await db
    .insert(users)
    .values({ ...input, tenantId })
    .returning();
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
  const [user] = await db
    .update(users)
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

### tRPC Route (wrapping the service)

```typescript
// packages/api/src/trpc/routes/users.ts
import { router } from "../router";
import { protectedProcedure } from "../router";
import { createUserSchema, updateUserSchema, getUserSchema } from "../../schemas/users";
import * as userService from "../../services/users";
import { inngest } from "../../events/schemas";

export const userRouter = router({
  list: protectedProcedure.query(async ({ ctx }) => {
    return userService.listUsers(ctx.tenantId);
  }),
  getById: protectedProcedure
    .input(getUserSchema)
    .query(async ({ ctx, input }) => {
      return userService.getUserById(ctx.tenantId, input.id);
    }),
  create: protectedProcedure
    .input(createUserSchema)
    .mutation(async ({ ctx, input }) => {
      const user = await userService.createUser(ctx.tenantId, input);
      await inngest.send({
        name: "user.created",
        data: { tenantId: ctx.tenantId, userId: user.id, email: user.email },
      });
      return user;
    }),
  update: protectedProcedure
    .input(updateUserSchema)
    .mutation(async ({ ctx, input }) => {
      const user = await userService.updateUser(ctx.tenantId, input);
      await inngest.send({
        name: "user.updated",
        data: { tenantId: ctx.tenantId, userId: user.id, changes: Object.keys(input).filter((k) => k !== "id") },
      });
      return user;
    }),
});
```

### TanStack Query Integration + HydrateClient

```typescript
// packages/api/src/trpc/client.ts
import { createTRPCClient, httpBatchLink } from "@trpc/client";
import { createServerSideHelpers } from "@trpc/react-query/server";
import superjson from "superjson";
import type { AppRouter } from "./router";

export const trpc = createTRPCClient<AppRouter>({
  links: [httpBatchLink({ url: "/api/trpc", transformer: superjson })],
});

export const createSSRHelpers = (ctx: Context) =>
  createServerSideHelpers<AppRouter>({ router: appRouter, ctx, transformer: superjson });
```

```tsx
// app/users/page.tsx -- Server Component with prefetch
import { HydrateClient } from "@trpc/react-query/rsc";
import { createSSRHelpers } from "@/api/trpc/client";

export default async function UsersPage() {
  const helpers = await createSSRHelpers(await getServerContext());
  await helpers.user.list.prefetch();
  return (
    <HydrateClient>
      <UserList />
    </HydrateClient>
  );
}
```

## Server Actions (wrapping the same service)

```typescript
// packages/api/src/actions/user-actions.ts
"use server";

import { createSafeActionClient } from "next-safe-action";
import { auth } from "@clerk/nextjs/server";
import { createUserSchema, updateUserSchema } from "../schemas/users";
import * as userService from "../services/users";
import { inngest } from "../events/schemas";

const authAction = createSafeActionClient()
  .use(async ({ next }) => {
    const { userId, orgId } = await auth();
    if (!userId || !orgId) throw new Error("Unauthorized");
    return next({ ctx: { userId, tenantId: orgId } });
  })
  .use(async ({ next, ctx }) => {
    // Rate limiting middleware (e.g., Upstash ratelimit)
    return next({ ctx });
  });

export const createUserAction = authAction
  .schema(createUserSchema)
  .action(async ({ parsedInput, ctx }) => {
    const user = await userService.createUser(ctx.tenantId, parsedInput);
    await inngest.send({
      name: "user.created",
      data: { tenantId: ctx.tenantId, userId: user.id, email: user.email },
    });
    return user;
  });

export const updateUserAction = authAction
  .schema(updateUserSchema)
  .action(async ({ parsedInput, ctx }) => {
    const user = await userService.updateUser(ctx.tenantId, parsedInput);
    await inngest.send({
      name: "user.updated",
      data: { tenantId: ctx.tenantId, userId: user.id, changes: Object.keys(parsedInput).filter((k) => k !== "id") },
    });
    return user;
  });
```

## OpenAPI Spec Generation from tRPC

```typescript
// packages/api/src/trpc/openapi.ts
import { generateOpenApiDocument } from "trpc-to-openapi";
import { appRouter } from "./router";

export const openApiDocument = generateOpenApiDocument(appRouter, {
  title: "App API", version: "1.0.0", baseUrl: "https://api.example.com",
});
// Serve at GET /api/openapi.json via a Next.js route handler
```

Add `.meta()` to tRPC routes for OpenAPI exposure:

```typescript
getById: protectedProcedure
  .meta({ openapi: { method: "GET", path: "/users/{id}" } })
  .input(getUserSchema)
  .query(async ({ ctx, input }) => userService.getUserById(ctx.tenantId, input.id)),
```

## CLAUDE.md Agent Rules Template

```markdown
## Service Layer Rules
- Business logic goes in `packages/api/src/services/`. Server Actions and tRPC routes are thin wrappers.
- Never put logic directly in a Server Action without a corresponding tRPC route.
- When adding a feature: update service -> update tRPC route -> update Server Action -> update API docs.
- Every mutation must emit an Inngest event for downstream processing.
- Zod schemas live in `packages/api/src/schemas/` and are shared across tRPC and Server Actions.
- Services accept `tenantId` as the first parameter -- never read auth context inside a service.
- tRPC routes handle auth via middleware context. Server Actions handle auth via next-safe-action middleware.
```

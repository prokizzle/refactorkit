# Realtime Updates, Push Notifications, and Cache Invalidation

Reference for premium Next.js apps. Covers realtime transport selection,
zero-vendor web push, unified notification infrastructure, cache invalidation
wired to an event bus, SSE streaming, and self-hosted alternatives.

---

## 1. Realtime Strategy Decision Matrix

| Path | Realtime Solution | Push (Web+Mobile) | Self-Hosted Option |
|------|------------------|-------------------|-------------------|
| Composable | Ably LiveSync (Neon partnership) | Novu (open-source) | Soketi + Novu Docker |
| Supabase | Supabase Realtime (Broadcast) | Novu | Supabase self-host + Novu |
| Firebase | Firestore/Realtime DB | FCM (built-in) | N/A |
| Collaborative | PartyKit (Cloudflare) | OneSignal or Novu | Soketi + custom |

Pick one realtime transport per project. Do not mix WebSocket providers.

---

## 2. Web Push Setup (Zero Vendor, Free)

Uses the `web-push` npm package with VAPID keys. No third-party service required.

### Generate VAPID Keys

```bash
npx web-push generate-vapid-keys
```

Store `NEXT_PUBLIC_VAPID_PUBLIC_KEY` and `VAPID_PRIVATE_KEY` in `.env.local`.

### Service Worker (`public/sw.js`)

```typescript
// public/sw.js
self.addEventListener("push", (event) => {
  const data = event.data?.json() ?? { title: "Notification", body: "" };
  event.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body,
      icon: data.icon ?? "/icon-192.png",
      data: { url: data.url ?? "/" },
    })
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const url = event.notification.data?.url ?? "/";
  event.waitUntil(clients.openWindow(url));
});
```

### Client-Side Subscription

```typescript
// lib/push/subscribe.ts
export async function subscribeToPush(): Promise<PushSubscription | null> {
  if (!("serviceWorker" in navigator) || !("PushManager" in window)) {
    return null;
  }
  const registration = await navigator.serviceWorker.register("/sw.js");
  const subscription = await registration.pushManager.subscribe({
    userAgentAuthentication: true,
    applicationServerKey: process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY,
  });
  // Persist subscription to database via Server Action
  await saveSubscription(subscription.toJSON());
  return subscription;
}
```

### Server Action to Send Push

```typescript
// app/actions/send-push.ts
"use server";

import webpush from "web-push";

webpush.setVapidDetails(
  "mailto:team@example.com",
  process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY!,
  process.env.VAPID_PRIVATE_KEY!
);

export async function sendPushNotification(
  subscription: webpush.PushSubscription,
  payload: { title: string; body: string; url?: string }
) {
  await webpush.sendNotification(subscription, JSON.stringify(payload));
}
```

---

## 3. Novu Unified Notifications (Recommended for Multi-Channel)

MIT-licensed, self-hostable via Docker. Provides in-app, email, push
(APNs/FCM), SMS, and Slack from a single workflow definition.

### Drop-In Inbox Component

```typescript
// app/components/notification-inbox.tsx
import { Inbox } from "@novu/nextjs";

export function NotificationInbox() {
  return (
    <Inbox
      applicationIdentifier={process.env.NEXT_PUBLIC_NOVU_APP_ID!}
      subscriberId={currentUser.id}
    />
  );
}
```

### Code-First Workflow Definition

```typescript
// packages/notifications/workflows/order-shipped.ts
import { workflow } from "@novu/framework";

export const orderShipped = workflow("order-shipped", async ({ step, payload }) => {
  await step.inApp("in-app-alert", () => ({
    body: `Order ${payload.orderId} has shipped.`,
  }));

  await step.email("email-notification", () => ({
    subject: `Your order ${payload.orderId} shipped`,
    body: renderShippedEmail(payload),
  }));

  await step.push("mobile-push", () => ({
    title: "Order Shipped",
    body: `Order ${payload.orderId} is on its way.`,
  }));
});
```

### Provider-Agnostic Interface

```typescript
// packages/notifications/index.ts
export interface NotificationSender {
  send(userId: string, event: string, payload: Record<string, unknown>): Promise<void>;
}

export class NovuNotificationSender implements NotificationSender {
  async send(userId: string, event: string, payload: Record<string, unknown>) {
    await novu.trigger(event, { to: { subscriberId: userId }, payload });
  }
}
```

Place this in `packages/notifications/` so the rest of the app never imports
Novu directly.

---

## 4. Cache Invalidation Strategy

Wire cache invalidation to the event bus from day one. Never invalidate inline
inside mutations.

| Mechanism | When to Use |
|-----------|------------|
| `"use cache"` + `cacheTag()` | Default caching for all server data fetching |
| `revalidateTag(tag)` | Lazy invalidation from Inngest event handlers, Route Handlers, webhooks |
| `updateTag(tag)` | Immediate read-your-own-writes -- Server Actions only |
| `revalidatePath(path)` | Invalidate everything on a page (use sparingly) |
| On-demand ISR webhook | External events (CMS publish) -> Route Handler -> revalidateTag() |

### Mutation -> Event Bus -> Invalidation + Realtime + Push

```typescript
// app/actions/publish-post.ts
"use server";

import { updateTag } from "next/cache";
import { inngest } from "@/lib/inngest";

export async function publishPost(postId: string) {
  const post = await db.post.update({
    where: { id: postId },
    data: { status: "published", publishedAt: new Date() },
  });

  // Immediate read-your-own-writes for the acting user
  updateTag(`post-${postId}`);

  // Fire event -- all side effects happen in the handler
  await inngest.send({
    name: "post/published",
    data: { postId, authorId: post.authorId, slug: post.slug },
  });
}
```

```typescript
// inngest/functions/on-post-published.ts
import { revalidateTag } from "next/cache";
import { inngest } from "@/lib/inngest";
import { broadcastRealtime } from "@/lib/realtime";
import { sendPushToFollowers } from "@/lib/push";

export const onPostPublished = inngest.createFunction(
  { id: "on-post-published" },
  { event: "post/published" },
  async ({ event }) => {
    const { postId, authorId, slug } = event.data;

    // 1. Invalidate caches
    revalidateTag(`post-${postId}`);
    revalidateTag(`feed`);
    revalidateTag(`author-${authorId}-posts`);

    // 2. Broadcast realtime update to connected clients
    await broadcastRealtime(`feed`, {
      type: "NEW_POST",
      postId,
      slug,
    });

    // 3. Send push notifications to followers
    await sendPushToFollowers(authorId, {
      title: "New Post",
      body: `New post published`,
      url: `/posts/${slug}`,
    });
  }
);
```

### Cached Data Fetching with Tags

```typescript
// app/posts/[slug]/page.tsx
import { cacheTag } from "next/cache";

async function getPost(slug: string) {
  "use cache";
  cacheTag(`post-${slug}`);
  return db.post.findUnique({ where: { slug } });
}
```

---

## 5. Agent Rule

> Every `"use cache"` entry MUST have a `cacheTag()`. Every mutation MUST
> `revalidateTag()` via the event bus. Never invalidate cache inline in
> mutations.

---

## 6. SSE for Lightweight Streaming

Use Server-Sent Events for LLM token streaming, progress bars, or any
unidirectional server-to-client flow.

### Route Handler with ReadableStream

```typescript
// app/api/stream/route.ts
export const runtime = "nodejs"; // or "edge" for no-timeout SSE

export async function GET() {
  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder();

      for await (const token of generateTokens()) {
        controller.enqueue(encoder.encode(`data: ${JSON.stringify({ token })}\n\n`));
      }

      controller.enqueue(encoder.encode("data: [DONE]\n\n"));
      controller.close();
    },
  });

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
```

### Vercel Timeout Limits

| Plan | Max Duration |
|------|-------------|
| Hobby | 10 seconds |
| Pro | 60 seconds |
| Enterprise | 300 seconds |

Use `export const runtime = "edge"` for long-lived SSE connections that exceed
the Node.js function timeout. Edge functions have no execution time limit on
Vercel but do have CPU time limits.

---

## 7. Soketi (Self-Hosted Pusher Alternative)

Pusher Protocol v7 compatible. Swap the host and port in your Pusher client
config -- the SDK and API remain identical.

### Why Soketi

- A $5 VPS gives unlimited connections vs the $49 Pusher Starter plan (500
  connection cap).
- Built on uWebSockets.js: benchmarked at 8.5x faster than Fastify for
  WebSocket handling.
- Drop-in replacement: same `pusher-js` client SDK, same `pusher` server SDK.

### Configuration Swap

```typescript
// lib/realtime/pusher.ts
import PusherServer from "pusher";
import PusherClient from "pusher-js";

const isSelftHosted = process.env.REALTIME_PROVIDER === "soketi";

export const pusherServer = new PusherServer({
  appId: process.env.PUSHER_APP_ID!,
  key: process.env.NEXT_PUBLIC_PUSHER_KEY!,
  secret: process.env.PUSHER_SECRET!,
  ...(isSelftHosted
    ? { host: process.env.SOKETI_HOST!, port: process.env.SOKETI_PORT!, useTLS: false }
    : { cluster: process.env.PUSHER_CLUSTER! }),
});

export const pusherClient = new PusherClient(process.env.NEXT_PUBLIC_PUSHER_KEY!, {
  ...(isSelftHosted
    ? { wsHost: process.env.NEXT_PUBLIC_SOKETI_HOST!, wsPort: Number(process.env.NEXT_PUBLIC_SOKETI_PORT!), forceTLS: false, disableStats: true, enabledTransports: ["ws", "wss"] }
    : { cluster: process.env.NEXT_PUBLIC_PUSHER_CLUSTER! }),
});
```

### Broadcast Helper

```typescript
// lib/realtime/broadcast.ts
import { pusherServer } from "./pusher";

export async function broadcastRealtime(
  channel: string,
  data: Record<string, unknown>
) {
  await pusherServer.trigger(channel, "update", data);
}
```

This helper is the single call site for all realtime broadcasts, used by
Inngest event handlers as shown in section 4.

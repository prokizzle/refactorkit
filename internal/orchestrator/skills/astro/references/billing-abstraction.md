# Billing Abstraction Layer

Provider-agnostic billing for a Turborepo monorepo. Lives in `packages/billing/`.

---

## 1. Provider-Agnostic Interface

**`packages/billing/src/types.ts`**

```typescript
export interface Plan {
  id: string;
  name: string;
  description: string;
  priceInCents: number;
  currency: string;
  interval: "month" | "year";
  features: string[];
  metadata?: Record<string, string>;
}

export interface Subscription {
  id: string;
  planId: string;
  status: "active" | "canceled" | "past_due" | "trialing" | "paused";
  currentPeriodStart: Date;
  currentPeriodEnd: Date;
  cancelAtPeriodEnd: boolean;
  customerId: string;
  metadata?: Record<string, string>;
}

export interface CheckoutParams {
  planId: string;
  customerId: string;
  successUrl: string;
  cancelUrl: string;
  metadata?: Record<string, string>;
}

export interface CheckoutSession {
  id: string;
  url: string | null;
  clientSecret: string | null;
  status: "open" | "complete" | "expired";
}

export interface SubscriptionParams {
  customerId: string;
  planId: string;
  trialDays?: number;
  metadata?: Record<string, string>;
}

export interface WebhookEvent {
  type: string;
  data: Record<string, unknown>;
}

export interface BillingProvider {
  createCheckoutSession(params: CheckoutParams): Promise<CheckoutSession>;
  createSubscription(params: SubscriptionParams): Promise<Subscription>;
  cancelSubscription(subscriptionId: string): Promise<void>;
  getSubscription(subscriptionId: string): Promise<Subscription>;
  listPlans(): Promise<Plan[]>;
  handleWebhook(payload: string, signature: string): Promise<WebhookEvent>;
}
```

---

## 2. Stripe Adapter

**`packages/billing/src/adapters/stripe.ts`**

Uses Stripe Node SDK v17 with Embedded Checkout (stays on-domain, no redirect).

```typescript
import Stripe from "stripe";
import type {
  BillingProvider, CheckoutParams, CheckoutSession,
  SubscriptionParams, Subscription, Plan, WebhookEvent,
} from "../types";

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: "2024-12-18.acacia",
  typescript: true,
});

export class StripeAdapter implements BillingProvider {
  async createCheckoutSession(params: CheckoutParams): Promise<CheckoutSession> {
    // Embedded Checkout -- returns clientSecret, no redirect URL
    const session = await stripe.checkout.sessions.create({
      mode: "subscription",
      line_items: [{ price: params.planId, quantity: 1 }],
      customer: params.customerId,
      ui_mode: "embedded",
      return_url: params.successUrl + "?session_id={CHECKOUT_SESSION_ID}",
      metadata: params.metadata,
    });
    return {
      id: session.id,
      url: null,
      clientSecret: session.client_secret,
      status: session.status === "open" ? "open" : "complete",
    };
  }

  async createSubscription(params: SubscriptionParams): Promise<Subscription> {
    const sub = await stripe.subscriptions.create({
      customer: params.customerId,
      items: [{ price: params.planId }],
      trial_period_days: params.trialDays,
      metadata: params.metadata,
    });
    return this.mapSubscription(sub);
  }

  async cancelSubscription(subscriptionId: string): Promise<void> {
    await stripe.subscriptions.update(subscriptionId, {
      cancel_at_period_end: true,
    });
  }

  async resumeSubscription(subscriptionId: string): Promise<Subscription> {
    const sub = await stripe.subscriptions.update(subscriptionId, {
      cancel_at_period_end: false,
    });
    return this.mapSubscription(sub);
  }

  async getSubscription(subscriptionId: string): Promise<Subscription> {
    const sub = await stripe.subscriptions.retrieve(subscriptionId);
    return this.mapSubscription(sub);
  }

  async listPlans(): Promise<Plan[]> {
    const prices = await stripe.prices.list({
      active: true,
      expand: ["data.product"],
    });
    return prices.data.map((price) => {
      const product = price.product as Stripe.Product;
      return {
        id: price.id,
        name: product.name,
        description: product.description ?? "",
        priceInCents: price.unit_amount ?? 0,
        currency: price.currency,
        interval: price.recurring?.interval === "year" ? "year" : "month",
        features: product.features?.map((f) => f.name ?? "") ?? [],
        metadata: product.metadata,
      };
    });
  }

  async handleWebhook(payload: string, signature: string): Promise<WebhookEvent> {
    // Use in a Next.js Route Handler (app/api/webhooks/stripe/route.ts)
    const event = stripe.webhooks.constructEvent(
      payload,
      signature,
      process.env.STRIPE_WEBHOOK_SECRET!,
    );
    return { type: event.type, data: event.data.object as Record<string, unknown> };
  }

  private mapSubscription(sub: Stripe.Subscription): Subscription {
    return {
      id: sub.id,
      planId: sub.items.data[0]?.price.id ?? "",
      status: sub.status as Subscription["status"],
      currentPeriodStart: new Date(sub.current_period_start * 1000),
      currentPeriodEnd: new Date(sub.current_period_end * 1000),
      cancelAtPeriodEnd: sub.cancel_at_period_end,
      customerId: sub.customer as string,
      metadata: sub.metadata,
    };
  }
}
```

**Embedded Checkout client component (`apps/web/components/checkout.tsx`):**

```typescript
"use client";
import { loadStripe } from "@stripe/stripe-js";
import { EmbeddedCheckoutProvider, EmbeddedCheckout } from "@stripe/react-stripe-js";

const stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);

export function CheckoutForm({ clientSecret }: { clientSecret: string }) {
  return (
    <EmbeddedCheckoutProvider stripe={stripePromise} options={{ clientSecret }}>
      <EmbeddedCheckout />
    </EmbeddedCheckoutProvider>
  );
}
```

---

## 3. Clerk Billing Adapter

**`packages/billing/src/adapters/clerk.ts`**

Clerk Billing wraps Stripe with a managed UI and simplified API.

```typescript
// Server-side: use Clerk's auth() to check plan access.
// No SDK adapter needed -- Clerk handles checkout via <PricingTable />.

// Client: render pricing table in any page
// import { PricingTable } from "@clerk/nextjs";
// <PricingTable />

// Feature gating with has():
import { auth } from "@clerk/nextjs/server";

export async function checkFeatureAccess(permission: string): Promise<boolean> {
  const { has } = await auth();
  return has({ permission });
}
```

**Clerk Billing limitations (document for team awareness):**

- USD only -- no multi-currency support
- Beta status -- API surface may change
- No tax/VAT calculation built in
- No metered or usage-based billing
- 0.7% platform fee on top of Stripe processing fees
- Limited webhook events compared to Stripe direct

**When to switch from Clerk Billing to Stripe direct:**

- Multi-currency or international tax requirements
- Metered/usage-based pricing models
- Volume where 0.7% fee exceeds cost of direct integration
- Need for Stripe Billing Portal, invoicing, or revenue recognition
- Complex coupon/promotion logic

---

## 4. Feature Gating Pattern

### Middleware (organization-scoped)

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";

const isPremiumRoute = createRouteMatcher(["/dashboard/analytics(.*)"]);

export default clerkMiddleware(async (auth, req) => {
  if (isPremiumRoute(req)) {
    const { has, orgId } = await auth();
    if (!orgId) {
      return Response.redirect(new URL("/select-org", req.url));
    }
    if (!has({ permission: "premium" })) {
      return Response.redirect(new URL("/upgrade", req.url));
    }
  }
});
```

### Service layer (plan-aware data access)

```typescript
// packages/billing/src/gating.ts
type PlanTier = "free" | "pro" | "enterprise";

export function getDataLimit(tier: PlanTier): number {
  const limits: Record<PlanTier, number> = {
    free: 100,
    pro: 10_000,
    enterprise: Infinity,
  };
  return limits[tier];
}

export function filterByTier<T>(items: T[], tier: PlanTier): T[] {
  const limit = getDataLimit(tier);
  return items.slice(0, limit);
}
```

### Client-side conditional rendering

```typescript
// Use Clerk's <Protect> component to show/hide UI based on permissions
import { Protect } from "@clerk/nextjs";

export function AnalyticsCard() {
  return (
    <Protect permission="premium" fallback={<UpgradeBanner />}>
      <PremiumAnalytics />
    </Protect>
  );
}
```

---

## 5. Webhook to Inngest Event Bus

**Route Handler (`apps/web/app/api/webhooks/stripe/route.ts`):**

```typescript
import { NextRequest } from "next/server";
import { StripeAdapter } from "@repo/billing/adapters/stripe";
import { inngest } from "@/lib/inngest";

const billing = new StripeAdapter();

export async function POST(req: NextRequest) {
  const payload = await req.text();
  const signature = req.headers.get("stripe-signature")!;

  const event = await billing.handleWebhook(payload, signature);

  // Fan out to Inngest for async processing
  await inngest.send({
    name: `billing.${event.type.replaceAll(".", "_")}`,
    data: event.data,
  });

  return new Response("OK", { status: 200 });
}
```

**Inngest handler (`apps/web/inngest/billing.ts`):**

```typescript
import { inngest } from "@/lib/inngest";
import { db } from "@repo/db";
import { revalidateTag } from "next/cache";

export const handleSubscriptionUpdated = inngest.createFunction(
  { id: "billing-subscription-updated" },
  { event: "billing.customer_subscription_updated" },
  async ({ event }) => {
    const { id, status, customer } = event.data as {
      id: string;
      status: string;
      customer: string;
    };

    // 1. Update DB
    await db.subscription.upsert({
      where: { stripeSubscriptionId: id },
      update: { status },
      create: { stripeSubscriptionId: id, stripeCustomerId: customer, status },
    });

    // 2. Invalidate cache
    revalidateTag(`subscription-${customer}`);

    // 3. Send notification (chain another Inngest event if needed)
    await inngest.send({
      name: "notification.send",
      data: { userId: customer, template: "subscription-updated", status },
    });
  },
);
```

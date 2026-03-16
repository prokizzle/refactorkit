# Email Marketing and User Syncing Reference

Reference for integrating email marketing, drip campaigns, and user event syncing in a premium Next.js app.

---

## Email Stack Decision Matrix

| Approach | Drip Campaigns | Free Tier | Self-Hosted | Complexity |
|----------|---------------|-----------|-------------|------------|
| Loops (managed) | Built-in visual builder | 1K contacts, 4K emails/mo | No | Low |
| Inngest + Resend (code-owned) | Code as step functions | Inngest 50K exec + Resend 3K emails | Inngest: Yes | Medium |
| Customer.io | Best-in-class visual | Startup program only | No | Medium |

---

## Loops Setup (Recommended for Rapid Start)

- **Native Clerk integration**: auto-sync users on signup via webhook -- no custom code needed
- **21 pre-built drip templates**: onboarding, re-engagement, product updates
- **Transactional + marketing in one platform**: no separate providers for password resets vs campaigns
- **Setup**: Create Loops account -> Settings > Integrations > Clerk -> copy webhook URL -> in Clerk Dashboard add endpoint with `user.created`, `user.updated`, `user.deleted` events -> configure templates in Loops visual builder

---

## Inngest + Resend Code-Owned Path (Recommended for Ownership)

Full control over email logic using Inngest step functions and Resend for delivery.

### Inngest Email Sequence Function

```typescript
// apps/web/inngest/functions/welcome-sequence.ts
import { inngest } from "../client";
import { resend } from "@/lib/resend";
import { WelcomeEmail, OnboardingEmail, FeatureHighlightEmail } from "@repo/email";

export const welcomeSequence = inngest.createFunction(
  { id: "welcome-sequence" },
  { event: "user.created" },
  async ({ event, step }) => {
    await step.run("send-welcome", async () => {
      await resend.emails.send({
        from: "hello@yourdomain.com",
        to: event.data.email,
        subject: "Welcome!",
        react: WelcomeEmail({ name: event.data.firstName }),
      });
    });

    await step.sleep("wait-3-days", "3d");

    await step.run("send-onboarding", async () => {
      await resend.emails.send({
        from: "hello@yourdomain.com",
        to: event.data.email,
        subject: "Getting started",
        react: OnboardingEmail({ name: event.data.firstName }),
      });
    });

    await step.sleep("wait-7-days", "7d");

    await step.run("send-feature-highlight", async () => {
      await resend.emails.send({
        from: "hello@yourdomain.com",
        to: event.data.email,
        subject: "Features you might have missed",
        react: FeatureHighlightEmail({ name: event.data.firstName }),
      });
    });
  }
);
```

> Note: Inngest free tier has a 7-day max sleep limit between steps. Paid plans support longer durations.

### Resend Client Setup

```typescript
// apps/web/lib/resend.ts
import { Resend } from "resend";

export const resend = new Resend(process.env.RESEND_API_KEY);
```

---

## Clerk Webhook Setup for User Syncing

```typescript
// apps/web/app/api/webhooks/clerk/route.ts
import { headers } from "next/headers";
import { Webhook } from "svix";
import { WebhookEvent } from "@clerk/nextjs/server";
import { inngest } from "@/inngest/client";

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET;
  if (!WEBHOOK_SECRET) {
    throw new Error("Missing CLERK_WEBHOOK_SECRET env variable");
  }

  const headerPayload = await headers();
  const svixId = headerPayload.get("svix-id");
  const svixTimestamp = headerPayload.get("svix-timestamp");
  const svixSignature = headerPayload.get("svix-signature");

  if (!svixId || !svixTimestamp || !svixSignature) {
    return new Response("Missing svix headers", { status: 400 });
  }

  const payload = await req.json();
  const body = JSON.stringify(payload);

  const wh = new Webhook(WEBHOOK_SECRET);
  let evt: WebhookEvent;

  try {
    evt = wh.verify(body, {
      "svix-id": svixId,
      "svix-timestamp": svixTimestamp,
      "svix-signature": svixSignature,
    }) as WebhookEvent;
  } catch (err) {
    console.error("Webhook verification failed:", err);
    return new Response("Invalid signature", { status: 400 });
  }

  switch (evt.type) {
    case "user.created":
      await inngest.send({
        name: "user.created",
        data: {
          clerkId: evt.data.id,
          email: evt.data.email_addresses[0]?.email_address,
          firstName: evt.data.first_name,
          lastName: evt.data.last_name,
        },
      });
      break;

    case "user.updated":
      await inngest.send({
        name: "user.updated",
        data: { clerkId: evt.data.id, email: evt.data.email_addresses[0]?.email_address },
      });
      break;

    case "user.deleted":
      await inngest.send({
        name: "user.deleted",
        data: { clerkId: evt.data.id },
      });
      break;
  }

  return new Response("OK", { status: 200 });
}
```

---

## React Email Templates (packages/email/)

Shared email templates using `@react-email/components`. Place in the monorepo `packages/email/` directory.

### Example: Welcome Email

```typescript
// packages/email/emails/welcome.tsx
import { Body, Container, Head, Heading, Html, Link, Preview, Section, Text } from "@react-email/components";

interface WelcomeEmailProps {
  name?: string;
}

export const WelcomeEmail = ({ name = "there" }: WelcomeEmailProps) => (
  <Html>
    <Head />
    <Preview>Welcome to the platform</Preview>
    <Body style={{ backgroundColor: "#f6f9fc", fontFamily: "sans-serif" }}>
      <Container style={{ margin: "0 auto", padding: "40px 20px", maxWidth: "560px" }}>
        <Section style={{ backgroundColor: "#ffffff", borderRadius: "8px", padding: "32px" }}>
          <Heading style={{ fontSize: "24px", marginBottom: "16px" }}>Welcome, {name}</Heading>
          <Text style={{ fontSize: "16px", lineHeight: "24px", color: "#525f7f" }}>
            Thanks for signing up. Complete your profile, explore the dashboard, and connect your first integration.
          </Text>
          <Link href="https://yourdomain.com/dashboard" style={{ display: "inline-block", backgroundColor: "#000000", color: "#ffffff", padding: "12px 24px", borderRadius: "6px", textDecoration: "none", fontSize: "14px", marginTop: "16px" }}>
            Go to Dashboard
          </Link>
        </Section>
      </Container>
    </Body>
  </Html>
);

export default WelcomeEmail;
```

### Templates to Create

- `welcome.tsx` -- sent immediately on signup
- `onboarding-nudge.tsx` -- sent 3 days after signup if key actions not completed
- `invoice.tsx` -- sent on successful payment via Stripe webhook
- `password-reset.tsx` -- transactional, triggered by Clerk
- `feature-announcement.tsx` -- broadcast template for new feature launches

---

## Jitsu (Self-Hosted Segment Alternative)

Add Jitsu when you need event routing to a data warehouse (e.g., ClickHouse, BigQuery, Postgres).

- **License**: MIT, **Deploy**: Docker Compose or Kubernetes, **Cloud tier**: 200K events/mo free
- **Use case**: pipe user events to both your email provider and analytics warehouse
- **Setup**: add `@jitsu/js` client-side SDK, configure destinations in Jitsu dashboard
- Only introduce Jitsu when the app needs a central event bus beyond email triggers.

---

## PostHog Event Tracking

Track user actions for behavioral email triggers (e.g., "user has not completed onboarding after 3 days").

```typescript
// apps/web/lib/posthog.ts
import { PostHog } from "posthog-node";

export const posthog = new PostHog(process.env.NEXT_PUBLIC_POSTHOG_KEY!, {
  host: process.env.NEXT_PUBLIC_POSTHOG_HOST,
});

// Track an event
posthog.capture({
  distinctId: user.clerkId,
  event: "onboarding_completed",
  properties: { step: "profile_setup" },
});
```

Use PostHog cohorts or feature flags to segment users, then trigger Inngest events based on membership for targeted email campaigns.

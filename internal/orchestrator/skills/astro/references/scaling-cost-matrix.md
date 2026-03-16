# Scaling Cost Matrix and Debt Forecast

Reference for cost projections and lock-in analysis across the premium Next.js stack.
All costs are approximate as of March 2026 and should be verified with current pricing pages.

---

## Per-Tool Cost Matrix

| Tool | Free Tier Limit | ~1K users/mo | ~10K users/mo | ~100K users/mo | ~1M users/mo |
|------|----------------|-------------|---------------|---------------|-------------|
| Neon | 0.5GB, 190 CU-hr | $0 | $19 | ~$69 | ~$200+ |
| Supabase | 500MB, 50K MAU | $0 | $25 | $599 | Custom |
| Firebase Firestore | 1GB, 50K reads/day | $0 | ~$25 | ~$200 | ~$1000+ |
| Clerk | 10K MAU | $0 | $25+ | $100+ | $500+ |
| Stripe | No monthly fee | 2.9%+30c/txn | same | same | volume discount |
| PostHog | 1M events/mo | $0 | $0 | ~$450 | Custom |
| Sentry | 5K errors/mo | $0 | $26 | $80 | $320 |
| Inngest | 50K exec/mo | $0 | $0 | ~$50 | ~$200 |
| Resend | 3K emails/mo | $0 | $20 | $50 | Custom |
| Loops | 1K contacts | $0 | $49 | $159 | Custom |
| Upstash Redis | 10K cmd/day | $0 | $10 | $50 | $280 |
| Ably | Free tier | $0 | $29 | Custom | Custom |
| Vercel | Hobby | $0 | $20/user | $20/user | $20/user+usage |
| Payload CMS | Unlimited (self-host) | $0 | $0 | $0 | $0 |
| Sanity | 200K API/mo | $0 | $0 | $99 | $949 |

---

## Lock-In Risk Ratings

| Tool | Lock-in Risk | Refactor Cost | Data Portability | Exit Strategy |
|------|-------------|---------------|-----------------|---------------|
| Neon | Low | 1-2 weeks | Full (standard Postgres) | Dump + restore to any Postgres |
| Supabase | High | 4-8 weeks | Medium (proprietary auth, realtime) | Extract DB, rewrite auth + realtime |
| Firebase | Very High | 8-16 weeks | Low (proprietary APIs) | Full rewrite of data layer |
| Clerk | Medium | 2-4 weeks | Medium (user data exportable) | Swap to NextAuth/Auth.js |
| Stripe | Low | 1 week | Full (standard API) | Swap adapter in packages/billing |
| PostHog | Low | 1 week | Full (self-hostable) | Self-host or swap to Mixpanel |
| Inngest | Medium | 2-3 weeks | N/A (code-based) | Rewrite to BullMQ + Redis |

Key takeaway: the composable stack (Neon + Clerk + Stripe + PostHog) has the lowest
aggregate lock-in. Supabase and Firebase trade convenience for significantly higher
switching costs.

---

## Total Monthly Cost by Ecosystem Path

| Scale | Composable | Supabase | Firebase |
|-------|-----------|----------|----------|
| 1K users | $0 | $0 | $0 |
| 10K users | ~$50-100 | ~$75-150 | ~$100-200 |
| 100K users | ~$400-700 | ~$800-1200 | ~$1000-1500 |
| 1M users | ~$1500-3000 | ~$3000-5000 | ~$5000-10000 |

Composable = Neon + Clerk + Stripe + PostHog + Inngest + Resend + Vercel.
Supabase = Supabase (DB + auth + realtime) + Stripe + PostHog + Vercel.
Firebase = Firestore + Firebase Auth + Cloud Functions + Stripe + PostHog + Vercel.

---

## Debt Forecast Template

For each technology decision in the project, document the following:

```
### [Tool Name]
- Current cost: $X/mo
- Next milestone cost: $Y/mo at Z users
- Migration effort: N weeks if we need to switch
- What triggers the switch: [specific threshold or event]
- Self-hosted alternative: [tool name] + [complexity rating]
```

Example entry:

```
### PostHog (Analytics)
- Current cost: $0/mo (under 1M events)
- Next milestone cost: ~$450/mo at 100K users
- Migration effort: 1 week
- What triggers the switch: event volume exceeds 5M/mo or need data residency
- Self-hosted alternative: PostHog Docker + Medium complexity
```

Review this forecast quarterly. Update cost figures when crossing a pricing tier
or when a vendor changes their pricing model.

---

## Self-Hosted Alternatives

| Managed Tool | Self-Hosted Alternative | Complexity | Monthly Savings at 100K users |
|-------------|----------------------|------------|------------------------------|
| PostHog Cloud | PostHog Docker | Medium | ~$400 |
| Inngest Cloud | Inngest OSS | Medium | ~$50 |
| Novu Cloud | Novu Docker | Medium | ~$100 |
| Ably | Soketi | Low | ~$100-500 |
| Sanity | Payload CMS | Low | ~$99 |

Self-hosting adds operational overhead (monitoring, upgrades, backups). Factor in
~$50-200/mo for infrastructure and 2-4 hours/mo of maintenance per self-hosted tool
when calculating true savings.

---

## Decision Rules

1. Start on free tiers. Do not pre-optimize for scale you have not reached.
2. When a tool exceeds $100/mo, evaluate the self-hosted alternative.
3. When a tool exceeds $500/mo, actively plan migration or negotiate enterprise pricing.
4. Never adopt a "Very High" lock-in tool without documenting the exit strategy first.
5. Re-run this cost analysis at every 10x growth milestone.

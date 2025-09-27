---
sidebar_position: 1
title: Frequently Asked Questions
---

# Frequently Asked Questions

Answers to the questions creators, newsletter teams, and automation engineers ask most often.

## Getting started

**How do I access the API?**  
Request a free trial or paid plan at [sales@upworkjobsapi.com](mailto:sales@upworkjobsapi.com). Your API key is issued within minutes and works instantly with the [quick start guide](/docs/getting-started).

**Do I need a credit card for the trial?**  
No. The Starter tier includes 1,000 calls each month without a credit card. Upgrade when you are ready to scale.

**How fresh is the data?**  
Jobs flow into the API continuously. In most cases new briefs appear within 15-30 minutes of being posted on Upwork.

## Plans and usage

**What is included in each plan?**
- **Starter:** 1,000 calls/month, seven-day retention, email support.
- **Professional:** 25,000 calls/month, webhook delivery, 30-day retention, priority support.
- **Business:** 100,000 calls/month, historical lookups, Discord/Telegram delivery modules, account manager.
- **Enterprise:** Custom scale, dedicated infrastructure, full archive access, bespoke integrations.

**Can I change plans later?**  
Yes. Upgrades take effect immediately. Downgrades apply at the next billing cycle. All adjustments are prorated.

**What happens if I exceed my quota?**  
Professional and Business plans charge $0.01 for each additional call. Starter requests pause until the next cycle or an upgrade. Enterprise overages follow your contract.

## Technical details

**What format do responses use?**  
All endpoints return JSON with a consistent envelope (`success`, `data`, `count`, `last_updated`). See the [endpoint reference](/docs/api/endpoints) for full schema details.

**Which filters are available?**  
You can filter by payment verification, spend, skills/tags, category, posting time, and more. Review the [filtering guide](/docs/api/filtering) for the full matrix.

**Do you support webhooks?**  
Webhook delivery is included in Professional plans and above. You can push events directly to n8n, Zapier, Make, or any HTTPS endpoint.

**Can I integrate with Discord or Telegram?**  
Yes. Use Zapier, Make, or our direct webhook payloads to post formatted digests into channels. Business plans include dedicated modules and onboarding templates.

## Data and compliance

**How accurate is the dataset?**  
We filter spam and duplicates automatically and track buyer metadata to maintain a 99.5% accuracy rate.

**Is historical data available?**  
Starter retains seven days, Professional thirty, Business six months, and Enterprise unlocks the full archive with replay tooling.

**Is the API compliant with GDPR/SOC 2?**  
Yes. Data is encrypted in transit and at rest. Enterprise plans include data processing agreements and customised compliance support.

## Troubleshooting

**I received a 401 error. What now?**  
Confirm the header is `X-API-KEY`, ensure the value is correct, and check that the key has not been rotated or revoked.

**My query returns zero results.**  
Start broader: remove one filter at a time, verify parameter names, and confirm you are using ISO 8601 timestamps where required.

**Requests seem slow.**  
Limit the number of records per call, narrow filters, and cache recent responses. Network latency in automation platforms can also contribute to delays.

## Support

- Technical support: [support@upworkjobsapi.com](mailto:support@upworkjobsapi.com) (response within plan-specific SLA)
- Billing: [accounts@upworkjobsapi.com](mailto:accounts@upworkjobsapi.com)
- Sales and partnerships: [sales@upworkjobsapi.com](mailto:sales@upworkjobsapi.com)

Still need help? [Book a consultation](mailto:sales@upworkjobsapi.com?subject=Consultation%20Request) and we will walk through your use case, filters, and automation plan.

---
sidebar_position: 4
title: Getting Started
---

# Getting Started

Spin up Upwork Jobs API, capture your first briefs, and connect them to the channels where you plan, automate, and publish. The workflow below takes most teams less than ten minutes.

## Quick start in three steps

### 1. Request access
- Start with the free tier for 1,000 monthly calls: [Create free account](mailto:sales@upworkjobsapi.com?subject=Free%20Trial%20Signup).
- Paid plans unlock higher limits, historical lookups, and dedicated support: [Compare plans](/docs/pricing).
- Your API key arrives by email; keep it available for the next steps.

### 2. Verify your credentials

```bash
curl -H "X-API-KEY: your-api-key" \
     "https://api.upworkjobsapi.com/health"
```

Expected response:

```json
{"success": true, "message": "API is healthy", "count": 0}
```

### 3. Fetch your first batch

```bash
curl -H "X-API-KEY: your-api-key" \
     "https://api.upworkjobsapi.com/jobs?limit=5&payment_verified=true&category=writing-translation"
```

Each job includes stable IDs, buyer spend, budgets, tags, and timestamps so you can immediately rank which briefs fit your next issue or segment.

## Authenticate securely

Send your key in the `X-API-KEY` header on every request.

```http
GET /jobs?limit=20 HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key
Content-Type: application/json
```

Best practices:
- Store keys in environment variables or your team secret manager.
- Use separate keys for development, staging, and production automations.
- Rotate keys quarterly and remove unused credentials promptly.

## Map data to your stack

| Target | How to connect | Recommended fields |
|--------|----------------|--------------------|
| Airtable / Notion | Use Zapier, Make, or n8n to push API responses into tables each morning. | `title`, `budget.fixed_amount`, `buyer.payment_verified`, `skills`, `url`, `posted_on` |
| Idea backlogs | Parse the JSON into your prompt templates to surface pain points, solution framing, and positioning cues. | `description`, `tags`, `client_activity.total_applicants`, `buyer.total_spent` |
| Newsletters & digests | Schedule a daily fetch, score briefs, and send highlights to ESPs or chat communities. | `title`, `budget`, `posted_on`, curated metadata |

## Automation recipes

### n8n
1. Add an **HTTP Request** node with `GET https://api.upworkjobsapi.com/jobs`.
2. Configure headers with `X-API-KEY` and any query parameters (for example `tags=ai,marketing`).
3. Use a **Set** node to reshape fields into your newsletter template.
4. Deliver to Notion, Google Sheets, or MailerLite nodes downstream.

### Zapier
1. Create a **Schedule** trigger (daily or hourly).
2. Add a **Webhooks by Zapier** step (GET) targeting the `/jobs` endpoint.
3. Filter or format records with **Formatter** utilities.
4. Send the curated list to Airtable, Slack, Discord, or your ESP.

### Make (Integromat)
1. Start with a **HTTP** module and set authentication headers.
2. Insert a **Array aggregator** to batch jobs into a single payload.
3. Push results to Google Docs for script outlines or to Telegram for curated drops.

## Creator-focused examples

### Python: score briefs for newsletter fit
```python
import requests

API_URL = "https://api.upworkjobsapi.com/jobs"
HEADERS = {"X-API-KEY": "your-api-key"}
PARAMS = {
    "payment_verified": True,
    "tags": "newsletter,content marketing",
    "limit": 15,
    "sort": "posted_on_desc"
}

response = requests.get(API_URL, headers=HEADERS, params=PARAMS, timeout=10)
response.raise_for_status()
jobs = response.json()["data"]

newsletter_ready = [job for job in jobs if job["buyer"]["total_spent"] >= 10000]
```

### Node.js: push high-signal jobs to Discord
```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'https://api.upworkjobsapi.com',
  headers: { 'X-API-KEY': process.env.UPWORK_API_KEY }
});

const { data } = await api.get('/jobs', {
  params: { payment_verified: true, limit: 5, tags: 'ai,research' }
});

const topBriefs = data.data
  .map((job) => `- **${job.title}** - $${job.budget?.fixed_amount ?? 'N/A'} | ${job.buyer.total_spent} lifetime spend`)
  .join('\n');

await axios.post(process.env.DISCORD_WEBHOOK_URL, {
  content: `Fresh AI briefs from Upwork:\n${topBriefs}`
});
```

### cURL: Telegram digest via Make webhook
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"limit": 10, "tags": "copywriting,marketing"}' \
  "https://hook.make.com/your-make-webhook-id"
```
Connect the webhook to a Make scenario that calls the Upwork Jobs API, formats the highlights, and posts them into your Telegram channel.

## Next milestones

1. [Dive into endpoint specifics](/docs/api/endpoints) for full schema details.
2. [Fine-tune filters](/docs/api/filtering) to align with your beat or niche.
3. [Confirm rate limits](/docs/support/rate-limits) before scheduling automations.
4. [Reach out to support](/docs/support/contact) for custom integrations or onboarding workshops.

Stay close to buyer demand, build better ideas, and keep your audience subscribed with fresher insights.

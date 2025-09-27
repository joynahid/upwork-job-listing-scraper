---
sidebar_position: 2
title: API Endpoints
---

# API Endpoints

Upwork Jobs API exposes concise REST endpoints with predictable JSON payloads. All responses share the same envelope so you can map fields into automations without custom parsing.

## Health check

### `GET /health`

Verify connectivity, authentication, and uptime. Ideal for scheduled monitors or preflight checks in automation platforms.

**Request**
```http
GET /health HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key
```

**Response**
```json
{
  "success": true,
  "message": "API is healthy",
  "data": [],
  "count": 0,
  "last_updated": "2024-10-24T12:00:32Z"
}
```

## Jobs endpoint

### `GET /jobs`

Pull curated Upwork job postings enriched with buyer intelligence. Combine query parameters to match your beat, niche, or audience needs.

**Base URL**: `https://api.upworkjobsapi.com/jobs`

**Request example**
```http
GET /jobs?limit=10&payment_verified=true&tags=newsletter,saas HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key
```

**Response example**
```json
{
  "success": true,
  "data": [
    {
      "id": "upwork-872341",
      "title": "Launch a weekly AI founder newsletter",
      "description": "We need a researcher-writer to source stories and trends...",
      "job_type": 2,
      "status": 1,
      "contractor_tier": 2,
      "posted_on": "2024-10-24T08:12:43Z",
      "category": {
        "name": "Writing & Translation",
        "slug": "writing-translation",
        "group": "Sales & Marketing",
        "group_slug": "sales-marketing"
      },
      "budget": {
        "fixed_amount": 2500,
        "currency": "USD"
      },
      "hourly_budget": null,
      "buyer": {
        "payment_verified": true,
        "country": "US",
        "city": "Austin",
        "timezone": "America/Chicago",
        "total_spent": 84500,
        "total_assignments": 28,
        "total_jobs_with_hires": 21
      },
      "skills": ["newsletter", "ai research", "marketing"],
      "tags": ["founder stories", "growth marketing"],
      "client_activity": {
        "total_applicants": 12,
        "total_hired": 4,
        "total_invited_to_interview": 7,
        "unanswered_invites": 1,
        "invitations_sent": 9,
        "last_buyer_activity": "2024-10-24T07:55:12Z"
      },
      "url": "https://www.upwork.com/jobs/~01abc123",
      "last_visited_at": "2024-10-24T09:05:03Z",
      "duration_label": "3 to 6 months",
      "engagement": "part-time",
      "is_private": false,
      "privacy_reason": ""
    }
  ],
  "count": 1,
  "last_updated": "2024-10-24T08:14:03Z"
}
```

## Response schema

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Indicates whether the request succeeded. |
| `data` | array | Collection of job objects. Empty array if no matches. |
| `count` | integer | Number of items in `data`. |
| `last_updated` | string (ISO 8601) | Timestamp when the response was generated. |
| `message` | string | Present on error responses to explain the failure. |

### Job object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Stable identifier for deduplication and caching. |
| `title` | string | Project headline provided by the client. |
| `description` | string | Full brief containing requirements and scope. |
| `url` | string | Direct Upwork link for manual review. |
| `posted_on` | string | Publish timestamp (ISO 8601). |
| `last_visited_at` | string | Most recent client activity timestamp. |
| `job_type` | integer | `1` for hourly, `2` for fixed price contracts. |
| `status` | integer | `1` open, `2` closed. |
| `contractor_tier` | integer | Experience level: `1` entry, `2` intermediate, `3` expert. |
| `category` | object | Includes `name`, `slug`, `group`, and `group_slug`. |
| `budget` | object | `fixed_amount` (number) and `currency` (string) when available. |
| `hourly_budget` | object/null | Contains `min` and `max` hourly rates when hourly. |
| `buyer` | object | Client profile with verification status and spend history. |
| `skills` | array | Primary skills required. |
| `tags` | array | Additional keywords and themes. |
| `client_activity` | object | Application, invitation, and hire counts. |
| `duration_label` | string | Human-readable duration estimate. |
| `engagement` | string | Typical schedule (full-time, part-time, etc.). |
| `is_private` | boolean | Indicates if listing details are restricted. |
| `privacy_reason` | string | Reason when `is_private` is true. |

## Error handling

Error responses retain the same envelope so your automations can branch on `success` or HTTP status.

```json
{
  "success": false,
  "message": "Invalid budget_min parameter",
  "data": [],
  "count": 0,
  "last_updated": "2024-10-24T12:00:00Z"
}
```

Common status codes:

| Status | Meaning | Suggested action |
|--------|---------|------------------|
| 400 | Validation error | Review query parameters and value formats. |
| 401 | Authentication failure | Confirm the `X-API-KEY` header and rotate if compromised. |
| 404 | Endpoint not found | Verify the path. |
| 429 | Rate limit exceeded | Implement retries with exponential backoff or upgrade your plan. |
| 500 | Internal server error | Retry after a short delay or contact support. |

## Integration patterns

| Scenario | Recommended flow |
|----------|------------------|
| n8n | HTTP node -> Set/Function node for scoring -> Notion/Airtable/Slack modules. |
| Zapier | Schedule trigger -> Webhooks (GET) -> Formatter -> ESP/Discord modules. |
| Make | HTTP module -> Array aggregator -> Telegram/Google Sheets/Discord modules. |
| Direct webhooks | Use `/jobs` with `webhook=true` (Professional+) to receive pushed updates without polling. |
| Discord/Telegram | Connect via Zapier, Make, or direct webhooks and format summaries into bullet lists or embeds. |

## Testing checklist

1. Call `/health` to confirm connectivity and authentication.
2. Request `/jobs?limit=1` to verify schema mapping.
3. Add filters incrementally and log both HTTP status and payload size.
4. Store `last_updated` to reconcile batching jobs in downstream systems.
5. Monitor rate limit headers and configure retries at the platform level.

Need something custom? Email [support@upworkjobsapi.com](mailto:support@upworkjobsapi.com) and share your workflow diagram so we can help validate schema mapping or provide tailored snippets.

---
sidebar_position: 3
title: Filtering & Search
---

# Filtering & Search

Combine filters to isolate the briefs that matter for your creative pipeline. All filters are additive: start broad, then narrow by buyer quality, budget, or niche.

## Quick recipes

### High-intent newsletter briefs
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?tags=newsletter,ghostwriting&payment_verified=true&budget_min=1500"
```

### Recurring content retainers
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?contractor_tier=2&duration_label=ongoing&buyer.total_jobs_with_hires_min=5"
```

### Fresh posts in your niche
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?category=marketing-branding-sales&posted_after=2024-10-20T00:00:00Z&sort=posted_on_desc"
```

## Filter reference

### Core parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `limit` | integer | Number of results (1-50, default 20). | `limit=15` |
| `offset` | integer | Pagination offset, multiples of `limit`. | `offset=20` |
| `payment_verified` | boolean | Require verified payment method. | `payment_verified=true` |
| `category` | string | Category slug. | `category=writing-translation` |
| `category_group` | string | Broader category grouping. | `category_group=marketing` |
| `country` | string | Two-letter country code. | `country=US` |

### Budget controls

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `budget_min` | number | Minimum fixed budget in currency units. | `budget_min=2000` |
| `budget_max` | number | Maximum fixed budget. | `budget_max=10000` |
| `hourly_min` | number | Minimum hourly rate. | `hourly_min=75` |
| `hourly_max` | number | Maximum hourly rate. | `hourly_max=150` |

### Job structure

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `job_type` | integer | `1` hourly, `2` fixed price. | `job_type=2` |
| `status` | integer | `1` open, `2` closed. | `status=1` |
| `contractor_tier` | integer | `1` entry, `2` intermediate, `3` expert. | `contractor_tier=3` |
| `duration_label` | string | Free-text duration label. | `duration_label=ongoing` |
| `engagement` | string | Engagement type such as `part-time`. | `engagement=part-time` |

### Time windows

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `posted_after` | ISO 8601 | Jobs posted after timestamp. | `posted_after=2024-10-22T00:00:00Z` |
| `posted_before` | ISO 8601 | Jobs posted before timestamp. | `posted_before=2024-10-24T23:59:59Z` |
| `last_visited_after` | ISO 8601 | Client activity after timestamp. | `last_visited_after=2024-10-23T00:00:00Z` |

### Skills and tags

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `tags` | comma-separated string | Match any of the provided tags. | `tags=podcast,newsletter` |
| `skills` | comma-separated string | Alias for `tags`. | `skills=seo,copywriting` |

### Buyer intelligence

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `buyer.total_spent_min` | number | Minimum historical spend (USD). | `buyer.total_spent_min=20000` |
| `buyer.total_assignments_min` | number | Minimum completed jobs. | `buyer.total_assignments_min=10` |
| `buyer.total_jobs_with_hires_min` | number | Minimum hires recorded. | `buyer.total_jobs_with_hires_min=5` |

### Sorting

| Value | Description |
|-------|-------------|
| `posted_on_asc` | Oldest jobs first. |
| `posted_on_desc` | Newest jobs first (default). |
| `last_visited_asc` | Least recently active clients first. |
| `last_visited_desc` | Most recently active clients first. |
| `budget_asc` | Lower budgets first. |
| `budget_desc` | Higher budgets first. |

## Advanced strategies

### Prioritise high-spend buyers
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?buyer.total_spent_min=50000&buyer.total_jobs_with_hires_min=8&payment_verified=true"
```

### Surface emerging topics
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?tags=ai,founder stories&posted_after=2024-10-21T00:00:00Z"
```

### Build community digests
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?limit=25&sort=posted_on_desc&category=marketing-branding-sales"
```

## Optimisation tips

- Use `limit` judiciously; multiple smaller calls are easier to buffer in automation tools.
- Cache results between runs to stay well within rate limits.
- Log both the request URL and `last_updated` timestamp to trace batches.
- Combine filters gradually; test each new parameter before adding another.

## Troubleshooting

| Issue | Likely cause | Fix |
|-------|--------------|-----|
| Empty results | Filters too specific | Remove filters one at a time to identify the blocker. |
| 400 error | Parameter typo or invalid value | Confirm parameter names and data types. |
| Slow response | Large limits with broad filters | Narrow your query or page through results. |

Need a filter that is not listed? Contact [support@upworkjobsapi.com](mailto:support@upworkjobsapi.com) and share an example brief; we can suggest field combinations or roadmap the enhancement.

---
sidebar_position: 2
title: API Endpoints
---

# API Endpoints

Our API provides clean, structured access to Upwork job data through simple REST endpoints. All responses are in JSON format with consistent structure.

## 🏥 Health Check

### `GET /health`

Check if the API service is running and responsive.

**Request:**
```http
GET /health HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key
```

**Response:**
```json
{
  "success": true,
  "message": "API is healthy",
  "last_updated": "2024-09-27T12:34:56Z",
  "data": [],
  "count": 0
}
```

**Use Cases:**
- Monitor API availability
- Health checks in your applications
- Validate API key functionality

---

## 💼 Jobs Endpoint

### `GET /jobs`

Retrieve Upwork job postings with advanced filtering and search capabilities.

**Base URL:** `https://api.upworkjobsapi.com/jobs`

**Request:**
```http
GET /jobs?limit=10&payment_verified=true&budget_min=1000 HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "upwork-123456",
      "title": "Senior React Developer for SaaS Platform",
      "description": "We're looking for an experienced React developer...",
      "job_type": 2,
      "status": 1,
      "contractor_tier": 2,
      "posted_on": "2024-09-27T10:30:00Z",
      "category": {
        "name": "Web Development",
        "slug": "web-development",
        "group": "Development & IT",
        "group_slug": "development-it"
      },
      "budget": {
        "fixed_amount": 5000,
        "currency": "USD"
      },
      "buyer": {
        "payment_verified": true,
        "country": "US",
        "city": "San Francisco",
        "timezone": "America/Los_Angeles",
        "total_spent": 125000,
        "total_assignments": 25,
        "total_jobs_with_hires": 18
      },
      "tags": ["react", "typescript", "saas"],
      "url": "https://www.upwork.com/jobs/~01abc123",
      "last_visited_at": "2024-09-27T11:15:00Z",
      "duration_label": "3 to 6 months",
      "engagement": "part-time",
      "skills": ["React", "TypeScript", "Node.js", "AWS"],
      "hourly_budget": null,
      "client_activity": {
        "total_applicants": 15,
        "total_hired": 2,
        "total_invited_to_interview": 5,
        "unanswered_invites": 1,
        "invitations_sent": 8,
        "last_buyer_activity": "2024-09-27T09:45:00Z"
      },
      "location": {
        "country": "US",
        "city": "San Francisco",
        "timezone": "America/Los_Angeles"
      },
      "is_private": false,
      "privacy_reason": ""
    }
  ],
  "count": 1,
  "last_updated": "2024-09-27T12:35:42Z"
}
```

## 📊 Response Fields Explained

### Job Information
| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique job identifier |
| `title` | string | Job posting title |
| `description` | string | Full job description |
| `url` | string | Direct link to Upwork job posting |
| `posted_on` | string | When the job was posted (ISO 8601) |
| `last_visited_at` | string | Last client activity timestamp |

### Budget & Pricing
| Field | Type | Description |
|-------|------|-------------|
| `budget.fixed_amount` | number | Fixed project budget (if applicable) |
| `budget.currency` | string | Currency code (USD, EUR, etc.) |
| `hourly_budget.min` | number | Minimum hourly rate |
| `hourly_budget.max` | number | Maximum hourly rate |

### Client Intelligence
| Field | Type | Description |
|-------|------|-------------|
| `buyer.payment_verified` | boolean | Client has verified payment method |
| `buyer.total_spent` | number | Total amount spent on Upwork |
| `buyer.total_assignments` | number | Number of completed projects |
| `buyer.total_jobs_with_hires` | number | Jobs where client hired someone |
| `buyer.country` | string | Client's country |

### Project Details
| Field | Type | Description |
|-------|------|-------------|
| `job_type` | number | 1=Hourly, 2=Fixed Price |
| `contractor_tier` | number | 1=Entry, 2=Intermediate, 3=Expert |
| `duration_label` | string | Expected project duration |
| `engagement` | string | full-time, part-time, etc. |
| `skills` | array | Required skills and technologies |

### Market Activity
| Field | Type | Description |
|-------|------|-------------|
| `client_activity.total_applicants` | number | Number of freelancers who applied |
| `client_activity.total_hired` | number | Number of freelancers hired |
| `client_activity.invitations_sent` | number | Invitations sent by client |
| `client_activity.last_buyer_activity` | string | Last time client was active |

## 🚨 Error Responses

### Bad Request (400)
```json
{
  "success": false,
  "message": "Invalid budget_min parameter",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```

### Unauthorized (401)
```json
{
  "success": false,
  "message": "Invalid or missing X-API-KEY header",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```

### Rate Limited (429)
```json
{
  "success": false,
  "message": "Rate limit exceeded. Upgrade your plan for higher limits.",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```

### Server Error (500)
```json
{
  "success": false,
  "message": "Internal server error. Please try again later.",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```

## 🔄 Response Consistency

All API responses follow the same structure:

- `success`: Boolean indicating if the request was successful
- `data`: Array of results (empty array if no results)
- `count`: Number of items in the data array
- `last_updated`: Timestamp when the response was generated
- `message`: Human-readable status message (present on errors)

## 📈 Performance Tips

### Optimize Your Requests
- **Use appropriate limits**: Don't request more data than you need
- **Cache responses**: Store results to reduce API calls
- **Filter effectively**: Use specific filters to get targeted results
- **Batch processing**: Process multiple jobs in single requests

### Rate Limit Management
- **Monitor usage**: Track your API call consumption
- **Implement backoff**: Handle rate limits gracefully
- **Upgrade when needed**: Higher plans have better rate limits

## 🛠️ Testing Your Integration

### Using cURL
```bash
# Test health endpoint
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/health"

# Test jobs endpoint
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?limit=1"
```

### Postman Collection
We provide a complete Postman collection with example requests:
[Download Postman Collection](mailto:support@upworkjobsapi.com?subject=Postman%20Collection%20Request)

---

**Need help with integration?**

[Contact Technical Support](mailto:support@upworkjobsapi.com)
[View Filtering Guide](/docs/api/filtering)

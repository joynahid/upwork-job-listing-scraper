---
sidebar_position: 1
title: Authentication
---

# API Authentication

Authenticate every request with your unique API key. The header is required for all endpoints, including the health check.

## API key header

Send the header exactly as shown below:

```http
GET /jobs HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key
Content-Type: application/json
```

- Use HTTPS for all requests.
- Keys are case sensitive; copy them without extra spaces.
- Authentication failures return `401 Unauthorized`.

## Getting a key

1. Request trial access or select a paid plan by emailing [sales@upworkjobsapi.com](mailto:sales@upworkjobsapi.com).
2. Receive your key via secure email within minutes.
3. Store the key in your secrets manager or environment variables.

## Security best practices

- **Server-side use only**: Never expose keys in client-side scripts or public repositories.
- **Environment isolation**: Maintain separate keys for development, staging, and production automations.
- **Rotation**: Rotate keys Quarterly or when teammates roll off a project.
- **Monitoring**: Track usage and revoke keys immediately if you notice unexpected traffic.

Example `.env` usage:

```bash
# .env
UPWORK_API_KEY=your-api-key

# local usage
curl -H "X-API-KEY: $UPWORK_API_KEY" \
     "https://api.upworkjobsapi.com/jobs?limit=10"
```

## Error responses

| Scenario | Status | Example message |
|----------|--------|-----------------|
| Missing header | 401 | `{"success": false, "message": "X-API-KEY header is required"}` |
| Invalid key | 401 | `{"success": false, "message": "Invalid or missing X-API-KEY header"}` |
| Rate limited | 429 | `{"success": false, "message": "Rate limit exceeded. Upgrade your plan for higher limits."}` |

## Rotating keys safely

1. Generate a new key in your console or by contacting support.
2. Update downstream services in order: staging -> production.
3. Confirm successful requests with the new key.
4. Deactivate the old key once traffic is stable.

## Example implementations

### Python
```python
import os
import requests

api_key = os.environ["UPWORK_API_KEY"]
headers = {"X-API-KEY": api_key, "Content-Type": "application/json"}
response = requests.get("https://api.upworkjobsapi.com/jobs", headers=headers, timeout=10)
response.raise_for_status()
```

### Node.js (Axios)
```javascript
import axios from 'axios';

const client = axios.create({
  baseURL: 'https://api.upworkjobsapi.com',
  headers: { 'X-API-KEY': process.env.UPWORK_API_KEY }
});

const { data } = await client.get('/jobs', { params: { limit: 5 } });
```

### PHP
```php
$headers = [
    'X-API-KEY: ' . getenv('UPWORK_API_KEY'),
    'Content-Type: application/json'
];

$context = stream_context_create([
    'http' => [
        'method' => 'GET',
        'header' => implode("\r\n", $headers)
    ]
]);

$response = file_get_contents('https://api.upworkjobsapi.com/jobs?limit=10', false, $context);
```

## Need help?

- Technical support: [support@upworkjobsapi.com](mailto:support@upworkjobsapi.com)
- Key rotations and plan upgrades: [sales@upworkjobsapi.com](mailto:sales@upworkjobsapi.com)
- Quick answers: [FAQ](/docs/support/faq)

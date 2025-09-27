---
sidebar_position: 1
title: Authentication
---

# API Authentication

Secure access to our API using your unique API key. All requests must include proper authentication to access premium job data.

## 🔑 API Key Authentication

Every API request requires your API key in the `X-API-KEY` header:

```http
GET /jobs HTTP/1.1
Host: api.upworkjobsapi.com
X-API-KEY: your-api-key-here
Content-Type: application/json
```

## 🚀 Getting Your API Key

### Free Trial
1. [Sign up for free trial](mailto:sales@upworkjobsapi.com?subject=Free%20Trial%20Signup)
2. Receive your API key via email within 5 minutes
3. Start making API calls immediately

### Paid Plans
1. [Choose your plan](/docs/pricing)
2. Complete payment setup
3. Receive production API key with higher limits

## 🔒 Security Best Practices

### Keep Your Key Secure
- **Never expose in client-side code** - API keys should only be used server-side
- **Store as environment variables** - Don't hardcode keys in your source code
- **Use HTTPS only** - All API calls must use secure connections
- **Rotate regularly** - Update your keys periodically for security

### Environment Variables
```bash
# .env file
UPWORK_API_KEY=your-api-key-here

# Usage in code
curl -H "X-API-KEY: $UPWORK_API_KEY" \
  "https://api.upworkjobsapi.com/jobs"
```

### Multiple Environments
Use different API keys for different environments:

- **Development**: Limited rate limits, test data
- **Staging**: Production-like environment for testing
- **Production**: Full access with your plan's limits

## ❌ Authentication Errors

### Invalid API Key
```json
{
  "success": false,
  "message": "Invalid or missing X-API-KEY header",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```
**Status Code**: `401 Unauthorized`

### Missing API Key
```json
{
  "success": false,
  "message": "X-API-KEY header is required",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```
**Status Code**: `401 Unauthorized`

### Rate Limit Exceeded
```json
{
  "success": false,
  "message": "Rate limit exceeded. Upgrade your plan for higher limits.",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z"
}
```
**Status Code**: `429 Too Many Requests`

## 🔄 API Key Management

### Rotating Keys
1. Generate new API key in your account dashboard
2. Update your applications with the new key
3. Test thoroughly before deactivating the old key
4. Deactivate old key once migration is complete

### Monitoring Usage
- Track API calls in your account dashboard
- Set up alerts for approaching rate limits
- Monitor for unusual usage patterns

### Multiple Keys
Enterprise customers can have multiple API keys for:
- Different applications or services
- Team member access control
- Environment separation

## 🛠️ Implementation Examples

### cURL
```bash
curl -H "X-API-KEY: your-api-key" \
     -H "Content-Type: application/json" \
     "https://api.upworkjobsapi.com/jobs?limit=10"
```

### Python (requests)
```python
import requests
import os

headers = {
    'X-API-KEY': os.getenv('UPWORK_API_KEY'),
    'Content-Type': 'application/json'
}

response = requests.get(
    'https://api.upworkjobsapi.com/jobs',
    headers=headers,
    params={'limit': 10}
)
```

### JavaScript (axios)
```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'https://api.upworkjobsapi.com',
  headers: {
    'X-API-KEY': process.env.UPWORK_API_KEY,
    'Content-Type': 'application/json'
  }
});

const response = await api.get('/jobs', {
  params: { limit: 10 }
});
```

### PHP
```php
$apiKey = $_ENV['UPWORK_API_KEY'];
$headers = [
    'X-API-KEY: ' . $apiKey,
    'Content-Type: application/json'
];

$context = stream_context_create([
    'http' => [
        'method' => 'GET',
        'header' => implode("\r\n", $headers)
    ]
]);

$response = file_get_contents(
    'https://api.upworkjobsapi.com/jobs?limit=10',
    false,
    $context
);
```

## 🆘 Need Help?

### Common Issues
- **Key not working?** Check for extra spaces or characters
- **Getting 401 errors?** Verify the header name is exactly `X-API-KEY`
- **Rate limits?** [Upgrade your plan](/docs/pricing) for higher limits

### Support
- [Technical Support](mailto:support@upworkjobsapi.com) - For authentication issues
- [Sales Team](mailto:sales@upworkjobsapi.com) - For plan upgrades
- [FAQ](/docs/support/faq) - Common questions and solutions

---

**Ready to start making authenticated API calls?**

[Get Your Free API Key](mailto:sales@upworkjobsapi.com?subject=Free%20Trial%20Signup)

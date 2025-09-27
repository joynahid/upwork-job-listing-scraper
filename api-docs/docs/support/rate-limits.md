---
sidebar_position: 2
title: Rate Limits & Usage
---

# Rate Limits & Usage

Understand how rate limits work and optimize your API usage for the best performance and cost efficiency.

## Plan limits by tier

| Plan | Requests/Minute | Monthly Calls | Burst Limit |
|------|----------------|---------------|-------------|
| **Starter** | 10 | 1,000 | 20 |
| **Professional** | 100 | 25,000 | 200 |
| **Business** | 500 | 100,000 | 1,000 |
| **Enterprise** | Custom | Unlimited | Custom |

## How the limiter works

### Request Counting
- Each API call counts as one request
- Failed requests (4xx/5xx errors) still count toward limits
- Cached responses don't count against your limit

### Time Windows
- Rate limits are calculated per minute
- Limits reset at the start of each minute
- Burst limits allow temporary spikes above the per-minute rate

### Response Headers
Every API response includes rate limit information:

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1695825600
X-RateLimit-Burst: 200
```

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Requests per minute for your plan |
| `X-RateLimit-Remaining` | Requests left in current window |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |
| `X-RateLimit-Burst` | Maximum burst requests allowed |

## Example rate limit response

When you exceed your rate limit, you'll receive:

```json
{
  "success": false,
  "message": "Rate limit exceeded. Try again in 30 seconds.",
  "data": [],
  "count": 0,
  "last_updated": "2024-09-27T12:00:00Z",
  "retry_after": 30
}
```

**Status Code**: `429 Too Many Requests`

## Recommended safeguards

### 1. Implement Exponential Backoff

```python
import time
import requests
from random import uniform

def api_request_with_backoff(url, headers, params, max_retries=3):
    for attempt in range(max_retries):
        response = requests.get(url, headers=headers, params=params)
        
        if response.status_code == 200:
            return response.json()
        elif response.status_code == 429:
            # Rate limited - wait and retry
            retry_after = int(response.headers.get('Retry-After', 60))
            wait_time = retry_after + uniform(0, 10)  # Add jitter
            time.sleep(wait_time)
        else:
            # Other error - don't retry
            response.raise_for_status()
    
    raise Exception("Max retries exceeded")
```

### 2. Monitor Rate Limit Headers

```javascript
const axios = require('axios');

async function makeApiRequest(url, headers, params) {
    try {
        const response = await axios.get(url, { headers, params });
        
        // Check rate limit status
        const remaining = parseInt(response.headers['x-ratelimit-remaining']);
        const limit = parseInt(response.headers['x-ratelimit-limit']);
        
        if (remaining < limit * 0.1) {  // Less than 10% remaining
            console.warn(`Rate limit warning: ${remaining}/${limit} requests remaining`);
        }
        
        return response.data;
    } catch (error) {
        if (error.response?.status === 429) {
            const retryAfter = parseInt(error.response.headers['retry-after']);
            console.log(`Rate limited. Retry after ${retryAfter} seconds`);
        }
        throw error;
    }
}
```

### 3. Implement Request Queuing

```python
import asyncio
import aiohttp
from asyncio import Semaphore

class RateLimitedClient:
    def __init__(self, requests_per_minute=100):
        self.semaphore = Semaphore(requests_per_minute)
        self.last_request_time = 0
        self.min_interval = 60 / requests_per_minute
    
    async def make_request(self, url, headers, params):
        async with self.semaphore:
            # Ensure minimum interval between requests
            now = asyncio.get_event_loop().time()
            time_since_last = now - self.last_request_time
            if time_since_last < self.min_interval:
                await asyncio.sleep(self.min_interval - time_since_last)
            
            async with aiohttp.ClientSession() as session:
                async with session.get(url, headers=headers, params=params) as response:
                    self.last_request_time = asyncio.get_event_loop().time()
                    return await response.json()
```

## Optimising usage

### 1. Smart Caching
Cache responses to reduce API calls:

```python
import redis
import json
import hashlib

class APICache:
    def __init__(self, redis_client, ttl=300):  # 5 minute TTL
        self.redis = redis_client
        self.ttl = ttl
    
    def cache_key(self, params):
        # Create unique key from parameters
        param_str = json.dumps(params, sort_keys=True)
        return f"upwork_jobs:{hashlib.md5(param_str.encode()).hexdigest()}"
    
    def get(self, params):
        key = self.cache_key(params)
        cached = self.redis.get(key)
        return json.loads(cached) if cached else None
    
    def set(self, params, data):
        key = self.cache_key(params)
        self.redis.setex(key, self.ttl, json.dumps(data))

# Usage
cache = APICache(redis.Redis())

def get_jobs_cached(params):
    # Check cache first
    cached_result = cache.get(params)
    if cached_result:
        return cached_result
    
    # Make API call if not cached
    result = make_api_request(params)
    cache.set(params, result)
    return result
```

### 2. Batch Processing
Process multiple jobs efficiently:

```python
def process_jobs_in_batches(filters, batch_size=50):
    all_jobs = []
    offset = 0
    
    while True:
        # Request batch
        params = {**filters, 'limit': batch_size, 'offset': offset}
        response = make_api_request(params)
        
        jobs = response['data']
        if not jobs:
            break
            
        all_jobs.extend(jobs)
        offset += batch_size
        
        # Rate limit friendly delay
        time.sleep(1)  # 1 second between batches
    
    return all_jobs
```

### 3. Efficient Filtering
Use specific filters to reduce data transfer:

```python
# Inefficient - gets all jobs then filters locally
all_jobs = get_jobs({'limit': 50})
high_budget_jobs = [job for job in all_jobs if job['budget']['fixed_amount'] > 5000]

# Efficient - filters on server side
high_budget_jobs = get_jobs({
    'budget_min': 5000,
    'payment_verified': True,
    'limit': 20
})
```

## Monitoring usage

### Track Your Consumption
```python
class UsageTracker:
    def __init__(self):
        self.daily_calls = 0
        self.monthly_calls = 0
        self.last_reset = datetime.now()
    
    def record_call(self):
        self.daily_calls += 1
        self.monthly_calls += 1
        
        # Reset daily counter
        if datetime.now().date() > self.last_reset.date():
            self.daily_calls = 1
            self.last_reset = datetime.now()
    
    def get_usage_stats(self):
        return {
            'daily_calls': self.daily_calls,
            'monthly_calls': self.monthly_calls,
            'calls_remaining': self.plan_limit - self.monthly_calls
        }
```

### Set Usage Alerts
```python
def check_usage_alerts(usage_stats, plan_limit):
    usage_percentage = (usage_stats['monthly_calls'] / plan_limit) * 100
    
    if usage_percentage > 90:
        send_alert("Critical: 90% of monthly API calls used")
    elif usage_percentage > 75:
        send_alert("Warning: 75% of monthly API calls used")
    elif usage_percentage > 50:
        send_alert("Info: 50% of monthly API calls used")
```

## Scaling strategies

### When to Upgrade Your Plan

**Upgrade to Professional if:**
- You're consistently hitting Starter limits
- You need more than 1,000 calls per month
- You want faster rate limits (100/min vs 10/min)

**Upgrade to Business if:**
- You need more than 25,000 calls per month
- You want priority support and SLA
- You need historical data access
- You're building a commercial application

**Consider Enterprise if:**
- You need unlimited API calls
- You require custom rate limits
- You need on-premise deployment
- You want dedicated support

### Horizontal Scaling
For high-volume applications:

```python
import asyncio
import aiohttp
from concurrent.futures import ThreadPoolExecutor

class DistributedAPIClient:
    def __init__(self, api_keys):
        self.api_keys = api_keys
        self.key_index = 0
    
    def get_next_key(self):
        key = self.api_keys[self.key_index]
        self.key_index = (self.key_index + 1) % len(self.api_keys)
        return key
    
    async def make_distributed_request(self, params):
        # Use different API keys to multiply rate limits
        api_key = self.get_next_key()
        headers = {'X-API-KEY': api_key}
        
        async with aiohttp.ClientSession() as session:
            async with session.get(url, headers=headers, params=params) as response:
                return await response.json()
```

## Troubleshooting rate limits

### Common Issues

**Problem**: Getting rate limited despite low usage
- **Cause**: Burst requests exceeding per-minute limit
- **Solution**: Spread requests evenly over time

**Problem**: Inconsistent rate limit behavior  
- **Cause**: Multiple API keys or shared infrastructure
- **Solution**: Use single key per application, implement proper queuing

**Problem**: Rate limits seem lower than advertised
- **Cause**: Including failed requests in count
- **Solution**: Fix request errors, implement proper retry logic

### Debug Rate Limit Issues

```python
def debug_rate_limits(response):
    headers = response.headers
    print(f"Rate Limit: {headers.get('X-RateLimit-Limit')}")
    print(f"Remaining: {headers.get('X-RateLimit-Remaining')}")
    print(f"Reset Time: {headers.get('X-RateLimit-Reset')}")
    print(f"Burst Limit: {headers.get('X-RateLimit-Burst')}")
    
    if response.status_code == 429:
        retry_after = headers.get('Retry-After')
        print(f"Retry After: {retry_after} seconds")
```

---

**Need higher rate limits?**

[Upgrade Your Plan](/docs/pricing)
[Contact Sales](mailto:sales@upworkjobsapi.com?subject=Rate%20Limit%20Inquiry)

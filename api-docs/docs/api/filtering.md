---
sidebar_position: 3
title: Filtering & Search
---

# Filtering & Search

Find exactly the jobs you need with our powerful filtering system. Combine multiple filters to target specific opportunities, clients, and market segments.

## 🎯 Quick Filter Examples

### High-Value Verified Clients
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?payment_verified=true&budget_min=3000"
```

### Recent Web Development Jobs
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?category=web-development&posted_after=2024-09-20T00:00:00Z"
```

### Enterprise Clients in Tech
```bash
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?tags=react,typescript&buyer.total_spent_min=50000"
```

## 📋 Complete Filter Reference

### Basic Filters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `limit` | integer | Number of results (1-50, default 20) | `limit=10` |
| `payment_verified` | boolean | Client has verified payment method | `payment_verified=true` |
| `category` | string | Job category slug | `category=web-development` |
| `category_group` | string | Category group slug | `category_group=development-it` |
| `country` | string | Client country (2-letter code) | `country=US` |

### Budget Filters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `budget_min` | number | Minimum fixed budget | `budget_min=1000` |
| `budget_max` | number | Maximum fixed budget | `budget_max=10000` |
| `hourly_min` | number | Minimum hourly rate | `hourly_min=50` |
| `hourly_max` | number | Maximum hourly rate | `hourly_max=150` |

### Job Type & Status

| Parameter | Type | Description | Values |
|-----------|------|-------------|--------|
| `job_type` | integer | Contract type | `1`=Hourly, `2`=Fixed Price |
| `status` | integer | Job status | `1`=Open, `2`=Closed |
| `contractor_tier` | integer | Experience level | `1`=Entry, `2`=Intermediate, `3`=Expert |

### Time-Based Filters

| Parameter | Type | Description | Format |
|-----------|------|-------------|--------|
| `posted_after` | ISO8601 | Jobs posted after this date | `2024-09-20T00:00:00Z` |
| `posted_before` | ISO8601 | Jobs posted before this date | `2024-09-27T23:59:59Z` |

### Skills & Tags

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `tags` | string | Comma-separated required skills | `tags=react,typescript,aws` |
| `skills` | string | Alternative to tags | `skills=python,django` |

### Sorting Options

| Parameter | Values | Description |
|-----------|--------|-------------|
| `sort` | `posted_on_asc` | Oldest jobs first |
| `sort` | `posted_on_desc` | Newest jobs first (default) |
| `sort` | `last_visited_asc` | Least recently active clients |
| `sort` | `last_visited_desc` | Most recently active clients |
| `sort` | `budget_asc` | Lowest budget first |
| `sort` | `budget_desc` | Highest budget first |

## 🔍 Advanced Filtering Strategies

### Target High-Spending Clients
Find clients who regularly hire and spend significant amounts:

```bash
# Clients who spent $25k+ and hired 5+ freelancers
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?buyer.total_spent_min=25000&buyer.total_jobs_with_hires_min=5"
```

### Find Urgent Projects
Look for jobs with high activity and recent posting:

```bash
# Posted in last 24 hours with multiple applicants
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?posted_after=2024-09-26T00:00:00Z&client_activity.total_applicants_min=10"
```

### Niche Technology Stacks
Target specific technology combinations:

```bash
# React + Node.js + AWS projects
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?tags=react,nodejs,aws&job_type=2"
```

### Geographic Targeting
Focus on specific regions or time zones:

```bash
# US clients in PST timezone
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?country=US&location.timezone=America/Los_Angeles"
```

## 🎨 Filter Combinations

### Freelancer Lead Generation
```bash
# High-value, verified clients in your niche
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?payment_verified=true&budget_min=2000&category=web-development&contractor_tier=2"
```

### Agency Business Development
```bash
# Large projects from established clients
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?budget_min=10000&buyer.total_spent_min=50000&job_type=2"
```

### Market Research
```bash
# Recent trends in AI/ML hiring
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?tags=machine-learning,artificial-intelligence&posted_after=2024-09-01T00:00:00Z&sort=posted_on_desc"
```

### Competitive Analysis
```bash
# Monitor specific skill combinations
curl -H "X-API-KEY: your-key" \
  "https://api.upworkjobsapi.com/jobs?tags=python,tensorflow&budget_min=5000&sort=budget_desc"
```

## 📊 Filter Performance Tips

### Optimize Your Queries

**✅ Good Practices:**
- Use specific categories instead of broad searches
- Combine budget and verification filters for quality leads
- Set reasonable limits to avoid unnecessary data transfer
- Cache results for repeated queries

**❌ Avoid:**
- Overly broad searches without filters
- Requesting maximum limits unnecessarily
- Frequent identical queries (use caching)
- Complex regex patterns in text searches

### Rate Limit Management
- More specific filters = faster responses
- Cached results don't count against rate limits
- Batch similar queries when possible

## 🔧 Implementation Examples

### Python with Multiple Filters
```python
import requests

params = {
    'payment_verified': True,
    'budget_min': 3000,
    'category': 'web-development',
    'tags': 'react,typescript',
    'country': 'US',
    'limit': 20,
    'sort': 'posted_on_desc'
}

response = requests.get(
    'https://api.upworkjobsapi.com/jobs',
    headers={'X-API-KEY': 'your-key'},
    params=params
)
```

### JavaScript Dynamic Filtering
```javascript
const buildQuery = (filters) => {
  const params = new URLSearchParams();
  
  Object.entries(filters).forEach(([key, value]) => {
    if (value !== null && value !== undefined) {
      params.append(key, value);
    }
  });
  
  return params.toString();
};

const filters = {
  payment_verified: true,
  budget_min: 2000,
  category: 'design-creative',
  limit: 15
};

const queryString = buildQuery(filters);
const url = `https://api.upworkjobsapi.com/jobs?${queryString}`;
```

### PHP Filter Builder
```php
class UpworkJobFilter {
    private $filters = [];
    
    public function paymentVerified($verified = true) {
        $this->filters['payment_verified'] = $verified;
        return $this;
    }
    
    public function budgetRange($min, $max = null) {
        $this->filters['budget_min'] = $min;
        if ($max) $this->filters['budget_max'] = $max;
        return $this;
    }
    
    public function category($category) {
        $this->filters['category'] = $category;
        return $this;
    }
    
    public function build() {
        return http_build_query($this->filters);
    }
}

// Usage
$filter = new UpworkJobFilter();
$query = $filter->paymentVerified()
               ->budgetRange(1000, 5000)
               ->category('web-development')
               ->build();
```

## 🆘 Common Filter Issues

### Invalid Parameters
- **Problem**: Getting 400 errors with filter combinations
- **Solution**: Check parameter names and value formats
- **Example**: Use `payment_verified=true`, not `payment_verified=1`

### No Results
- **Problem**: Filters too restrictive, returning empty results
- **Solution**: Gradually remove filters to find the issue
- **Tip**: Start broad, then narrow down

### Slow Responses
- **Problem**: Complex queries taking too long
- **Solution**: Use more specific filters and lower limits
- **Optimization**: Cache frequently used filter combinations

---

**Need help optimizing your filters?**

[Contact Support](mailto:support@upworkjobsapi.com)
[View API Examples](/docs/api/endpoints)

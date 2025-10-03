# 🚀 Live Upwork Job Scraper - Get Fresh Job Listings Instantly

**The Ultimate Upwork Job Discovery Tool**

Stop manually browsing Upwork for hours! Get real-time job listings with complete client data, budgets, competition analysis, and market intelligence - all delivered instantly to your dashboard.

**Perfect for:**
- 💼 **Freelancers** seeking high-paying projects
- 🏢 **Agencies** managing multiple clients
- 📊 **Market Researchers** analyzing freelance trends
- 👔 **Business Owners** finding talent opportunities
- 🔍 **Job Seekers** discovering new opportunities

## ✨ What Makes This Different?

### 🎯 Complete Job Intelligence
- **Budget & Rate Analysis**: Exact budgets, hourly rates, and client spending history
- **Client Profiles**: Payment verification, location, total assignments, and hiring patterns
- **Competition Insights**: Applicant counts, interview invitations, and market demand
- **Real-Time Updates**: Fresh postings updated continuously
- **Skills & Categories**: Detailed skill requirements and category information

### ⚡ Instant Results
- **No Manual Browsing**: Skip hours of searching through Upwork pages
- **Bulk Data Export**: Get hundreds of jobs in seconds, not hours
- **Smart Filtering**: Focus on jobs that match your exact criteria
- **Rich Data Format**: CSV, JSON - use anywhere you need

### 🔗 URL-Based Search
- **Copy & Paste**: Just copy your Upwork search URL
- **All Filters Supported**: Any filter combination you use on Upwork
- **No Manual Configuration**: Filters are automatically extracted from the URL
- **Easy Updates**: Change your search criteria by updating the URL

## 💰 Incredibly Affordable Pricing

### 💸 Pay Only for Results
- **Ultra-Low Cost**: Just **$0.015 per job** (incredibly affordable!)
- **Example**: 100 jobs = **$1.50**
- **Bulk Value**: 1000 jobs = **$15.00**
- **No Hidden Fees**: Pay only for successfully retrieved jobs

### 🧾 Smart Billing
- ✅ **Only Successful Jobs Count** - Failed requests are FREE
- 🎯 **No Minimum Charges** - Start with just 1 job
- 📈 **Transparent Costs** - See exactly what you're paying for
- 🎁 **Volume Discounts** - More jobs = better value

## 📦 What You Get

### 📋 Complete Job Data
Every job includes:
- **Job Details**: Title, description, budget, hourly rates
- **Client Intelligence**: Payment status, spending history, location, hiring patterns
- **Competition Metrics**: Number of applicants, interviews, proposals
- **Timing Data**: Posted dates, last activity, relative timestamps
- **Requirements**: Skills, experience level, project duration

## 🎬 Simple Setup

Just 3 easy steps:

1. 🔍 **Search on Upwork**: Go to [Upwork Jobs](https://www.upwork.com/nx/search/jobs/) and set your filters
2. 📋 **Copy the URL**: Copy the full URL from your browser (includes all your filters)
3. ▶️ **Paste & Run**: Paste the URL into the actor input and click Start!

### 📝 Example URLs

**🐍 Python Developer Jobs (Hourly, $50+/hr, Payment Verified)**
```
https://www.upwork.com/nx/search/jobs/?q=python%20developer&job_type=hourly&hourly_rate=50-&payment_verified=1
```

**🌐 Expert Web Developers (Fixed Price, $1000+ Budget)**
```
https://www.upwork.com/nx/search/jobs/?q=web%20development&contractor_tier=3&amount=1000-
```

**🇺🇸 US Clients Only (Any Job Type)**
```
https://www.upwork.com/nx/search/jobs/?client_country=US&payment_verified=1
```

**⚛️ Recent React Jobs (Last 24 hours)**
```
https://www.upwork.com/nx/search/jobs/?q=react&sort=recency
```

## 🔧 Supported URL Parameters

The actor automatically extracts and converts these Upwork URL parameters:

### 🔎 Search & Filters
- `q` - Search keywords (e.g., "python developer")
- `payment_verified` - Client payment verification (1 = verified)
- `job_type` or `t` - Job type (0 = hourly, 1 = fixed-price)
- `contractor_tier` - Experience level (1 = entry, 2 = intermediate, 3 = expert)
- `job_status` - Job status (open, closed)

### 🌍 Location
- `client_country` or `location` - Client country code (e.g., US, GB, CA)

### 💵 Budget & Rates
- `amount` - Fixed budget range (e.g., "1000-5000" or "500-")
- `hourly_rate` - Hourly rate range (e.g., "25-100" or "50-")

### 👤 Client Filters
- `client_hires` - Client hiring history range (e.g., "1-9", "10-")
- `client_total_spent` - Client total spend range
- `client_total_feedback` - Client feedback score

### 🏷️ Skills & Tags
- `skills` - Required skills (comma-separated)
- `tags` - Job tags

### 📅 Job Details
- `duration_v3` - Project duration (week, month, semester, ongoing)
- `workload` - Workload requirements
- `contract_to_hire` - Contract-to-hire opportunities (1 = yes)

### 🔄 Sorting
- `sort` - Sort order (recency, relevance, client_rating, duration)

## 📄 Sample Output

Here's what you get for each job:

```json
{
  "id": "1971966031232790280",
  "title": "Customer Service / Operations Manager",
  "description": "We're seeking someone who communicates with empathy...",
  "job_type": "fixed-price",
  "status": "open",
  "contractor_tier": "expert",
  "posted_on": "2025-09-27T15:51:54Z",
  "category": {
    "name": "Customer Service & Tech Support",
    "slug": "customer-service-tech-support",
    "group": "Customer Service",
    "group_slug": "customer-service"
  },
  "budget": {
    "fixed_amount": 0,
    "currency": "USD"
  },
  "buyer": {
    "payment_verified": true,
    "country": "UNITED STATES",
    "city": "Austin",
    "timezone": "America/Chicago (UTC-05:00)",
    "total_spent": 32812.32,
    "total_assignments": 21,
    "total_jobs_with_hires": 17
  },
  "tags": ["jsi_contractToHire", "contractToHireSet", "searchable"],
  "url": "https://www.upwork.com/freelance-jobs/apply/...",
  "last_visited_at": "2025-09-28T00:54:39Z",
  "posted_on_relative": "7 hours ago",
  "duration_label": "More than 6 months",
  "engagement": "Less than 30 hrs/week",
  "skills": ["Customer Service", "Operations Management"],
  "hourly_budget": {
    "min": 15.0,
    "max": 30.0,
    "currency": "USD"
  },
  "client_activity": {
    "total_applicants": 1,
    "total_hired": 0,
    "total_invited_to_interview": 0,
    "unanswered_invites": 0,
    "invitations_sent": 0,
    "last_buyer_activity": "2025-09-27T15:58:47.031Z"
  },
  "location": {
    "country": "United States",
    "city": "Austin",
    "timezone": "America/Chicago (UTC-05:00)"
  },
  "is_private": false,
  "privacy_reason": "",
  "scraped_at": "2025-09-27T23:00:05Z"
}
```


## 🏆 Why Choose This Actor?

### 🚀 Lightning Fast
- **Get results in seconds** - no more waiting around
- **High-performance system** for maximum speed
- **Real-time data** - always up-to-date information

### 💎 Premium Data Quality
- **Complete job intelligence** - budgets, client history, competition metrics
- **Structured format** - clean, consistent data you can rely on
- **Always fresh** - continuously updated job listings

### ⚡ Easy to Use
- **Simple URL input** - just copy and paste from Upwork
- **All filters supported** - any combination you can use on Upwork
- **Ready-to-export** - perfect for spreadsheets and analysis

---

**Ready to get started?** 🎉 Simply copy your Upwork search URL, paste it into the actor, and get fresh, structured job data in seconds!

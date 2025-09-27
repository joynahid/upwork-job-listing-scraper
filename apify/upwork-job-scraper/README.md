# Live Upwork Job Scraper

🚀 **Real-time Upwork job listings scraper with advanced bot detection evasion, Cloudflare bypass, and comprehensive data extraction.**

Perfect for freelancers, agencies, and job market research. Get fresh job postings with client details, budgets, skills, and competition analysis.

## 📊 What You Get

This actor retrieves comprehensive job data from Upwork including:

- **Job Details**: Title, description, budget, hourly rates
- **Client Information**: Company size, industry, location, spending history
- **Project Specs**: Duration, experience level, skills required
- **Market Intelligence**: Proposal counts, competition analysis
- **Fresh Data**: Real-time job postings updated continuously

## 💰 Pricing & Usage

### **Cost-Effective Pricing**
- **Rate**: `0.001 credits per job processed`
- **Example**: 100 jobs = 0.1 credits (~$0.01)
- **Maximum**: 1000 jobs per run = 1 credit

### **What Counts as Usage**
- Each job record retrieved from our API counts as 1 unit
- You're only charged for jobs actually processed and returned
- Failed requests or empty results don't count toward usage

### **Usage Tracking**
```
Usage tracked: 50 jobs (Charged for 50 jobs processed)
💰 Total cost: 0.05 credits
```

## ⚡ How It Works

### **Data Source**
This actor connects to our **live Upwork scraping API** that:
- Runs 24/7 with advanced bot detection evasion
- Bypasses Cloudflare protection automatically  
- Maintains fresh job data with real-time updates
- Uses residential proxies and human-like behavior

### **Retrieval Process**
1. **API Connection**: Connects to `upworkjobscraperapi.nahidhq.com`
2. **Authentication**: Uses secure API key authentication
3. **Data Fetch**: Retrieves up to your specified job limit
4. **Processing**: Transforms raw data into structured format
5. **Output**: Saves to Apify dataset with friendly UI

### **Data Freshness**
- Jobs are scraped continuously from Upwork
- API updates every few minutes with new postings
- You get the most recent jobs available at runtime
- Last updated timestamp included in API response

## ⏱️ Performance & Delays

### **Expected Runtime**
- **Small runs (1-50 jobs)**: 10-30 seconds
- **Medium runs (51-200 jobs)**: 30-60 seconds  
- **Large runs (201-1000 jobs)**: 1-3 minutes

### **Potential Delays**
- **API Response Time**: 2-10 seconds (depends on job count)
- **Data Processing**: ~0.1 seconds per job
- **Network Latency**: 1-5 seconds (geographic location)
- **Apify Platform**: 5-15 seconds (container startup)

### **Factors Affecting Speed**
- **Job Count**: More jobs = longer processing time
- **API Load**: High demand may cause slight delays
- **Data Complexity**: Rich job data takes time to process
- **Network Conditions**: Internet speed affects API calls

## 🔧 Configuration

### **Input Parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `maxJobs` | Integer | 50 | Maximum number of jobs to retrieve (1-1000) |

### **Example Configuration**
```json
{
  "maxJobs": 100
}
```

## 📋 Output Format

### **Dataset Schema**
Each job record contains:

```json
{
  "job_id": "Virtual-Assistant-Needed_~021971791858643452240",
  "title": "Virtual Assistant Needed for Research",
  "description": "We are seeking an experienced virtual assistant...",
  "url": "https://www.upwork.com/jobs/...",
  "hourly_rate": "$25.00 - $47.00/hr",
  "budget": "$400.00",
  "experience_level": "Intermediate",
  "job_type": "Fixed-price",
  "skills": ["Virtual Assistance", "Research", "Data Entry"],
  "client_location": "United States",
  "client_company_size": "Mid-sized company (10-99 people)",
  "client_industry": "Health & Fitness",
  "posted_date": "Posted 11 minutes ago",
  "proposals_count": "Less than 5",
  "duration": "1-3 months",
  "project_type": "One-time project",
  "work_hours": "Less than 30 hrs/week",
  "member_since": "Oct 20, 2015",
  "total_spent": "$23K",
  "total_hires": "59",
  "last_visited_at": "2025-09-27T04:23:28Z",
  "scraped_at": "2025-09-27T05:33:52Z",
  "raw_data": { /* Complete API response */ }
}
```

### **Output Views**
- **Overview**: Quick summary with key job details
- **Detailed**: Comprehensive view with all available data
- **Export**: Download as CSV, Excel, or JSON

## 🛡️ Reliability & Compliance

### **Anti-Detection Features**
- Advanced browser fingerprint randomization
- Residential proxy rotation
- Human-like interaction patterns
- Cloudflare bypass technology
- Session persistence

### **Data Quality**
- Schema validation for all records
- Error handling with graceful fallbacks
- Duplicate prevention
- Rich parsing of budgets, skills, and metadata

### **Compliance**
- Respects Upwork's terms of service
- No aggressive scraping or rate limit violations
- Ethical data collection practices
- Transparent usage tracking

## 🚨 Important Notes

### **Rate Limits**
- Maximum 1000 jobs per run
- Recommended: 50-200 jobs for optimal performance
- Higher limits may experience longer delays

### **Data Availability**
- Jobs depend on current Upwork listings
- Some jobs may become unavailable during processing
- Empty results possible during low activity periods

### **API Dependencies**
- Requires active API service at `upworkjobscraperapi.nahidhq.com`
- Service availability: 99.9% uptime
- Automatic failover and retry mechanisms

## 📞 Support

### **Common Issues**
- **Empty Results**: Try reducing `maxJobs` or running at different times
- **Slow Performance**: Large job counts take longer to process
- **API Errors**: Temporary service issues, retry in a few minutes

### **Contact**
- **Technical Issues**: Check actor logs for detailed error messages
- **API Status**: Monitor service health via API endpoint
- **Feature Requests**: Submit via Apify Console feedback

## 🔄 Version History

- **v0.1.7**: Fixed dataset schema structure, improved UI presentation
- **v0.1.6**: Enhanced field mapping, added client industry data
- **v0.1.5**: Updated schema to match real API structure
- **v0.1.4**: Added comprehensive dataset schema
- **v0.1.3**: Reduced usage charges to 0.001 per job
- **v0.1.2**: Initial release with basic functionality

---

**Ready to get started?** Configure your job limit and run the actor to get fresh Upwork job listings in seconds! 🚀

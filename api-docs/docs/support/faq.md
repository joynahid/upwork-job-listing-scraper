---
sidebar_position: 1
title: Frequently Asked Questions
---

# Frequently Asked Questions

Get quick answers to common questions about our Upwork Jobs API service.

## 🚀 Getting Started

### How do I get started with the API?
1. [Sign up for a free trial](mailto:sales@upworkjobsapi.com?subject=Free%20Trial%20Signup) - no credit card required
2. Receive your API key via email within 5 minutes
3. Make your first API call using our [Getting Started guide](/docs/getting-started)
4. Explore the data and upgrade to a paid plan when ready

### Do I need a credit card for the free trial?
No! Our free trial includes 1,000 API calls with no credit card required. You only need to provide payment information when upgrading to a paid plan.

### How quickly can I get access?
API keys are generated automatically and sent via email within 5 minutes of signup. You can start making API calls immediately.

## 💰 Pricing & Plans

### What's included in the free plan?
- 1,000 API calls per month
- Access to all job data fields
- Basic filtering options
- Email support
- 7-day data retention

### Can I upgrade or downgrade my plan?
Yes! You can change plans anytime. Upgrades take effect immediately, and downgrades take effect at your next billing cycle. All changes are prorated.

### Do you offer annual discounts?
Yes! Save 20% on Professional plans and 25% on Business plans with annual billing.

### What happens if I exceed my plan limits?
- **Professional/Business**: Pay $0.01 per additional API call
- **Starter**: API calls are blocked until next month or upgrade
- **Enterprise**: Custom overage rates based on your agreement

## 🔧 Technical Questions

### What data format does the API return?
All responses are in JSON format with consistent structure. See our [API documentation](/docs/api/endpoints) for detailed response schemas.

### How fresh is the job data?
Job data is updated continuously throughout the day. Most new postings appear in our API within 15-30 minutes of being posted on Upwork.

### Can I get historical data?
- **Starter**: 7 days of data retention
- **Professional**: 30 days of data retention  
- **Business**: 6 months of historical data
- **Enterprise**: Full historical access available

### What's your API uptime?
We maintain 99.9% uptime with monitoring and alerts. Enterprise customers can get custom SLA agreements with higher guarantees.

### Do you have rate limits?
Yes, to ensure fair usage:
- **Starter**: 10 requests per minute
- **Professional**: 100 requests per minute
- **Business**: 500 requests per minute
- **Enterprise**: Custom limits based on needs

## 🔒 Security & Compliance

### How secure is my data?
- All API calls use HTTPS encryption
- API keys are encrypted at rest
- We're SOC 2 Type II compliant
- Regular security audits and monitoring

### Do you store my API requests?
We log basic request metadata (timestamp, endpoint, response code) for monitoring and support purposes. We don't store the actual response data you receive.

### Is the API GDPR compliant?
Yes, we follow GDPR guidelines for data handling and provide data processing agreements for Enterprise customers.

### Can I use this data commercially?
Yes, all our plans allow commercial use of the data. Please respect Upwork's terms of service and applicable data protection laws.

## 🎯 Use Cases & Integration

### What can I build with this API?
Popular use cases include:
- Lead generation tools for freelancers and agencies
- Market research and trend analysis
- Job recommendation engines
- Competitive intelligence dashboards
- Automated client prospecting systems

### Do you provide SDKs or libraries?
We provide code examples in Python, JavaScript, PHP, and cURL. Full SDKs are available for Enterprise customers.

### Can I integrate with my CRM?
Yes! Many customers integrate our API with Salesforce, HubSpot, and other CRM systems. We can provide integration guidance.

### Do you support webhooks?
Webhook notifications are available on Business and Enterprise plans. Get real-time alerts when new jobs match your criteria.

## 📊 Data Quality & Coverage

### How accurate is the job data?
We maintain 99.5% data accuracy through automated validation and spam filtering. All job postings are verified before inclusion in our API.

### What job categories are covered?
We cover all Upwork categories including:
- Development & IT
- Design & Creative
- Writing & Translation
- Sales & Marketing
- Admin & Customer Support
- Finance & Accounting
- Engineering & Architecture
- Legal services

### Do you filter out spam or low-quality jobs?
Yes! Our advanced filtering removes:
- Duplicate postings
- Spam and fake jobs
- Jobs with suspicious client activity
- Postings that violate Upwork's terms

### How many jobs are available daily?
We process thousands of new job postings daily. The exact number varies based on market activity and your filtering criteria.

## 🛠️ Troubleshooting

### I'm getting 401 authentication errors
- Check that your API key is correct
- Ensure you're using the header name `X-API-KEY`
- Verify there are no extra spaces or characters
- Make sure your key hasn't expired

### My requests are being rate limited
- Check your current plan's rate limits
- Implement exponential backoff in your code
- Consider upgrading to a higher plan
- Cache responses to reduce API calls

### I'm not getting any results
- Verify your filters aren't too restrictive
- Check that parameter names and values are correct
- Try a broader search first, then narrow down
- Ensure you're using the correct date format for time filters

### The API is responding slowly
- Use more specific filters to reduce response size
- Lower your limit parameter if you don't need many results
- Check if you're making too many concurrent requests
- Consider caching frequently requested data

## 📞 Support & Contact

### How do I get technical support?
- **Email**: [support@upworkjobsapi.com](mailto:support@upworkjobsapi.com)
- **Response Time**: Within 24 hours (faster for paid plans)
- **Business/Enterprise**: Priority support with faster response times

### Can I schedule a demo or consultation?
Yes! [Contact our sales team](mailto:sales@upworkjobsapi.com?subject=Demo%20Request) to schedule a personalized demo or discuss your specific use case.

### Do you offer custom development?
Enterprise customers can request custom features, integrations, and data processing. Contact us to discuss your requirements.

### How do I report a bug or issue?
Email [support@upworkjobsapi.com](mailto:support@upworkjobsapi.com) with:
- Description of the issue
- Your API key (last 4 characters only)
- Example request that's causing problems
- Expected vs actual behavior

## 💡 Best Practices

### How can I optimize my API usage?
- Use specific filters to get targeted results
- Cache responses when appropriate
- Implement proper error handling and retries
- Monitor your usage to avoid unexpected overages

### What's the best way to handle rate limits?
- Implement exponential backoff
- Spread requests over time instead of bursts
- Use webhooks for real-time updates instead of polling
- Cache frequently accessed data

### How should I store the API data?
- Follow data protection regulations in your jurisdiction
- Don't store personal information unnecessarily
- Implement proper data retention policies
- Ensure secure storage and access controls

---

**Still have questions?**

[Contact Support](mailto:support@upworkjobsapi.com)
[Schedule Consultation](mailto:sales@upworkjobsapi.com?subject=Consultation%20Request)

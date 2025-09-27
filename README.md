# Upwork Job Listing Scraper

An advanced Apify actor for scraping Upwork job listings with sophisticated bot detection evasion and structured data extraction.

## 🚀 Features

- **Advanced Bot Detection Evasion**: Uses CamoufoxPlugin with anti-fingerprinting technology
- **Structured Search Parameters**: Build search queries with filters for budget, location, experience level, etc.
- **Cloudflare Challenge Handling**: Automatic detection and solving of Cloudflare Turnstile challenges
- **Detailed Job Information**: Optional extraction of comprehensive job details from individual job pages
- **Rich Data Parsing**: Extract structured data with value objects for budgets, skills, time posted, etc.
- **Apify Integration**: Full Apify platform integration with datasets, key-value store, and monitoring

## 📋 Input Parameters

### Search Configuration

- **Search Parameters**: Structured filters for building Upwork search queries
  - Keywords (e.g., "python automation")  
  - Budget range (min/max in USD)
  - Hourly rate range (min/max in USD)
  - Experience level (entry-level, intermediate, expert)
  - Location filters (worldwide, Americas, Europe, etc.)
  - Payment verification requirement
  - Sort order (recency, relevance, budget, client rating)

- **Custom Search URLs**: Use custom Upwork search URLs instead of structured parameters

### Processing Options

- **Extract Details**: Enable detailed job information extraction (slower but comprehensive)
- **Max Jobs**: Maximum number of jobs to process (1-1000, default: 50)
- **Rate Limiting**: Configurable delays between requests (1-60 seconds)

### Debug & Output Options

- **Take Screenshots**: Capture screenshots for debugging
- **Debug Mode**: Enable verbose logging
- **Output Format**: Choose between basic or enhanced output
- **Include Raw Data**: Include original scraped data alongside parsed data

## 📊 Output Data

### Basic Job Data

```json
{
  "type": "basic_job",
  "title": "Python Automation Specialist",
  "description": "Build automated workflows...",
  "budget_raw": "$500-1000",
  "budget": {
    "amount": 500,
    "currency": "USD", 
    "type": "fixed"
  },
  "skills": ["Python", "Automation", "Web Scraping"],
  "posted_time": {
    "value": "2 hours ago",
    "is_fresh": true,
    "urgency_level": "Fresh"
  },
  "proposals": {
    "count": 5,
    "competition_level": "Low competition"
  },
  "url": "https://www.upwork.com/jobs/...",
  "job_id": "123456789"
}
```

### Enhanced Job Data (when extract_details=true)

Includes all basic data plus:

```json
{
  "detailed": {
    "job_id": "123456789",
    "experience_level": "Intermediate",
    "job_type": "Fixed Price",
    "client_info": {
      "name": "Tech Company",
      "location": "United States",
      "rating": "4.8/5",
      "total_spent": "$50,000+",
      "hire_rate": "85%"
    }
  }
}
```

## 🔧 Technical Implementation

### Architecture

- **Crawlee Framework**: Uses Crawlee for Python with Playwright crawler
- **CamoufoxPlugin**: Advanced browser with anti-fingerprinting capabilities  
- **Value Objects**: Domain-driven design with rich data modeling
- **Pydantic Validation**: Type-safe input validation and serialization

### Anti-Detection Features

- **Browser Fingerprint Randomization**: Camoufox handles user agent rotation, viewport randomization
- **Residential Proxy Support**: Optional Apify residential proxy integration
- **Human-like Behavior**: Randomized delays, mouse movements, and interaction patterns
- **Cloudflare Handling**: Automatic detection and solving of Turnstile challenges
- **Session Management**: Persistent browser sessions to avoid repeated challenges

### Data Quality

- **Schema Validation**: All data is validated against Pydantic models
- **Error Handling**: Graceful fallbacks for parsing errors
- **Duplicate Prevention**: URL-based deduplication
- **Rich Parsing**: Extract budgets, skills, time posted, and competition levels

## 🚦 Usage Examples

### Basic Job Search

```json
{
  "searchParameters": {
    "keywords": "python automation",
    "minHourlyRate": 25,
    "paymentVerified": true,
    "location": "Americas"
  },
  "maxJobs": 20
}
```

### Advanced Search with Details

```json
{
  "searchParameters": {
    "keywords": "machine learning",
    "minBudget": 1000,
    "experienceLevel": "expert",
    "location": "worldwide"
  },
  "extractDetails": true,
  "maxJobs": 10,
  "debugMode": true
}
```

### Custom URL Search

```json
{
  "customSearchUrls": [
    "https://www.upwork.com/nx/search/jobs/?q=python&sort=recency"
  ],
  "takeScreenshots": true
}
```

## 🔒 Privacy & Compliance

- Uses residential proxies to respect rate limits
- Implements human-like browsing patterns
- Respects robots.txt and website terms of service
- No personal data collection beyond public job listings

## 📈 Performance

- **Concurrent Processing**: Handles multiple jobs simultaneously
- **Smart Caching**: Reduces redundant requests
- **Efficient Extraction**: JavaScript-based DOM parsing
- **Resource Management**: Automatic cleanup and memory optimization

## 🛠 Development

Built with:
- **Apify SDK** for actor infrastructure
- **Crawlee** for web crawling
- **Camoufox** for advanced browser automation
- **Pydantic** for data validation
- **Playwright** for browser control

## 🐳 Docker Usage

### Quick Start with Docker Compose

1. **Setup Environment**:
   ```bash
   cp env.example .env
   # Edit .env with your Apify token and configuration
   ```

2. **Build and Run**:
   ```bash
   # Production mode
   make build && make up
   
   # Development mode (with live code reload)
   make dev
   ```

3. **View Logs**:
   ```bash
   make logs
   ```

### Available Docker Commands

- `make build` - Build the Docker image
- `make up` - Start services in production mode
- `make dev` - Start services in development mode with live reload
- `make down` - Stop and remove containers
- `make logs` - Show container logs
- `make clean` - Clean up containers, images, and volumes
- `make shell` - Open shell in running container
- `make setup-dev` - Setup development environment

### Environment Variables

Copy `env.example` to `.env` and configure:

- `APIFY_TOKEN` - Your Apify API token
- `PROXY_URL` - Upstream proxy URL (e.g. Squid or direct provider)
- `APIFY_LOG_LEVEL` - Logging level (INFO, DEBUG)
- `HEADLESS` - Run browser in headless mode (true/false)

### Development Setup

For development with live code reload:

```bash
# Setup development environment
make setup-dev

# Start development server
make dev

# The container will automatically reload when you change source code
```

## 📞 Support

For issues, feature requests, or questions:
1. Check the actor logs for detailed error information
2. Use debug mode for troubleshooting
3. Review the input schema validation errors
4. Contact support through Apify Console

## 📄 License

This actor is designed for legitimate business use cases such as market research, competitive analysis, and job market insights. Users are responsible for compliance with Upwork's terms of service and applicable laws.

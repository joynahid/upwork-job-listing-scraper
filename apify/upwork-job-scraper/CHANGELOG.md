# Changelog

## Version 0.2 - URL Parser Update

### Major Changes

#### Simplified Input Schema
- **BREAKING CHANGE**: Input schema now requires only `upworkUrl` and optional `maxJobs`
- Removed all individual filter fields (50+ fields reduced to 2)
- Users now paste their Upwork search URL directly

#### New URL Parser Module
- Added `src/url_parser.py` - Comprehensive URL parser for Upwork search URLs
- Automatically extracts and converts URL parameters to Go API format
- Supports all Upwork filters:
  - Search keywords (`q`)
  - Job types (`job_type`, `t`)
  - Experience levels (`contractor_tier`)
  - Budget ranges (`amount`, `hourly_rate`)
  - Client filters (`client_country`, `client_hires`, etc.)
  - Skills and tags
  - Sorting options
  - And many more...

#### Updated Configuration
- `src/config.py` - Refactored to use URL parser
- Simplified validation logic
- Better error messages when URL is invalid

#### Streamlined API Wrapper
- `src/api_wrapper.py` - Simplified filter handling
- Removed complex filter mapping (now handled by URL parser)
- Filters are passed directly to Go API in correct format

#### Updated Documentation
- `README.md` - Complete rewrite focusing on URL-based approach
- Added usage examples with real URLs
- New `USAGE_EXAMPLES.md` with 10+ real-world examples
- Clear instructions on how to get Upwork URLs

### Migration Guide

**Old Way (v0.1)**:
```json
{
    "maxJobs": 20,
    "paymentVerified": true,
    "jobType": "hourly",
    "hourlyMin": 50,
    "contractorTier": "intermediate",
    "country": "US"
}
```

**New Way (v0.2)**:
```json
{
    "upworkUrl": "https://www.upwork.com/nx/search/jobs/?payment_verified=1&job_type=hourly&hourly_rate=50-&contractor_tier=2&client_country=US",
    "maxJobs": 20
}
```

### Benefits

1. **Easier to Use**: Just copy URL from Upwork, no manual field mapping
2. **More Powerful**: Access ALL Upwork filters, not just predefined ones
3. **Less Error-Prone**: Visual confirmation on Upwork before copying URL
4. **Future-Proof**: New Upwork filters automatically supported
5. **Faster Setup**: Reduce configuration time from minutes to seconds

### Technical Details

- URL parsing handles URL encoding automatically
- Supports both numeric codes and string labels for enums
- Range parsing for budgets and rates (e.g., "50-100" or "50-")
- Comma-separated values for multiple selections
- Boolean conversion for toggle filters

### Files Changed

- `.actor/input_schema.json` - New simplified schema
- `.actor/INPUT_SCHEMA.json` - New simplified schema (uppercase)
- `.actor/actor.json` - Updated description and version
- `.actor/INPUT.json` - New example input
- `src/url_parser.py` - **NEW FILE**
- `src/config.py` - Refactored
- `src/api_wrapper.py` - Simplified
- `README.md` - Complete rewrite
- `USAGE_EXAMPLES.md` - **NEW FILE**
- `CHANGELOG.md` - **NEW FILE**

### Compatibility

- ✅ Go API: Fully compatible
- ✅ Apify Dataset: Same output format
- ✅ Environment Variables: No changes required
- ❌ Input Schema: **Breaking change** - requires URL-based input

### Notes

- Old actor runs with v0.1 input format will need to be updated
- API endpoint and key requirements unchanged
- Output format remains identical
- All existing integrations downstream should work without changes


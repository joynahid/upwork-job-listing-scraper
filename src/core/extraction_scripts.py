"""JavaScript extraction scripts for comprehensive Upwork job data extraction."""

# Comprehensive job extraction script for job listings pages
JOB_LISTING_EXTRACTION_SCRIPT = """
{
    const jobs = [];

    // Find job listing elements on search pages
    const jobSelectors = [
        '[data-test="job-tile"]',
        '[data-cy="job-tile"]',
        '.job-tile',
        '.up-card-section',
        'article',
        '[class*="job"][class*="card"]',
        'div[data-v-]'
    ];

    let jobElements = [];
    let selectedSelector = null;
    for (const selector of jobSelectors) {
        jobElements = document.querySelectorAll(selector);
        if (jobElements.length > 0) {
            selectedSelector = selector;
            break;
        }
    }

    if (jobElements.length === 0) {
        const patterns = ['div > div > div', 'section', '[role="article"]'];
        for (const pattern of patterns) {
            const elements = document.querySelectorAll(pattern);
            if (elements.length >= 3) {
                jobElements = elements;
                selectedSelector = pattern;
                break;
            }
        }
    }

    const normalizeText = (text) => text ? text.replace(/\\s+/g, ' ').trim() : '';

    jobElements.forEach((jobElement, index) => {
        try {
            // Extract URL first to determine if this is a valid job
            let jobUrl = '';
            const titleSelectors = ['h2 a', 'h3 a', '.job-tile-title a', '[data-test="job-title"] a', '.up-n-link'];
            for (const selector of titleSelectors) {
                const link = jobElement.querySelector(selector);
                if (link && link.href) {
                    jobUrl = link.href;
                    break;
                }
            }

            // Only process if we have a valid job URL
            if (jobUrl && jobUrl.includes('/jobs/')) {
                jobs.push({
                    index: index + 1,
                    url: jobUrl,
                    // We'll get full details when we visit the individual job page
                    extracted_from: 'listing_page'
                });
            }
        } catch (error) {
            // Ignore individual job extraction errors
        }
    });

    return jobs;
}
"""

with open("src/core/extract_data.js") as file:
    JOB_DETAIL_SCRIPT = file.read()

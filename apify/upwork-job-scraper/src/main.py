"""Upwork Job Scraper API Wrapper for Apify.

This Actor connects to a Go API backend to fetch Upwork job listings in real-time.
It processes jobs one by one and saves them to the Apify dataset with proper formatting.
"""

from __future__ import annotations

import asyncio
import os
from datetime import datetime
from typing import Any, Dict, List

import httpx
from apify import Actor


class UpworkJobAPIWrapper:
    """Wrapper class for the Upwork Job Go API."""

    def __init__(self, api_endpoint: str, api_key: str, debug_mode: bool = False):
        """Initialize the API wrapper.

        Args:
            api_endpoint: The Go API endpoint URL
            api_key: API key for authentication
            debug_mode: Enable debug logging
        """
        self.api_endpoint = api_endpoint.rstrip("/")
        self.api_key = api_key
        self.debug_mode = debug_mode

        # Create HTTP client with timeout
        self.client = httpx.AsyncClient(
            timeout=30.0, headers={"X-API-KEY": self.api_key}
        )

    async def fetch_jobs(self, max_jobs: int, filters: Dict[str, Any] = None) -> Dict[str, Any]:
        """Fetch jobs from the Go API.

        Args:
            max_jobs: Maximum number of jobs to fetch
            filters: Optional filters to apply to the job search

        Returns:
            API response containing job data

        Raises:
            httpx.HTTPError: If API request fails
        """
        jobs_url = f"{self.api_endpoint}/jobs"
        params = {"limit": max_jobs} if max_jobs else {}
        
        # Add filters to params if provided
        if filters:
            # Map filter names to API parameter names
            filter_mapping = {
                "paymentVerified": "payment_verified",
                "categoryGroup": "category_group",
                "jobType": "job_type",
                "contractorTier": "contractor_tier",
                "postedAfter": "posted_after",
                "postedBefore": "posted_before",
                "budgetMin": "budget_min",
                "budgetMax": "budget_max"
            }
            
            for filter_key, filter_value in filters.items():
                if filter_value is not None:
                    api_param = filter_mapping.get(filter_key, filter_key)
                    if filter_key == "tags" and isinstance(filter_value, list):
                        params[api_param] = ",".join(filter_value)
                    else:
                        params[api_param] = filter_value

        if self.debug_mode:
            Actor.log.info(f"🔍 Fetching jobs from: {jobs_url} (limit: {max_jobs})")
            if params:
                Actor.log.info(f"🎯 Filters: {params}")

        try:
            response = await self.client.get(jobs_url, params=params)
            response.raise_for_status()

            data = response.json()

            if not data.get("success", False):
                raise Exception(
                    f"API returned error: {data.get('message', 'Unknown error')}"
                )

            if self.debug_mode:
                Actor.log.info(f"✅ Successfully fetched {data.get('count', 0)} jobs")

            return data

        except httpx.HTTPError as e:
            Actor.log.error(f"❌ HTTP error fetching jobs: {e}")
            raise
        except Exception as e:
            Actor.log.error(f"❌ Error fetching jobs: {e}")
            raise

    async def health_check(self) -> bool:
        """Check if the API is healthy.

        Returns:
            True if API is healthy, False otherwise
        """
        health_url = f"{self.api_endpoint}/health"

        try:
            response = await self.client.get(health_url)
            response.raise_for_status()

            data = response.json()
            is_healthy = data.get("success", False)

            if self.debug_mode:
                status = "✅ healthy" if is_healthy else "❌ unhealthy"
                Actor.log.info(f"API health check: {status}")

            return is_healthy

        except Exception as e:
            if self.debug_mode:
                Actor.log.error(f"❌ Health check failed: {e}")
            return False

    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()


async def process_jobs_simple(
    jobs: List[Dict[str, Any]], debug_mode: bool, max_jobs: int
) -> int:
    """Process jobs simply - save them to Apify dataset exactly as received from Go API.

    Args:
        jobs: List of job data from API (JobDTO format)
        debug_mode: Enable debug logging
        max_jobs: Maximum number of jobs to process

    Returns:
        Number of jobs processed
    """
    processed_count = 0
    total_jobs = len(jobs)

    Actor.log.info(f"🚀 Processing {total_jobs} jobs from API")
    
    total_jobs_to_process = min(max_jobs, total_jobs)

    for job in jobs[:total_jobs_to_process]:
        try:
            # Pass through the job data exactly as received from Go API
            # The Go API returns JobDTO objects that match our schema
            output_job = job.copy()  # Make a copy to avoid modifying original
            
            # Add scraped timestamp
            output_job["scraped_at"] = datetime.now().isoformat()

            # Push to Apify dataset with pay-per-event charging
            await Actor.push_data(output_job, 'api-result')

            processed_count += 1

            # Log progress
            job_title = output_job.get('title', 'Unknown Title')
            Actor.log.info(
                f"✅ Saved job {processed_count}: {job_title[:50]}..."
            )

            if debug_mode:
                Actor.log.info(f"   📊 Job ID: {output_job.get('id', 'N/A')}")
                Actor.log.info(f"   🏷️ Job Type: {output_job.get('job_type', 'N/A')}")
                Actor.log.info(f"   🎯 Contractor Tier: {output_job.get('contractor_tier', 'N/A')}")
                Actor.log.info(f"   🔒 Private: {output_job.get('is_private', False)}")
                
                # Log budget info
                budget = output_job.get('budget', {})
                if budget and budget.get('fixed_amount'):
                    Actor.log.info(f"   💰 Budget: ${budget.get('fixed_amount')} {budget.get('currency', 'USD')}")
                
                # Log hourly info
                hourly = output_job.get('hourly_budget', {})
                if hourly and (hourly.get('min') or hourly.get('max')):
                    min_rate = hourly.get('min', 0)
                    max_rate = hourly.get('max', 0)
                    currency = hourly.get('currency', 'USD')
                    Actor.log.info(f"   💵 Hourly: ${min_rate}-${max_rate} {currency}")

        except Exception as e:
            Actor.log.error(f"❌ Failed to process job: {e}")
            if debug_mode:
                Actor.log.error(f"   Job data: {job}")
            continue

    Actor.log.info(f"🎉 Processing completed: {processed_count} jobs saved")
    return processed_count


async def main() -> None:
    """Main entry point for the Apify Actor."""

    async with Actor:
        # Get Actor input
        actor_input = await Actor.get_input() or {}

        # Extract configuration from environment variables
        api_key = os.getenv("API_KEY")
        api_endpoint = os.getenv("API_ENDPOINT", "https://upworkjobscraperapi.nahidhq.com")
        debug_mode = os.getenv("DEBUG_MODE", "false").lower() == "true"
        
        # Get configuration from actor input
        max_jobs = actor_input.get("maxJobs", 20)
        
        # Build filters object from individual input fields
        filters = {}
        filter_fields = [
            "paymentVerified", "category", "categoryGroup", "status", 
            "jobType", "contractorTier", "country", "tags", 
            "postedAfter", "postedBefore", "budgetMin", "budgetMax", "sort"
        ]
        
        for field in filter_fields:
            value = actor_input.get(field)
            if value is not None and value != "":  # Skip empty strings and None values
                if field == "tags" and isinstance(value, str):
                    # Convert comma-separated string to list for tags
                    filters[field] = [tag.strip() for tag in value.split(",") if tag.strip()]
                else:
                    filters[field] = value

        # Validate required inputs
        if not api_key:
            Actor.log.error(
                "❌ API key is required. Please set 'API_KEY' environment variable."
            )
            await Actor.exit()
            return

        Actor.log.info("🚀 Starting Upwork Job Scraper API Wrapper")

        if debug_mode:
            Actor.log.info("📊 Configuration:")
            Actor.log.info(f"   API Endpoint: {api_endpoint}")
            Actor.log.info(f"   Max Jobs: {max_jobs}")
            Actor.log.info(f"   Debug Mode: {debug_mode}")
            Actor.log.info(f"   Filters: {filters}")

        # Initialize API wrapper
        api_wrapper = UpworkJobAPIWrapper(api_endpoint, api_key, debug_mode)

        try:
            # Health check
            Actor.log.info("🔍 Checking API health...")
            if not await api_wrapper.health_check():
                Actor.log.error(
                    "❌ API health check failed. Please ensure the Go API is running and accessible."
                )
                await Actor.exit()
                return

            Actor.log.info("✅ API is healthy")

            # Fetch jobs from API
            Actor.log.info("📥 Fetching jobs from Go API...")
            api_response = await api_wrapper.fetch_jobs(max_jobs, filters)

            jobs = api_response.get("data", [])
            total_jobs_available = len(jobs)

            Actor.log.info(f"📊 Fetched {total_jobs_available} jobs from API")
            Actor.log.info(
                f"📊 API last updated: {api_response.get('last_updated', 'Unknown')}"
            )

            if not jobs:
                Actor.log.info("ℹ️ No jobs available from API")
                await Actor.exit()
                return

            # Process jobs simply
            processed_count = await process_jobs_simple(
                jobs, debug_mode, max_jobs
            )

            # Add usage tracking - charge for actual jobs processed using pay-per-event
            try:
                # Charge for initialization
                ret = await Actor.charge(event_name='api-result', count=processed_count)
                Actor.log.info(f"💰 [api-result] Tracked Usage: {ret}")

                # Charge for each job processed (this is handled automatically by push_data with 'api-result' event_name)
                Actor.log.info(f"💰 [api-result] Tracked Usage: {processed_count} jobs processed with 'api-result' pay-per-event charging")
            except Exception as e:
                Actor.log.warning(f"⚠️ Usage tracking failed: {e}")
                # Continue execution even if charging fails

            # Store run summary
            summary = {
                "total_jobs_available": total_jobs_available,
                "total_jobs_processed": processed_count,
                "max_jobs_limit": max_jobs,
                "api_endpoint": api_endpoint,
                "api_last_updated": api_response.get("last_updated"),
                "processed_at": datetime.now().isoformat(),
                "success": True,
            }

            # Save summary to key-value store
            await Actor.set_value("RUN_SUMMARY", summary)

            Actor.log.info(
                f"🎉 Successfully completed! Processed {processed_count} jobs out of {total_jobs_available} available."
            )

        except Exception as e:
            Actor.log.error(f"❌ Actor execution failed: {e}")

            # Store error summary
            error_summary = {
                "error": str(e),
                "error_type": type(e).__name__,
                "processed_at": datetime.now().isoformat(),
                "success": False,
            }

            await Actor.set_value("ERROR_SUMMARY", error_summary)
            raise

        finally:
            # Clean up
            await api_wrapper.close()
            Actor.log.info("🧹 Cleanup completed")


if __name__ == "__main__":
    asyncio.run(main())

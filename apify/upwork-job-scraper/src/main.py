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

    async def fetch_jobs(self, max_jobs: int) -> Dict[str, Any]:
        """Fetch jobs from the Go API.

        Args:
            max_jobs: Maximum number of jobs to fetch

        Returns:
            API response containing job data

        Raises:
            httpx.HTTPError: If API request fails
        """
        jobs_url = f"{self.api_endpoint}/jobs"
        params = {"limit": max_jobs} if max_jobs else {}

        if self.debug_mode:
            Actor.log.info(f"🔍 Fetching jobs from: {jobs_url} (limit: {max_jobs})")

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
    jobs: List[Dict[str, Any]], include_raw_data: bool, debug_mode: bool, max_jobs: int
) -> int:
    """Process jobs simply - just save them to Apify dataset.

    Args:
        jobs: List of job data from API
        include_raw_data: Whether to include raw data in output
        debug_mode: Enable debug logging

    Returns:
        Number of jobs processed
    """
    processed_count = 0
    total_jobs = len(jobs)

    Actor.log.info(f"🚀 Processing {total_jobs} jobs from API")
    
    total_jobs_to_process = min(max_jobs, total_jobs)

    for job in jobs[:total_jobs_to_process]:
        try:
            # Transform job data for Apify output
            job_data = job.get("data", {})
            output_job = {
                "job_id": job.get("job_id") or "unknown",
                "title": job_data.get("title") or "No title",
                "description": job_data.get("description") or "",
                "url": job_data.get("url") or "",
                "hourly_rate": job_data.get("hourly_rate"),
                "budget": job_data.get("budget"),
                "experience_level": job_data.get("experience_level"),
                "job_type": job_data.get("job_type"),
                "skills": job_data.get("skills") or [],
                "client_location": job_data.get("client_location"),
                "client_company_size": job_data.get("client_company_size"),
                "client_industry": job_data.get("client_industry"),
                "client_local_time": job_data.get("client_local_time"),
                "posted_date": job_data.get("posted_date"),
                "proposals_count": job_data.get("proposals_count"),
                "duration": job_data.get("duration"),
                "project_type": job_data.get("project_type"),
                "work_hours": job_data.get("work_hours"),
                "member_since": job_data.get("member_since"),
                "total_spent": job_data.get("total_spent"),
                "total_hires": job_data.get("total_hires"),
                "total_active": job_data.get("total_active"),
                "total_client_hours": job_data.get("total_client_hours"),
                "interviewing": job_data.get("interviewing"),
                "invites_sent": job_data.get("invites_sent"),
                "unanswered_invites": job_data.get("unanswered_invites"),
                "hires": job_data.get("hires"),
                "last_viewed_by_client": job_data.get("last_viewed_by_client"),
                "last_visited_at": job.get("last_visited_at"),
                "scraped_at": datetime.now().isoformat(),
            }

            # Add raw data if requested
            if include_raw_data:
                output_job["raw_data"] = job

            # Push to Apify dataset with pay-per-event charging
            await Actor.push_data(output_job, 'api-result')

            processed_count += 1

            # Log progress
            Actor.log.info(
                f"✅ Saved job {processed_count}: {output_job['title'][:50]}..."
            )

            if debug_mode:
                Actor.log.info(f"   📊 Job ID: {output_job['job_id']}")
                Actor.log.info(f"   💰 Rate: {output_job['hourly_rate']}")
                Actor.log.info(f"   📍 Location: {output_job['client_location']}")

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

        # Extract configuration
        api_key = os.getenv("API_KEY")
        api_endpoint = os.getenv("API_ENDPOINT", "http://localhost:8080")
        max_jobs = actor_input.get("maxJobs", 50)
        include_raw_data = True  # Always include raw data
        debug_mode = False  # Debug mode disabled for simplicity

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
            api_response = await api_wrapper.fetch_jobs(max_jobs)

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
                jobs, include_raw_data, debug_mode, max_jobs
            )

            # Add usage tracking - charge for actual jobs processed using pay-per-event
            try:
                # Charge for initialization
                await Actor.charge(event_name='actor-init')
                
                # Charge for each job processed (this is handled automatically by push_data with 'api-result' event_name)
                Actor.log.info(f"💰 Usage tracked: {processed_count} jobs processed with 'api-result' pay-per-event charging")
                
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

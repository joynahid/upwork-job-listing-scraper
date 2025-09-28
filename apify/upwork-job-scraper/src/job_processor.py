"""Job processing service for the Upwork Job Scraper Actor."""

from datetime import datetime
from typing import Any, Dict, List, Optional

from apify import Actor

try:
    from .api_wrapper import UpworkJobAPIWrapper
except ImportError:
    from api_wrapper import UpworkJobAPIWrapper


class JobProcessor:
    """Service class for processing Upwork jobs."""

    def __init__(self, api_wrapper: UpworkJobAPIWrapper):
        """Initialize with API wrapper."""
        self.api_wrapper = api_wrapper

    async def process_jobs_batch(
        self, max_jobs: int, filters: Dict[str, Any], debug_mode: bool
    ) -> Dict[str, Any]:
        """Process a batch of jobs and return summary."""
        try:
            # Fetch jobs from API
            Actor.log.info("📥 Fetching jobs...")
            api_response = await self.api_wrapper.fetch_jobs(max_jobs, filters)

            jobs = api_response.get("data", [])
            total_jobs_available = len(jobs)

            Actor.log.info(f"📊 Fetched {total_jobs_available} jobs")
            Actor.log.info(
                f"📊 Last updated: {api_response.get('last_updated', 'Unknown')}"
            )

            if not jobs:
                Actor.log.info("ℹ️ No jobs available from API")
                return self._create_summary(
                    0, 0, max_jobs, api_response.get("last_updated")
                )

            # Process jobs
            processed_count = await self._process_jobs_simple(
                jobs, debug_mode, max_jobs
            )

            # Create and store summary
            summary = self._create_summary(
                total_jobs_available,
                processed_count,
                max_jobs,
                api_response.get("last_updated"),
            )
            await Actor.set_value("RUN_SUMMARY", summary)

            Actor.log.info(
                f"🎉 Successfully completed! Processed {processed_count} jobs out of {total_jobs_available} available."
            )

            return summary

        except Exception as e:
            Actor.log.error(f"❌ Job processing failed: {e}")

            # Store error summary
            error_summary = {
                "error": str(e),
                "error_type": type(e).__name__,
                "processed_at": datetime.now().isoformat(),
                "success": False,
            }

            await Actor.set_value("ERROR_SUMMARY", error_summary)
            raise

    async def _process_jobs_simple(
        self, jobs: List[Dict[str, Any]], debug_mode: bool, max_jobs: int
    ) -> int:
        """Process jobs simply - save them to Apify dataset exactly as received from Go API."""
        processed_count = 0
        total_jobs = len(jobs)

        Actor.log.info(f"🚀 Processing {total_jobs} jobs from API")

        total_jobs_to_process = min(max_jobs, total_jobs)

        for job in jobs[:total_jobs_to_process]:
            try:
                # Pass through the job data exactly as received from Go API
                output_job = job.copy()  # Make a copy to avoid modifying original

                # Add scraped timestamp
                output_job["scraped_at"] = datetime.now().isoformat()

                # Push to Apify dataset with pay-per-event charging
                await Actor.push_data(output_job, "api-result")

                processed_count += 1

                # Log progress
                job_title = output_job.get("title", "Unknown Title")
                Actor.log.info(f"✅ Saved job {processed_count}: {job_title[:50]}...")

                if debug_mode:
                    self._log_job_details(output_job)

            except Exception as e:
                Actor.log.error(f"❌ Failed to process job: {e}")
                if debug_mode:
                    Actor.log.error(f"   Job data: {job}")
                continue

        Actor.log.info(f"🎉 Processing completed: {processed_count} jobs saved")
        return processed_count

    def _log_job_details(self, job: Dict[str, Any]) -> None:
        """Log detailed job information for debug mode."""
        Actor.log.info(f"   📊 Job Title: {job.get('title', 'N/A')}")
        Actor.log.info(f"   🏷️ Job Type: {job.get('job_type', 'N/A')}")
        Actor.log.info(f"   🎯 Contractor Tier: {job.get('contractor_tier', 'N/A')}")
        Actor.log.info(f"   🔒 Private: {job.get('is_private', False)}")

        # Log budget info
        budget = job.get("budget", {})
        if budget and budget.get("fixed_amount"):
            Actor.log.info(
                f"   💰 Budget: ${budget.get('fixed_amount')} {budget.get('currency', 'USD')}"
            )

        # Log hourly info
        hourly = job.get("hourly_budget", {})
        if hourly and (hourly.get("min") or hourly.get("max")):
            min_rate = hourly.get("min", 0)
            max_rate = hourly.get("max", 0)
            currency = hourly.get("currency", "USD")
            Actor.log.info(f"   💵 Hourly: ${min_rate}-${max_rate} {currency}")

    def _create_summary(
        self,
        total_available: int,
        processed: int,
        max_jobs: int,
        last_updated: Optional[str],
    ) -> Dict[str, Any]:
        """Create a summary of the processing results."""
        return {
            "total_jobs_available": total_available,
            "processed_count": processed,
            "max_jobs_limit": max_jobs,
            "api_last_updated": last_updated,
            "processed_at": datetime.now().isoformat(),
            "success": True,
        }

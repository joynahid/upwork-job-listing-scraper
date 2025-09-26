"""Upwork job service using Botasaurus driver for Cloudflare-aware scraping."""

from __future__ import annotations

import asyncio
import logging
import os
import random
from datetime import datetime
from typing import Any

from botasaurus_driver import Driver
from ..realtimedb import JobData, RealtimeJobDatabase

from ..schemas.input import ActorInput
from .extraction_scripts import JOB_LISTING_EXTRACTION_SCRIPT, JOB_DETAIL_SCRIPT

logger = logging.getLogger(__name__)


class UpworkJobService:
    """Main service for Upwork job scraping using Botasaurus."""

    def __init__(self, input_config: ActorInput, rt_db: RealtimeJobDatabase, data_store: Any = None):
        self.config: ActorInput = input_config
        self.driver: Driver | None = None
        self.comprehensive_jobs_found: list[dict] = []
        self.proxy_url: str | None = None
        self.rt_db: RealtimeJobDatabase = rt_db
        self.data_store = data_store
        self._initialized = False

    async def __aenter__(self):
        """Async context manager entry."""
        await self.initialize()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.cleanup()
        return False  # Don't suppress exceptions

    async def initialize(self) -> None:
        """Initialize Botasaurus driver and proxy configuration."""
        if self._initialized:
            return
            
        logger.info("Initializing Botasaurus driver")
        logger.info(
            "Configuration: max_jobs=%s, debug_mode=%s",
            self.config.max_jobs,
            self.config.debug_mode,
        )

        try:
            # Get proxy URL from environment
            self.proxy_url = os.getenv('PROXY_URL')
            if self.proxy_url:
                logger.info("Using HTTP proxy: %s", self.proxy_url)
            else:
                logger.info("No proxy configured - running without proxy")

            driver_kwargs: dict[str, Any] = {
                "headless": False,
                "wait_for_complete_page_load": True,
                "lang": "en-US,en",
            }

            if self.proxy_url:
                driver_kwargs["proxy"] = self.proxy_url

            logger.info("Starting Botasaurus driver with headless=%s", driver_kwargs["headless"])
            
            # Check for cancellation before creating driver
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                raise asyncio.CancelledError()
            
            self.driver = Driver(**driver_kwargs)

            if not self.config.debug_mode:
                try:
                    self.driver.enable_human_mode()
                    logger.info("Enabled Botasaurus human mode to mimic user interactions")
                except Exception as exc:  # pragma: no cover - best effort
                    logger.debug("Unable to enable human mode: %s", exc)

            search_urls_count = len(self.config.build_search_urls())
            logger.info("Ready to scrape %s search URLs", search_urls_count)
            self._initialized = True
            
        except Exception:
            # Cleanup on initialization failure
            await self.cleanup()
            raise

    async def run_scraping(self, search_urls: list[str]) -> None:
        """Run scraping workflow using Botasaurus."""
        if not self.driver:
            raise RuntimeError("Driver not initialized. Call initialize() before run_scraping().")

        for index, url in enumerate(search_urls, start=1):
            # Check if task is cancelled (proper asyncio way)
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                logger.info("Task cancelled - stopping scraping")
                raise asyncio.CancelledError()
                
            if len(self.comprehensive_jobs_found) >= self.config.max_jobs:
                logger.info("Reached max_jobs limit; stopping further processing")
                break

            logger.info("Processing search URL %s/%s: %s", index, len(search_urls), url)
            logger.info(f"Scraping search page {index}/{len(search_urls)}")

            try:
                await self._process_search_url(url)
            except asyncio.CancelledError:
                logger.info("Scraping cancelled during URL processing")
                raise  # Re-raise to properly propagate cancellation
            except Exception as exc:
                logger.error("Failed to process search URL %s: %s", url, exc, exc_info=True)

            if len(self.comprehensive_jobs_found) >= self.config.max_jobs:
                logger.info("Max jobs reached after processing search URL %s", url)
                break

            # Check for cancellation before delay
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                logger.info("Task cancelled - skipping delay and stopping")
                raise asyncio.CancelledError()
                
            await self._apply_random_delay()

        logger.info(
            "🏁 Real-time job scraping session finished with %s jobs saved individually", len(self.comprehensive_jobs_found)
        )

    async def _process_search_url(self, url: str) -> None:
        """Process a single Upwork search page with Botasaurus."""
        if not self.driver:
            raise RuntimeError("Driver not initialized")

        self.driver.get(url, bypass_cloudflare=True, wait=10, timeout=120)
        try:
            self.driver.detect_and_bypass_cloudflare()
        except Exception as exc:  # pragma: no cover - best effort
            logger.debug("Cloudflare detection bypass warning: %s", exc)

        try:
            self.driver.wait_for('[data-test="job-tile"]', timeout=15)
        except Exception:
            logger.debug("Primary selector not found; continuing with fallback extraction")

        if self.config.take_screenshots:
            screenshot_path = f"search_page_{len(self.comprehensive_jobs_found)}.png"
            try:
                self.driver.save_screenshot(screenshot_path)
                logger.info("Saved search page screenshot to %s", screenshot_path)
            except Exception as exc:
                logger.warning("Failed to save screenshot %s: %s", screenshot_path, exc)

        # Extract job URLs from search page
        job_urls = self._extract_job_urls_from_page()
        logger.info("Extracted %s job URLs from search page", len(job_urls))

        # Process each job URL to get comprehensive data
        new_job_urls = []
        seen_job_ids = set()  # Track job_ids in this batch for uniqueness checking
        
        for job_url in job_urls:
            if len(self.comprehensive_jobs_found) >= self.config.max_jobs:
                break

            # Extract job_id from URL for RocksDB tracking
            job_id = None
            try:
                # Remove query parameters first
                url_without_params = job_url.split("?")[0]
                
                # Split by "/" and get the last non-empty part
                parts = [part for part in url_without_params.split("/") if part]
                
                if parts:
                    # The job ID is typically the last part, which looks like a tilde followed by numbers
                    last_part = parts[-1]
                    
                    # If the last part starts with ~, it's likely the job ID
                    if last_part.startswith("~"):
                        job_id = last_part
                    else:
                        # Fallback to old logic
                        job_id = last_part
                        
                logger.info("🔍 Extracted job_id: '%s' from URL: %s", job_id, job_url)
            except Exception:
                logger.warning("Could not extract job_id from URL: %s", job_url)

            # Check for duplicates in this batch
            if job_id:
                if job_id in seen_job_ids:
                    logger.warning("🔄 DUPLICATE job_id in batch: %s - skipping", job_id)
                    continue
                seen_job_ids.add(job_id)
                
                should_process = self.rt_db.should_process(job_id)
                logger.info("Job %s should_process: %s", job_id[:20] + "..." if len(job_id) > 20 else job_id, should_process)
                
                if not should_process:
                    logger.info("⏭️  Skipping job %s (already processed recently)", job_id[:20] + "..." if len(job_id) > 20 else job_id)
                    continue
                else:
                    logger.info("✅ Adding job %s to processing queue", job_id[:20] + "..." if len(job_id) > 20 else job_id)
            else:
                logger.warning("⚠️  No job_id extracted, will process anyway: %s", job_url)

            new_job_urls.append(job_url)

        logger.info(
            "📊 Found %s unique job_ids, %s new URLs to process individually in real-time (total processed=%s)",
            len(seen_job_ids),
            len(new_job_urls),
            len(self.comprehensive_jobs_found),
        )

        # Process job URLs individually in real-time
        if new_job_urls:
            await self._process_job_urls_individually(
                new_job_urls[: self.config.max_jobs - len(self.comprehensive_jobs_found)]
            )

        await self._apply_random_delay()

    async def _process_job_urls_individually(self, job_urls: list[str]) -> None:
        """Process job URLs one by one in real-time to save jobs immediately."""
        total_jobs = len(job_urls)

        logger.info(f"Processing {total_jobs} job URLs individually in real-time")

        for i, job_url in enumerate(job_urls, 1):
            # Check if we've reached the max jobs limit
            remaining_slots = self.config.max_jobs - len(self.comprehensive_jobs_found)
            if remaining_slots <= 0:
                logger.info("Reached max_jobs limit, stopping individual processing")
                break

            # Check for cancellation before processing each job
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                logger.info("Task cancelled during individual job processing")
                raise asyncio.CancelledError()

            logger.info(f"Processing job {i}/{total_jobs}: {job_url}")

            try:
                # Process single job URL and save immediately
                await self._process_single_job_url(job_url)
                
                # Log progress after each job
                successful_count = len([job for job in self.comprehensive_jobs_found if 'error' not in job or job.get('error') is None])
                failed_count = len(self.comprehensive_jobs_found) - successful_count
                
                logger.info("✅ Completed job %d/%d - Total successful: %d, Total failed: %d", 
                           i, total_jobs, successful_count, failed_count)

            except asyncio.CancelledError:
                logger.info("Individual job processing cancelled")
                raise
            except Exception as exc:
                logger.error("Failed to process individual job %s: %s", job_url, exc, exc_info=True)

            # Add small delay between individual jobs to be respectful
            if i < total_jobs:  # Don't delay after the last job
                await asyncio.sleep(1)

    async def _process_single_job_url(self, job_url: str) -> None:
        """Process a single job URL and save it immediately."""
        if not self.driver:
            raise RuntimeError("Driver not initialized")

        try:
            logger.debug(f"Opening job URL: {job_url}")
            
            # Open the job URL in a new tab
            tab = self.driver.open_link_in_new_tab(
                job_url, bypass_cloudflare=True, wait=3, timeout=120
            )
            
            # Process the job and save immediately
            await self._extract_and_push_comprehensive_job_from_tab(tab, job_url)
            
        except Exception as exc:
            logger.error(f"Failed to process single job URL {job_url}: {exc}", exc_info=True)
            
            # Record the failed extraction
            from ..realtimedb import FailedJobExtraction
            failed_job = FailedJobExtraction(
                url=job_url,
                error_message=f"Single job processing error: {exc}",
                raw_data=None
            )
            self.rt_db.record_failed_extraction(failed_job)
            
            # Still store a minimal fallback entry for compatibility
            fallback = {
                "type": "job",
                "url": job_url,
                "error": str(exc),
                "processed_at": datetime.now().isoformat(),
                "extraction_status": "failed"
            }
            if self.data_store:
                await self.data_store.push_data(fallback)

    async def _extract_and_push_comprehensive_job_from_tab(self, tab: Any, job_url: str) -> None:
        """Extract comprehensive job information from a pre-opened tab and save immediately."""
        if not self.driver:
            return

        try:
            # Switch to the specific tab
            logger.info(f"🔄 Processing job in real-time: {job_url}")
            self.driver.switch_to_tab(tab)

            try:
                self.driver.detect_and_bypass_cloudflare()
            except Exception as exc:  # pragma: no cover - best effort
                logger.debug("Cloudflare bypass warning on job detail: %s", exc)

            try:
                self.driver.wait_for('h1, [data-test="job-title"]', timeout=15)
                # Additional wait for dynamic content to load
                import time
                time.sleep(2)  # Give the page more time to load completely
            except Exception:
                logger.debug("Job title element not found immediately on detail page")

            # Extract comprehensive job details
            logger.debug(f"🔍 Extracting job details from: {job_url}")
            detailed_info = self._extract_job_details()

            if detailed_info and isinstance(detailed_info, dict):
                
                try:
                    # Create clean job data structure
                    clean_job_data = JobData(**detailed_info)
                    
                    clean_job_data.set_url(job_url)
                    logger.info("🔗 JobData object created with job_id: '%s'", clean_job_data.job_id)

                    # Mark job as seen in RocksDB immediately
                    self.rt_db.do_seen(clean_job_data)
                    logger.info("✅ Marked job %s as SEEN in database", clean_job_data.job_id[:20] + "..." if len(clean_job_data.job_id or "") > 20 else clean_job_data.job_id)

                    # Store for tracking
                    self.comprehensive_jobs_found.append(clean_job_data.model_dump())

                    # Push clean data immediately to data store
                    if self.data_store:
                        await self.data_store.push_data(clean_job_data.model_dump())
                        logger.info("💾 Job saved to data store in real-time: %s", clean_job_data.title or clean_job_data.job_id or job_url)
                    else:
                        logger.warning("No data store configured - job not persisted")
                    
                    logger.info("🎉 Successfully processed and saved job: %s", clean_job_data.title or clean_job_data.job_id or job_url)
                    
                except Exception as validation_exc:
                    # Handle validation errors gracefully - data doesn't meet JobData requirements
                    logger.warning("❌ Job data validation failed for %s: %s", job_url, validation_exc)
                    
                    # Record the failed extraction with detailed information
                    from ..realtimedb import FailedJobExtraction
                    failed_job = FailedJobExtraction(
                        url=job_url,
                        error_message=f"Data validation failed: {validation_exc}",
                        raw_data=detailed_info
                    )
                    self.rt_db.record_failed_extraction(failed_job)
                    
                    # Continue processing other jobs instead of crashing
                    return
            else:
                # No data extracted at all
                logger.warning("⚠️ No data extracted for job URL: %s", job_url)
                from ..realtimedb import FailedJobExtraction
                failed_job = FailedJobExtraction(
                    url=job_url,
                    error_message="No data could be extracted from job page",
                    raw_data=detailed_info
                )
                self.rt_db.record_failed_extraction(failed_job)

        except Exception as exc:
            logger.error("💥 Failed to extract job for %s: %s", job_url, exc, exc_info=True)
            
            # Record the failed extraction
            from ..realtimedb import FailedJobExtraction
            failed_job = FailedJobExtraction(
                url=job_url,
                error_message=f"Extraction error: {exc}",
                raw_data=None
            )
            self.rt_db.record_failed_extraction(failed_job)
            
            # Still store a minimal fallback entry for compatibility
            fallback = {
                "type": "job",
                "url": job_url,
                "error": str(exc),
                "processed_at": datetime.now().isoformat(),
                "extraction_status": "failed"
            }
            if self.data_store:
                await self.data_store.push_data(fallback)
        finally:
            try:
                logger.debug(f"🔒 Closing tab for job URL: {job_url}")
                tab.close()
            except Exception as close_exc:  # pragma: no cover - best effort
                logger.debug("Failed to close detail tab: %s", close_exc)

    def _extract_job_urls_from_page(self) -> list[str]:
        """Extract job URLs from the current search page using JavaScript."""
        if not self.driver:
            return []

        try:
            job_data = self.driver.run_js(JOB_LISTING_EXTRACTION_SCRIPT)
            logger.info("JavaScript extraction returned %s job URL objects", len(job_data or []))
        except Exception as exc:
            logger.error("Error running job URL extraction script: %s", exc, exc_info=True)
            return []

        job_urls: list[str] = []
        invalid_count = 0

        for i, data in enumerate(job_data or [], 1):
            try:
                url = data.get("url", "")
                if not url or not url.startswith("http"):
                    invalid_count += 1
                    continue

                job_urls.append(url)
            except Exception as exc:
                logger.warning("Failed to parse job URL object %s: %s", i, exc)
                invalid_count += 1

        logger.info(
            "Job URL parsing completed - %s valid URLs, %s invalid entries",
            len(job_urls),
            invalid_count,
        )
        return job_urls

    def _extract_job_details(self) -> dict[str, Any] | None:
        """Extract detailed job information from the current page."""
        if not self.driver:
            return None

        try:
            detail_data = self.driver.run_js(JOB_DETAIL_SCRIPT)
            return detail_data
        except Exception as exc:
            logger.error("Error extracting job details: %s", exc)
            return None

    async def _apply_random_delay(self) -> None:
        """Apply a random delay between configured min and max."""
        delay = random.uniform(self.config.delay_min, self.config.delay_max)
        logger.debug("Sleeping for %.2f seconds to mimic human behaviour", delay)
        
        # Use asyncio.sleep which respects cancellation
        try:
            await asyncio.sleep(delay)
        except asyncio.CancelledError:
            logger.debug("Delay interrupted by cancellation")
            raise

    async def cleanup(self) -> None:
        """Clean up resources with proper async handling."""
        if not self._initialized:
            return
            
        logger.info("Starting service cleanup...")

        # Step 1: Close browser tabs and driver
        if self.driver:
            try:
                logger.info("Attempting to close Botasaurus driver...")

                # Try to close all tabs first
                try:
                    await asyncio.wait_for(
                        asyncio.get_event_loop().run_in_executor(
                            None, self.driver.close_all_tabs
                        ), 
                        timeout=5.0
                    )
                    logger.debug("All tabs closed successfully")
                except asyncio.TimeoutError:
                    logger.warning("Tab closure timed out")
                except Exception as exc:
                    logger.debug("Failed to close all tabs: %s", exc)

                # Close the driver
                try:
                    await asyncio.wait_for(
                        asyncio.get_event_loop().run_in_executor(
                            None, self.driver.close
                        ), 
                        timeout=10.0
                    )
                    logger.info("Driver closed successfully")
                except asyncio.TimeoutError:
                    logger.warning("Driver close timed out, attempting quit")
                    try:
                        self.driver.quit()
                        logger.info("Driver quit completed")
                    except Exception as quit_exc:
                        logger.warning("Driver quit failed: %s", quit_exc)
                except Exception as exc:
                    logger.warning("Driver close failed: %s", exc)
                    try:
                        self.driver.quit()
                        logger.info("Fallback quit completed")
                    except Exception as quit_exc:
                        logger.warning("Fallback quit failed: %s", quit_exc)

            except Exception as driver_exc:
                logger.error("Driver cleanup error: %s", driver_exc)
            finally:
                self.driver = None

        # Step 2: Clean up any hanging processes
        await self._cleanup_hanging_processes()
        
        self._initialized = False
        logger.info("Service cleanup completed")

    async def _cleanup_hanging_processes(self) -> None:
        """Clean up any hanging Chrome processes with safety checks."""
        try:
            import os
            import subprocess
            import platform

            # Only attempt process cleanup on Unix-like systems
            if platform.system() not in ['Linux', 'Darwin']:
                logger.debug("Process cleanup skipped on %s", platform.system())
                return

            # Get current process ID to avoid killing ourselves
            current_pid = os.getpid()
            
            # Find Chrome processes with our pattern
            try:
                result = await asyncio.wait_for(
                    asyncio.get_event_loop().run_in_executor(
                        None, 
                        subprocess.run, 
                        ["pgrep", "-f", "bota.*chrome"], 
                        {"capture_output": True, "text": True}
                    ),
                    timeout=5.0
                )
                
                if result.returncode == 0 and result.stdout:
                    pids = result.stdout.strip().split("\n")
                    killed_count = 0
                    
                    for pid_str in pids:
                        try:
                            pid = int(pid_str)
                            if pid != current_pid:  # Don't kill ourselves
                                os.kill(pid, 9)  # SIGKILL
                                killed_count += 1
                                logger.debug(f"Killed hanging Chrome process: {pid}")
                        except (ValueError, ProcessLookupError):
                            # Process already gone or invalid PID
                            pass
                        except PermissionError as e:
                            logger.debug(f"Permission denied killing process {pid_str}: {e}")
                    
                    if killed_count > 0:
                        logger.info(f"Cleaned up {killed_count} hanging Chrome processes")
                        
            except subprocess.CalledProcessError:
                # pgrep found no processes, which is fine
                logger.debug("No hanging Chrome processes found")
            except asyncio.TimeoutError:
                logger.warning("Process cleanup timed out")

        except Exception as cleanup_exc:
            logger.debug("Process cleanup failed: %s", cleanup_exc)

        logger.info("Service cleanup completed")

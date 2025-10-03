"""Upwork job service using Botasaurus driver for Cloudflare-aware scraping."""

from __future__ import annotations

import asyncio
import json
import logging
import os
import random
import threading
from datetime import datetime, timezone
from pathlib import Path
import subprocess
import tempfile
import time
from typing import Any
from botasaurus_driver.driver import Tab
from botasaurus_driver.exceptions import CloudflareDetectionException
from botasaurus_driver import Driver
from bs4 import BeautifulSoup
from google.cloud import firestore

from src.firebase_provider import get_firebase_with_config
from ..schemas.input import ActorInput

logger = logging.getLogger(__name__)


class UpworkJobService:
    """Main service for Upwork job scraping using Botasaurus."""

    def __init__(
        self,
        input_config: ActorInput,
        data_store: Any = None,
    ):
        self.config: ActorInput = input_config
        self.driver: Driver | None = None
        self._total_jobs_processed: int = 0
        self.proxy_url: str | None = None
        self.data_store = data_store
        self._initialized = False
        self._driver_thread_lock = threading.RLock()
        browser_profile_env = os.getenv("BROWSER_PROFILE_PATH")
        default_profile = Path("browser_data") / "upwork_scraper_profile"
        self.browser_profile_path = (
            Path(browser_profile_env).expanduser()
            if browser_profile_env
            else default_profile
        ).resolve()
        self.browser_profile_path.mkdir(parents=True, exist_ok=True)
        self.firebase = get_firebase_with_config(
            service_account_path=os.environ["FIREBASE_SERVICE_ACCOUNT_PATH"],
        )

    async def __aenter__(self):
        """Async context manager entry."""
        await self.initialize()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.cleanup()
        return False  # Don't suppress exceptions

    @property
    def job_list_db(self) -> firestore.AsyncCollectionReference:
        return self.firebase.firestore.collection("job_list")

    @property
    def individual_job_db(self) -> firestore.AsyncCollectionReference:
        return self.firebase.firestore.collection("individual_jobs")

    @property
    def total_jobs_processed(self) -> int:
        """Total count of jobs processed in the current run."""
        return self._total_jobs_processed

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
            self.proxy_url = os.getenv("PROXY_URL")
            if self.proxy_url:
                logger.info("Using HTTP proxy: %s", self.proxy_url)
            else:
                logger.info("No proxy configured - running without proxy")

            driver_kwargs: dict[str, Any] = {
                "headless": False,
                "wait_for_complete_page_load": True,  # Don't wait for JS/dynamic content
                "lang": "en-US,en",
            }

            if self.proxy_url:
                driver_kwargs["proxy"] = self.proxy_url

            driver_kwargs["profile"] = str(self.browser_profile_path)

            logger.info(
                "Starting Botasaurus driver with headless=%s", driver_kwargs["headless"]
            )
            logger.info(
                "Using persistent browser profile at %s", self.browser_profile_path
            )

            # Check for cancellation before creating driver
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                raise asyncio.CancelledError()

            self.driver = Driver(**driver_kwargs)

            if not self.config.debug_mode:
                try:
                    self.driver.enable_human_mode()
                    logger.info(
                        "Enabled Botasaurus human mode to mimic user interactions"
                    )
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
            raise RuntimeError(
                "Driver not initialized. Call initialize() before run_scraping()."
            )

        for index, url in enumerate(search_urls, start=1):
            # Check if task is cancelled (proper asyncio way)
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                logger.info("Task cancelled - stopping scraping")
                raise asyncio.CancelledError()

            logger.info("Processing search URL %s/%s: %s", index, len(search_urls), url)
            logger.info(f"Scraping search page {index}/{len(search_urls)}")

            try:
                await self._process_search_url(url)
            except asyncio.CancelledError:
                logger.info("Scraping cancelled during URL processing")
                raise  # Re-raise to properly propagate cancellation
            except CloudflareDetectionException:
                logger.warning("Cloudflare detection exception - skipping URL %s", url)
                raise
            except Exception as exc:
                logger.error(
                    "Failed to process search URL %s: %s", url, exc, exc_info=True
                )

            # Check for cancellation before delay
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                logger.info("Task cancelled - skipping delay and stopping")
                raise asyncio.CancelledError()

            await self._apply_random_delay()

        logger.info(
            "🏁 Scraping session complete! Total: %s jobs saved to Firestore",
            self.total_jobs_processed,
        )

    async def handle_cloudflare_detection(
        self, retry_attempts: int = 3, next_delay: float = 1.0
    ) -> None:
        """Handle Cloudflare detection."""
        time.sleep(next_delay)
        try:
            self.driver.detect_and_bypass_cloudflare()
        except Exception as exc:  # pragma: no cover - best effort
            logger.debug("Cloudflare detection bypass warning: %s", exc)
            if retry_attempts > 0:
                logger.debug("Retrying Cloudflare detection bypass...")
                await self.handle_cloudflare_detection(
                    retry_attempts - 1, next_delay * 2
                )
            else:
                logger.error(
                    "Failed to bypass Cloudflare detection after %s attempts",
                    retry_attempts,
                )
                raise exc

    def gen_job_url(self, job: dict) -> str:
        """Generate the job URL."""
        if "ciphertext" not in job:
            raise ValueError("Job does not have a ciphertext")

        ciphertext = job["ciphertext"]
        return f"https://www.upwork.com/jobs/{ciphertext}?referrer_url_path=%2Fnx%2Fsearch%2Fjobs%2Fdetails%2F{ciphertext}"

    async def save_job_listing_details(self, job: dict) -> None:
        """Save the job details."""
        job_uid = job.get("uid")
        if not job_uid:
            logger.error(
                "Skipping job listing save; missing uid. Payload=%s",
                self._serialize_for_logging(job),
            )
            return

        logger.info(
            "Saving job listing to Firestore: uid=%s payload=%s",
            job_uid,
            self._serialize_for_logging(job),
        )
        await self.job_list_db.document(job_uid).set(job, merge=True)

    async def save_individual_job_details(self, job: dict) -> None:
        """Save the individual job details."""
        job_uid = job.get("uid")
        if not job_uid:
            logger.error(
                "Skipping individual job save; missing uid. Payload=%s",
                self._serialize_for_logging(job),
            )
            return

        logger.info(
            "Saving individual job to Firestore: uid=%s payload=%s",
            job_uid,
            self._serialize_for_logging(job),
        )
        await self.individual_job_db.document(job_uid).set(job, merge=True)

    @staticmethod
    def _serialize_for_logging(payload: Any) -> str:
        """Return a compact, truncated JSON string for logging Firestore payloads."""
        try:
            serialized = json.dumps(payload, default=str)
        except (TypeError, ValueError):
            serialized = repr(payload)

        max_length = 1500
        if len(serialized) > max_length:
            serialized = f"{serialized[:max_length]}... (truncated)"

        return serialized

    async def _process_search_url(self, url: str) -> None:
        """Process a single Upwork search page with Botasaurus."""
        if not self.driver:
            raise RuntimeError("Driver not initialized")

        self.driver.get(url, bypass_cloudflare=True, wait=10, timeout=120)
        try:
            await self.handle_cloudflare_detection()
        except Exception as exc:
            logger.debug("Cloudflare detection bypass warning: %s", exc)

        try:
            self.driver.wait_for("body > script:nth-child(10)", timeout=15)
        except Exception:
            logger.debug(
                "Primary selector not found; continuing with fallback extraction"
            )

        # Extract job URLs from search page
        job_list = self._extract_job_urls_from_page()
        logger.info("Extracted %s job URLs from search page", len(job_list))

        for job in job_list:
            await self.save_job_listing_details(job)

        # Process each job URL to get comprehensive data with immediate tab cleanup per job
        await self._process_job_urls_individually(job_list)

        await self._apply_random_delay()

    async def _process_job_urls_individually(self, job_list: list[dict]) -> None:
        """Process job URLs concurrently by opening all tabs simultaneously."""
        total_jobs = len(job_list)

        if total_jobs == 0:
            logger.info("No jobs to process individually in real-time")
            return

        logger.info(f"Processing {total_jobs} job URLs concurrently with no limits")

        # Temporary fix for the bug in Botasaurus Driver - dynamically add is_closed property
        if not hasattr(self.driver._tab.__class__, "is_closed"):
            self.driver._tab.__class__.is_closed = property(
                fget=lambda self: self.closed,
                fset=lambda self, value: setattr(self, "_is_closed_override", value),
            )

        # Step 1: Open ALL tabs first (lightweight, no page loading yet)
        logger.info(f"📂 Opening all {total_jobs} tabs (lightweight mode)...")
        open_tasks = [
            asyncio.create_task(self._open_tab_for_job(job, index, total_jobs))
            for index, job in enumerate(job_list, 1)
        ]

        try:
            # Wait for ALL tabs to finish opening before proceeding
            results = await asyncio.gather(*open_tasks, return_exceptions=True)
            
            # Filter out None results and exceptions
            tabs: list[tuple[Tab, dict]] = []
            for result in results:
                if isinstance(result, Exception):
                    logger.error(f"Tab opening failed: {result}")
                elif result is not None:
                    tabs.append(result)
            
            logger.info(f"✅ All {len(tabs)} tabs opened successfully!")
            
        except asyncio.CancelledError:
            logger.info("Task cancelled during tab opening; cleaning up")
            for task in open_tasks:
                task.cancel()
            raise

        # Step 2: Now visit and process each opened tab (heavy operations happen here)
        logger.info(f"🔄 Now loading and processing {len(tabs)} tabs one by one...")
        jobs_before = self._total_jobs_processed
        
        for index, (tab, job) in enumerate(tabs, 1):
            job_title = job.get("title", "Unknown")
            logger.info(f"Processing tab {index}/{len(tabs)}: {job_title}")
            
            try:
                await self._extract_and_push_comprehensive_job_from_tab(tab, job)
            except asyncio.CancelledError:
                logger.info("Processing cancelled")
                raise
            except Exception as exc:
                logger.error(
                    "Failed to process job %s: %s",
                    job_title,
                    exc,
                    exc_info=True,
                )
            finally:
                await self._close_tab(tab)
        
        # Summary
        jobs_saved = self._total_jobs_processed - jobs_before
        logger.info(f"✅ Batch complete: {jobs_saved}/{len(tabs)} jobs saved to Firestore")

    async def _open_tab_for_job(
        self, job: dict, index: int, total_jobs: int
    ) -> tuple[Tab, dict] | None:
        """Open a job detail tab immediately and return the tab with job data."""
        if not self.driver:
            raise RuntimeError("Driver not initialized")

        current_task = asyncio.current_task()
        if current_task and current_task.cancelled():
            raise asyncio.CancelledError()

        job_uid = job.get("uid", "unknown")
        logger.info(
            "Opening tab %s/%s for job %s",
            index,
            total_jobs,
            job_uid,
        )

        try:
            tab = await asyncio.to_thread(
                self._open_tab_blocking,
                job,
            )
            return tab, job
        except asyncio.CancelledError:
            logger.info("Tab opening cancelled for job %s", job_uid)
            raise
        except Exception as exc:
            logger.error(
                "Failed to open tab for job %s: %s",
                job_uid,
                exc,
                exc_info=True,
            )
            return None

    def _open_tab_blocking(self, job: dict) -> Tab:
        """Blocking helper that opens a tab with thread-safe driver access."""
        if not self.driver:
            raise RuntimeError("Driver not initialized")

        url = self.gen_job_url(job)
        with self._driver_thread_lock:
            # Open tab without waiting - let it load in background
            tab = self.driver.open_link_in_new_tab(
                url,
                bypass_cloudflare=False,  # Don't bypass yet, do it during processing
                wait=0,  # Don't wait for load, just open the tab
                timeout=120,
            )
            # Small delay to prevent browser overload
            time.sleep(0.3)
            return tab

    async def _close_tab(self, tab: Tab) -> None:
        """Close a tab safely, ignoring errors and avoiding double closes."""
        if getattr(tab, "is_closed", False):
            return

        if not self.driver:
            return

        close_fn = getattr(self.driver, "close_tab", None)

        try:
            await asyncio.to_thread(self._close_tab_blocking, tab, close_fn)
        except Exception as exc:
            logger.debug("Failed to close tab cleanly: %s", exc)

    def _close_tab_blocking(self, tab: Tab, close_fn) -> None:
        """Blocking helper to close tabs under driver lock."""
        with self._driver_thread_lock:
            if close_fn:
                close_fn(tab)
            else:
                tab.close()

    def _extract_job_data_blocking(self, tab: Tab, job_title: str) -> tuple[dict | None, str | None]:
        """Blocking helper to extract job data with thread-safe driver access."""
        if not self.driver:
            return None, None

        with self._driver_thread_lock:
            # Switch to the specific tab
            self.driver.switch_to_tab(tab)

            try:
                self.driver.detect_and_bypass_cloudflare()
            except Exception as exc:
                logger.debug("Cloudflare bypass warning on job detail: %s", exc)

            try:
                self.driver.wait_for('h1, [data-test="job-title"]', timeout=15)
            except Exception:
                logger.debug("Job title element not found immediately on detail page")

            # Extract comprehensive job details
            logger.debug(f"🔍 Extracting job details from: {job_title}")
            detailed_job = self.extract_nuxt_with_js_engine(self.driver.page_html)
            current_url = self.driver.current_url

            return detailed_job, current_url

    async def _extract_and_push_comprehensive_job_from_tab(
        self, tab: Tab, job: dict
    ) -> None:
        """Extract comprehensive job information from a pre-opened tab and save immediately."""
        if not self.driver:
            logger.error("Driver not available for job extraction")
            return

        job_title = job.get("title", "Unknown Job")
        job_uid = job.get("uid", "unknown")

        detailed_job: dict | None = None
        current_url: str | None = None

        try:
            # Extract data from tab with thread-safe driver access
            logger.info(f"🔄 Processing job in real-time: {job_title}")
            detailed_job, current_url = await asyncio.to_thread(
                self._extract_job_data_blocking, tab, job_title
            )

            if detailed_job is None:
                logger.error("💥 Failed to extract job details from: %s", job_title)
                return

            final_job_uid = job_uid or detailed_job.get("uid")
            if not final_job_uid:
                logger.error(
                    "💥 Missing job UID for %s; skipping save",
                    job_title,
                )
                return

            detailed_job["uid"] = final_job_uid

        except asyncio.CancelledError:
            logger.info("Job detail processing cancelled for %s", job_uid)
            raise
        except Exception as exc:
            logger.error(
                "💥 Failed to extract job for %s: %s", job_title, exc, exc_info=True
            )
            return
        else:
            if detailed_job is None:
                return

            detailed_job["url"] = current_url or detailed_job.get("url")
            detailed_job["scrape_metadata"] = {
                "last_visited_at": datetime.now(timezone.utc).isoformat(),
                "last_visited_by": "upwork_scraper",
            }
            self._flatten_sortable_fields(detailed_job)

            # Save job details to Firestore
            await self.save_individual_job_details(detailed_job)
            
            # Track processed job
            self._total_jobs_processed += 1
            logger.info(f"💾 Saved to Firestore: {job_title} (Job #{self._total_jobs_processed})")

    @staticmethod
    def _flatten_sortable_fields(job_data: dict) -> None:
        """
        Flatten key fields from nested structures to root level for Firestore ordering.
        Modifies job_data in place.
        """
        try:
            # Extract nested job details
            state = job_data.get("state", {})
            job_details = state.get("jobDetails", {})
            job_obj = job_details.get("job", {})

            # Also try alternative path
            if not job_obj:
                job_state = state.get("job", {})
                job_obj = job_state.get("job", {})

            # Flatten publishTime (most important for sorting)
            if "publishTime" in job_obj and job_obj["publishTime"]:
                job_data["publishTime"] = job_obj["publishTime"]

            # Flatten postedOn as fallback
            if "postedOn" in job_obj and job_obj["postedOn"]:
                job_data["postedOn"] = job_obj["postedOn"]

            # Flatten createdOn
            if "createdOn" in job_obj and job_obj["createdOn"]:
                job_data["createdOn"] = job_obj["createdOn"]

            # Flatten budget for potential budget sorting
            if "budget" in job_obj and job_obj["budget"]:
                budget = job_obj["budget"]
                if isinstance(budget, dict) and "amount" in budget:
                    job_data["budgetAmount"] = budget["amount"]

            # Flatten amount (alternative budget field)
            if "amount" in job_obj and job_obj["amount"]:
                amount = job_obj["amount"]
                if isinstance(amount, dict) and "amount" in amount:
                    job_data["fixedAmount"] = amount["amount"]

            # Flatten hourly budget for sorting
            if "hourlyBudgetMax" in job_obj:
                job_data["hourlyBudgetMax"] = job_obj["hourlyBudgetMax"]
            if "hourlyBudgetMin" in job_obj:
                job_data["hourlyBudgetMin"] = job_obj["hourlyBudgetMin"]

        except Exception as exc:
            logger.debug(f"Failed to flatten sortable fields: {exc}")

    @staticmethod
    def extract_nuxt_with_js_engine(html: str) -> dict | None:
        """
        Extract NUXT data using Node.js as a lightweight JS engine
        """

        soup = BeautifulSoup(html, "html.parser")

        # Find the script tag containing window.__NUXT__
        script_tags = soup.find_all("script")
        for script in script_tags:
            if script.string and "window.__NUXT__" in script.string:
                script_content = script.string.strip()

                # Create a temporary JS file to execute the script
                js_code = f"""
                // Create window object
                var window = {{}};
                
                // Execute the original script
                {script_content}
                
                // Output the result as JSON
                console.log(JSON.stringify(window.__NUXT__));
                """

                # Write to temporary file and execute with Node.js
                with tempfile.NamedTemporaryFile(
                    mode="w",
                    suffix=".js",
                    delete=False,
                    encoding="utf-8",
                ) as temp_file:
                    temp_file.write(js_code)
                    temp_file_path = temp_file.name

                try:
                    # Execute with Node.js
                    result = subprocess.run(
                        ["node", temp_file_path],
                        capture_output=True,
                        text=True,
                        encoding="utf-8",
                        check=True,
                    )

                    # Parse the JSON output
                    nuxt_data = json.loads(result.stdout.strip())
                    return nuxt_data

                except subprocess.CalledProcessError as e:
                    print(f"Node.js execution error: {e}")
                    print(f"stderr: {e.stderr}")
                    return None
                except json.JSONDecodeError as e:
                    print(f"JSON decode error: {e}")
                    print(f"stdout: {result.stdout[:200]}...")
                    return None
                finally:
                    # Clean up temp file
                    os.unlink(temp_file_path)

        print("No window.__NUXT__ section found")
        return None

    def _extract_job_urls_from_page(self) -> list[dict]:
        """Extract job URLs from the current search page using JavaScript."""
        if not self.driver:
            return []

        try:
            nuxt_state = self.extract_nuxt_with_js_engine(self.driver.page_html)
            job_list = nuxt_state.get("state", {}).get("jobsSearch", {}).get("jobs", [])
        except Exception as exc:
            logger.error(
                "Error running job URL extraction script: %s", exc, exc_info=True
            )
            return []

        # add metadata to the job list
        for job in job_list:
            job["scrape_metadata"] = {
                "last_visited_at": datetime.now(timezone.utc).isoformat(),
                "last_visited_by": "upwork_scraper",
            }

        return job_list

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

                for tab in self.driver._browser.tabs:
                    print(f"Closing tab: {tab}")
                    tab.close()

                self.driver.close()
            except Exception as driver_exc:
                logger.error("Driver cleanup error: %s", driver_exc)
            finally:
                self.driver = None

        self._initialized = False
        logger.info("Service cleanup completed")

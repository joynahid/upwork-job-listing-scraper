"""Upwork job service using Botasaurus driver for Cloudflare-aware scraping."""

from __future__ import annotations

import asyncio
import json
import logging
import os
import random
from datetime import datetime, timedelta
import subprocess
import tempfile
import time
from typing import Any
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
        self.comprehensive_jobs_found: list[dict] = []
        self.proxy_url: str | None = None
        self.data_store = data_store
        self._initialized = False
        self.firebase = get_firebase_with_config(
            service_account_path=os.environ["FIREBASE_SERVICE_ACCOUNT_PATH"],
        )
        self.staleness_threshold_seconds = os.getenv("STALENESS_THRESHOLD", 60)

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
                "headless": True,
                "wait_for_complete_page_load": True,  # Don't wait for JS/dynamic content
                "lang": "en-US,en",
            }

            if self.proxy_url:
                driver_kwargs["proxy"] = self.proxy_url

            logger.info(
                "Starting Botasaurus driver with headless=%s", driver_kwargs["headless"]
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
            except CloudflareDetectionException:
                logger.warning("Cloudflare detection exception - skipping URL %s", url)
                raise
            except Exception as exc:
                logger.error(
                    "Failed to process search URL %s: %s", url, exc, exc_info=True
                )

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
            "🏁 Real-time job scraping session finished with %s jobs saved individually",
            len(self.comprehensive_jobs_found),
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

    async def should_find_job_details(self, job: dict) -> bool:
        """Check if we should find the job details."""
        doc = await self.individual_job_db.document(job["uid"]).get()

        now = datetime.now()
        now.replace(tzinfo=None)

        scrape_metadata = doc.get("scrape_metadata")

        if scrape_metadata is None:
            return True

        last_visited_at = scrape_metadata.get("last_visited_at")

        if isinstance(last_visited_at, str):
            last_visited_at = datetime.fromisoformat(last_visited_at)
            last_visited_at.replace(tzinfo=None)
        else:
            last_visited_at = datetime.now() - timedelta(seconds=10000)
            last_visited_at.replace(tzinfo=None)

        seconds_ago = now - last_visited_at

        should_process = not doc.exists or seconds_ago > timedelta(
            seconds=self.staleness_threshold_seconds
        )

        logger.info(
            f"Job {job['title']} should process: {should_process}, seconds ago: {seconds_ago}"
        )

        return should_process

    def gen_job_url(self, job: dict) -> str:
        """Generate the job URL."""
        if "ciphertext" not in job:
            raise ValueError("Job does not have a ciphertext")

        ciphertext = job["ciphertext"]
        return f"https://www.upwork.com/jobs/{ciphertext}?referrer_url_path=%2Fnx%2Fsearch%2Fjobs%2Fdetails%2F{ciphertext}"

    async def save_job_listing_details(self, job: dict) -> None:
        """Save the job details."""
        await self.job_list_db.document(job["uid"]).set(job, merge=True)

    async def save_individual_job_details(self, job: dict) -> None:
        """Save the individual job details."""
        await self.individual_job_db.document(job["uid"]).set(job, merge=True)

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

        # Find by job_uid
        for job in job_list:
            should_process = await self.should_find_job_details(job)
            await self.save_job_listing_details(job)

            if not should_process:
                logger.info(
                    "Skipping job %s because it should not be processed", job["uid"]
                )
                continue

        # Process each job URL to get comprehensive data
        new_job_urls = []

        for job in job_list:
            job_id = job["uid"]

            if not await self.should_find_job_details(job):
                logger.warning(
                    "🔄 Job %s already processed recently - skipping", job_id
                )
                continue

            logger.info(
                "✅ Adding job %s to processing queue",
                job_id[:20] + "..." if len(job_id) > 20 else job_id,
            )

            new_job_urls.append(job)
        await self._process_job_urls_individually(
            new_job_urls[: self.config.max_jobs - len(self.comprehensive_jobs_found)]
        )

        await self._apply_random_delay()

    async def _process_job_urls_individually(self, job_list: list[dict]) -> None:
        """Process job URLs one by one in real-time to save jobs immediately."""
        total_jobs = len(job_list)

        if total_jobs == 0:
            logger.info("No jobs to process individually in real-time")
            return

        logger.info(f"Processing {total_jobs} job URLs individually in real-time")

        for i, job in enumerate(job_list, 1):
            # Check for cancellation before processing each job
            current_task = asyncio.current_task()
            if current_task and current_task.cancelled():
                logger.info("Task cancelled during individual job processing")
                raise asyncio.CancelledError()

            logger.info(f"Processing job {i}/{total_jobs}: {job['uid']}")

            try:
                await self._process_single_job_url(job)
            except asyncio.CancelledError:
                logger.info("Individual job processing cancelled")
                raise
            except Exception as exc:
                logger.error(
                    "Failed to process individual job %s: %s",
                    job["uid"],
                    exc,
                    exc_info=True,
                )

    async def _process_single_job_url(self, job: dict) -> None:
        """Process a single job URL and save it immediately."""
        if not self.driver:
            raise RuntimeError("Driver not initialized")

        try:
            logger.debug(f"Opening job URL: {job['uid']}")

            # Open the job URL in a new tab
            tab = self.driver.open_link_in_new_tab(
                self.gen_job_url(job), bypass_cloudflare=True, wait=3, timeout=120
            )

            # Process the job and save immediately
            await self._extract_and_push_comprehensive_job_from_tab(tab, job)

        except Exception as exc:
            logger.error(
                f"Failed to process single job URL {job['title']}: {exc}", exc_info=True
            )
            if self.data_store:
                await self.data_store.push_data(job)

    async def _extract_and_push_comprehensive_job_from_tab(
        self, tab: Any, job: dict
    ) -> None:
        """Extract comprehensive job information from a pre-opened tab and save immediately."""
        if not self.driver:
            return

        try:
            # Switch to the specific tab
            logger.info(f"🔄 Processing job in real-time: {job['title']}")
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
            logger.debug(f"🔍 Extracting job details from: {job['title']}")
            detailed_job = self.extract_nuxt_with_js_engine(self.driver.page_html)

            if detailed_job is None:
                logger.error("💥 Failed to extract job details from: %s", job["title"])
                return

            job_uid = job.get("uid") or detailed_job.get("uid")
            if not job_uid:
                logger.error(
                    "💥 Missing job UID for %s; skipping save",
                    job.get("title", "<unknown>"),
                )
                return

            detailed_job["uid"] = job_uid
            detailed_job["url"] = self.driver.current_url
            detailed_job["scrape_metadata"] = {
                "last_visited_at": datetime.now().isoformat(),
                "last_visited_by": "upwork_scraper",
            }
            await self.save_individual_job_details(detailed_job)
        except Exception as exc:
            logger.error(
                "💥 Failed to extract job for %s: %s", job["title"], exc, exc_info=True
            )
        finally:
            try:
                logger.debug(f"🔒 Closing tab for job URL: {job['title']}")
                tab.close()
            except Exception as close_exc:  # pragma: no cover - best effort
                logger.debug("Failed to close detail tab: %s", close_exc)

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
                    mode="w", suffix=".js", delete=False
                ) as temp_file:
                    temp_file.write(js_code)
                    temp_file_path = temp_file.name

                try:
                    # Execute with Node.js
                    result = subprocess.run(
                        ["node", temp_file_path],
                        capture_output=True,
                        text=True,
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
            with open("job_list_nuxt.json", "w") as f:
                json.dump(job_list, f, indent=4)
            print(job_list[0])
        except Exception as exc:
            logger.error(
                "Error running job URL extraction script: %s", exc, exc_info=True
            )
            return []

        # add metadata to the job list
        for job in job_list:
            job["scrape_metadata"] = {
                "last_visited_at": datetime.now().isoformat(),
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

                # Try to close all tabs first
                try:
                    await asyncio.wait_for(
                        asyncio.get_event_loop().run_in_executor(
                            None, self.driver.close_all_tabs
                        ),
                        timeout=5.0,
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
                        timeout=10.0,
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
            if platform.system() not in ["Linux", "Darwin"]:
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
                        {"capture_output": True, "text": True},
                    ),
                    timeout=5.0,
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
                            logger.debug(
                                f"Permission denied killing process {pid_str}: {e}"
                            )

                    if killed_count > 0:
                        logger.info(
                            f"Cleaned up {killed_count} hanging Chrome processes"
                        )

            except subprocess.CalledProcessError:
                # pgrep found no processes, which is fine
                logger.debug("No hanging Chrome processes found")
            except asyncio.TimeoutError:
                logger.warning("Process cleanup timed out")

        except Exception as cleanup_exc:
            logger.debug("Process cleanup failed: %s", cleanup_exc)

        logger.info("Service cleanup completed")

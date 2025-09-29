"""Upwork job listing scraper using Botasaurus for web scraping.

This scraper scrapes Upwork job listings using structured search parameters
and provides both basic and detailed job information extraction.
"""

from __future__ import annotations

import json
import logging
import os
from datetime import datetime
from pathlib import Path
from typing import Tuple

from .core.service import UpworkJobService
from .schemas.input import ActorInput
from botasaurus_driver.exceptions import CloudflareDetectionException


logger = logging.getLogger(__name__)


class DateTimeEncoder(json.JSONEncoder):
    """Custom JSON encoder that handles datetime objects."""

    def default(self, obj):
        if isinstance(obj, datetime):
            return obj.isoformat()
        return super().default(obj)


class SimpleDataStore:
    """Simple data storage for scraped jobs with context manager support."""

    def __init__(self, storage_dir: str = "./storage"):
        self.storage_dir = Path(storage_dir)
        self.datasets_dir = self.storage_dir / "datasets" / "default"
        self.key_value_dir = self.storage_dir / "key_value_stores" / "default"

        # Create directories
        self.datasets_dir.mkdir(parents=True, exist_ok=True)
        self.key_value_dir.mkdir(parents=True, exist_ok=True)

        self.job_counter = 0
        self._closed = False

    async def push_data(self, data: dict) -> None:
        """Store job data to JSON file."""
        if self._closed:
            raise RuntimeError("Cannot push data to closed SimpleDataStore")

        try:
            self.job_counter += 1
            filename = f"{self.job_counter:09d}.json"
            file_path = self.datasets_dir / filename

            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(data, f, indent=2, ensure_ascii=False, cls=DateTimeEncoder)
        except Exception as e:
            logging.getLogger(__name__).error(f"Failed to push data: {e}")
            raise

    async def set_value(self, key: str, value: dict) -> None:
        """Store key-value data."""
        if self._closed:
            raise RuntimeError("Cannot set value on closed SimpleDataStore")

        try:
            file_path = self.key_value_dir / f"{key}.json"
            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(value, f, indent=2, ensure_ascii=False, cls=DateTimeEncoder)
        except Exception as e:
            logging.getLogger(__name__).error(
                f"Failed to set value for key '{key}': {e}"
            )
            raise

    def close(self) -> None:
        """Mark the data store as closed."""
        self._closed = True
        logging.getLogger(__name__).debug("SimpleDataStore closed")

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()
        return False


def prepare_scraper_environment() -> Tuple[ActorInput, SimpleDataStore]:
    """Create actor input and data store using environment defaults."""
    storage_dir = os.getenv("LOCAL_STORAGE_DIR", "./storage")
    data_store = SimpleDataStore(storage_dir)

    actor_input_raw = {
        "search_parameters": {
            "sort_by": "recency",
        },
        "max_jobs": 10,
        "debug_mode": True,
    }

    logger.info("Using default input configuration with recency sorting")
    actor_input = ActorInput(**actor_input_raw)

    if actor_input.debug_mode:
        logger.info(
            "Input parameters: %s", actor_input.model_dump_json(indent=2)
        )

    return actor_input, data_store


async def run_scraper_iteration(
    service: UpworkJobService, actor_input: ActorInput, data_store: SimpleDataStore
) -> dict:
    """Run a single scraping iteration, returning the summary payload."""
    await service.initialize()

    previous_total = len(service.comprehensive_jobs_found)

    search_urls = actor_input.build_search_urls()
    logger.info("Generated %s search URLs", len(search_urls))

    if actor_input.debug_mode:
        for i, url in enumerate(search_urls, 1):
            logger.info("URL %s: %s", i, url)

    logger.info("Starting job scraping...")

    await service.run_scraping(search_urls)

    current_total = len(service.comprehensive_jobs_found)
    processed_this_run = max(current_total - previous_total, 0)

    logger.info(
        "Successfully processed %s comprehensive jobs in this iteration",
        processed_this_run,
    )

    summary = {
        "total_jobs_found": processed_this_run,
        "search_urls_count": len(search_urls),
        "extraction_type": "comprehensive",
        "max_jobs_limit": actor_input.max_jobs,
        "processed_at": datetime.now().isoformat(),
        "input_parameters": actor_input.model_dump(),
    }

    await data_store.set_value("RUN_SUMMARY", summary)
    logger.info(
        "Completed successfully - %s comprehensive jobs processed",
        processed_this_run,
    )

    return summary


async def main() -> None:
    """Main entry point for the Upwork job listing scraper."""
    # Configure logging with websocket error filtering
    log_level = os.getenv("LOG_LEVEL", "INFO").upper()
    logging.basicConfig(
        level=getattr(logging, log_level),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    )

    # Filter out excessive websocket error logs to prevent spam
    websocket_logger = logging.getLogger("websocket")
    websocket_logger.setLevel(logging.WARNING)  # Only show warnings and above

    # Initialize variables for cleanup
    service = None
    data_store = None

    try:
        actor_input, data_store = prepare_scraper_environment()

        logger.info("Starting Upwork job listing scraper")

        # Initialize the service
        service = UpworkJobService(actor_input, data_store)

        try:
            # Use async context manager for proper resource management
            async with service:
                await run_scraper_iteration(service, actor_input, data_store)

        except CloudflareDetectionException:
            logger.error("Cloudflare detection exception")
        except KeyboardInterrupt:
            logger.info("Scraper interrupted by user")
            raise
        except Exception as e:
            logger.error(f"Scraper execution failed: {e}", exc_info=True)

            # Store error summary with safe access
            try:
                error_summary = {
                    "error": str(e),
                    "error_type": type(e).__name__,
                    "processed_at": datetime.now().isoformat(),
                    "total_jobs_processed": len(service.comprehensive_jobs_found)
                    if service
                    else 0,
                }
                if data_store:
                    await data_store.set_value("ERROR_SUMMARY", error_summary)
            except Exception as summary_error:
                logger.error(f"Failed to store error summary: {summary_error}")

            # Re-raise the original exception
            raise

    except KeyboardInterrupt:
        logger.info("Application interrupted by user")
        raise
    except Exception as e:
        logger.error(f"Critical application error: {e}", exc_info=True)
        raise
    finally:
        # Clean up local resources (service cleanup handled by async context manager)
        logger.info("Starting cleanup process...")

        # Clean up data store
        if data_store:
            try:
                data_store.close()
                logger.info("Data store closed")
            except Exception as store_error:
                logger.error(f"Data store cleanup failed: {store_error}")

        logger.info("Cleanup process completed")

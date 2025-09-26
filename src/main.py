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

from .core.service import UpworkJobService
from .realtimedb import RealtimeJobDatabase
from .schemas.input import ActorInput


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


async def main() -> None:
    """Main entry point for the Upwork job listing scraper."""
    # Configure logging with websocket error filtering
    log_level = os.getenv("LOG_LEVEL", "INFO").upper()
    logging.basicConfig(
        level=getattr(logging, log_level),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    )
    logger = logging.getLogger(__name__)

    # Filter out excessive websocket error logs to prevent spam
    websocket_logger = logging.getLogger("websocket")
    websocket_logger.setLevel(logging.WARNING)  # Only show warnings and above

    # Initialize variables for cleanup
    service = None
    rt_db = None
    data_store = None

    try:
        # Get storage directory from environment
        storage_dir = os.getenv("LOCAL_STORAGE_DIR", "./storage")
        data_store = SimpleDataStore(storage_dir)

        try:
            # If no input provided, use default configuration with recency sorting
            actor_input_raw = {
                "search_parameters": {
                    "sort_by": "recency",
                    "payment_verified": True,
                    "keywords": "python automation",
                },
                "max_jobs": 10,
                "delay_min": 2.0,
                "delay_max": 5.0,
                "debug_mode": True,
            }
            logger.info("Using default input configuration with recency sorting")

            actor_input = ActorInput(**actor_input_raw)
        except Exception as e:
            logger.error(f"Invalid input configuration: {e}")
            # Try with minimal default configuration
            try:
                logger.info("Attempting to use minimal default configuration")
                actor_input = ActorInput(
                    max_jobs=10, delay_min=2.0, delay_max=5.0, debug_mode=True
                )
            except Exception as e2:
                logger.error(f"Failed to create default configuration: {e2}")
                return

        logger.info("Starting Upwork job listing scraper")
        if actor_input.debug_mode:
            logger.info(f"Input parameters: {actor_input.model_dump_json(indent=2)}")

        # Initialize database with proper error handling
        stale_threshold = int(actor_input.delay_max)
        try:
            rt_db = RealtimeJobDatabase(stale_threshold_seconds=stale_threshold)
            logger.info(
                "Initialized job database with stale threshold: %d seconds",
                stale_threshold,
            )
        except Exception as e:
            logger.error(f"Failed to initialize database: {e}")
            raise

        # Initialize the service
        service = UpworkJobService(actor_input, rt_db, data_store)

        try:
            # Use async context manager for proper resource management
            async with service:
                # Build search URLs from parameters
                search_urls = actor_input.build_search_urls()
                logger.info(f"Generated {len(search_urls)} search URLs")

                if actor_input.debug_mode:
                    for i, url in enumerate(search_urls, 1):
                        logger.info(f"URL {i}: {url}")

                logger.info("Starting job scraping...")

                # Run the scraping - cancellation will propagate naturally
                await service.run_scraping(search_urls)

            # Log summary
            total_jobs = len(service.comprehensive_jobs_found)
            logger.info(f"Successfully processed {total_jobs} comprehensive jobs")

            # Get database statistics
            db_stats = rt_db.get_stats()

            # Store summary statistics
            summary = {
                "total_jobs_found": total_jobs,
                "search_urls_count": len(search_urls),
                "extraction_type": "comprehensive",
                "max_jobs_limit": actor_input.max_jobs,
                "processed_at": datetime.now().isoformat(),
                "input_parameters": actor_input.model_dump(),
                "database_stats": db_stats,
            }

            await data_store.set_value("RUN_SUMMARY", summary)
            logger.info(
                f"Completed successfully - {total_jobs} comprehensive jobs processed"
            )

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
        # Clean up database and data store (service cleanup handled by async context manager)
        logger.info("Starting cleanup process...")

        # Clean up database connection
        if rt_db:
            try:
                rt_db.close()
                logger.info("Database connection closed")
            except Exception as db_error:
                logger.error(f"Database cleanup failed: {db_error}")

        # Clean up data store
        if data_store:
            try:
                data_store.close()
                logger.info("Data store closed")
            except Exception as store_error:
                logger.error(f"Data store cleanup failed: {store_error}")

        logger.info("Cleanup process completed")

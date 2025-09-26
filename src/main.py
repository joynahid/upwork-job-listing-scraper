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
    """Simple data storage for scraped jobs."""
    
    def __init__(self, storage_dir: str = "./storage"):
        self.storage_dir = Path(storage_dir)
        self.datasets_dir = self.storage_dir / "datasets" / "default"
        self.key_value_dir = self.storage_dir / "key_value_stores" / "default"
        
        # Create directories
        self.datasets_dir.mkdir(parents=True, exist_ok=True)
        self.key_value_dir.mkdir(parents=True, exist_ok=True)
        
        self.job_counter = 0
        
    async def push_data(self, data: dict) -> None:
        """Store job data to JSON file."""
        self.job_counter += 1
        filename = f"{self.job_counter:09d}.json"
        file_path = self.datasets_dir / filename
        
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False, cls=DateTimeEncoder)
    
    async def set_value(self, key: str, value: dict) -> None:
        """Store key-value data."""
        file_path = self.key_value_dir / f"{key}.json"
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(value, f, indent=2, ensure_ascii=False, cls=DateTimeEncoder)


async def main() -> None:
    """Main entry point for the Upwork job listing scraper."""
    # Configure logging
    log_level = os.getenv('LOG_LEVEL', 'INFO').upper()
    logging.basicConfig(
        level=getattr(logging, log_level), 
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    logger = logging.getLogger(__name__)

    # Get storage directory from environment
    storage_dir = os.getenv('LOCAL_STORAGE_DIR', './storage')
    data_store = SimpleDataStore(storage_dir)

    # Get input from environment or use defaults
    input_file = Path(storage_dir) / "key_value_stores" / "default" / "INPUT.json"
    actor_input_raw = {}
    
    if input_file.exists():
        try:
            with open(input_file, 'r', encoding='utf-8') as f:
                actor_input_raw = json.load(f)
        except Exception as e:
            logger.warning(f"Could not load input file: {e}")

    try:
        actor_input = ActorInput(**actor_input_raw)
    except Exception as e:
        logger.error(f"Invalid input configuration: {e}")
        return

    logger.info("Starting Upwork job listing scraper")
    if actor_input.debug_mode:
        logger.info(f"Input parameters: {actor_input.model_dump_json(indent=2)}")

    # Initialize RocksDB job database
    stale_threshold = int(actor_input.delay_max)
    rt_db = RealtimeJobDatabase(
        stale_threshold_seconds=stale_threshold  # Use max delay as stale threshold
    )
    logger.info("Initialized job database with stale threshold: %d seconds", stale_threshold)

    # Initialize the service
    service = UpworkJobService(actor_input, rt_db, data_store)

    try:
        # Initialize service
        await service.initialize()

        # Build search URLs from parameters
        search_urls = actor_input.build_search_urls()
        logger.info(f"Generated {len(search_urls)} search URLs")

        if actor_input.debug_mode:
            for i, url in enumerate(search_urls, 1):
                logger.info(f"URL {i}: {url}")

        logger.info("Starting job scraping...")

        # Run the scraping
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
        logger.info(f"Completed successfully - {total_jobs} comprehensive jobs processed")

    except Exception as e:
        logger.error(f"Scraper execution failed: {e}", exc_info=True)
        
        # Store error summary
        error_summary = {
            "error": str(e),
            "processed_at": datetime.now().isoformat(),
            "total_jobs_processed": len(service.comprehensive_jobs_found) if service else 0,
        }
        await data_store.set_value("ERROR_SUMMARY", error_summary)

    finally:
        # Clean up resources
        if service:
            await service.cleanup()

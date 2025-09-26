"""Upwork job listing scraper Apify Actor using Crawlee.

This actor scrapes Upwork job listings using structured search parameters
and provides both basic and detailed job information extraction.
"""

from __future__ import annotations

import logging
from datetime import datetime

from apify import Actor

from .core.service import UpworkJobService
from .schemas.input import ActorInput


async def main() -> None:
    """Main entry point for the Upwork job listing scraper actor."""
    async with Actor:
        # Configure logging
        logging.basicConfig(
            level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )
        logger = logging.getLogger(__name__)

        # Get and validate actor input
        actor_input_raw = await Actor.get_input() or {}

        try:
            actor_input = ActorInput(**actor_input_raw)
        except Exception as e:
            logger.error(f"Invalid actor input: {e}")
            await Actor.fail()
            return

        logger.info("Starting Upwork job listing scraper")
        if actor_input.debug_mode:
            logger.info(f"Input parameters: {actor_input.model_dump_json(indent=2)}")

        # Initialize the service
        service = UpworkJobService(actor_input)

        try:
            # Initialize Crawlee crawler
            await service.initialize()

            # Build search URLs from parameters
            search_urls = actor_input.build_search_urls()
            logger.info(f"Generated {len(search_urls)} search URLs")

            if actor_input.debug_mode:
                for i, url in enumerate(search_urls, 1):
                    logger.info(f"URL {i}: {url}")

            # Set initial status
            await Actor.set_status_message("Starting job scraping...")

            # Run the Crawlee-based scraping
            await service.run_scraping(search_urls)

            # Log summary
            total_jobs = len(service.comprehensive_jobs_found)
            logger.info(f"Successfully processed {total_jobs} comprehensive jobs")

            # Store summary statistics
            summary = {
                "total_jobs_found": total_jobs,
                "search_urls_count": len(search_urls),
                "extraction_type": "comprehensive",
                "max_jobs_limit": actor_input.max_jobs,
                "processed_at": datetime.now().isoformat(),
                "input_parameters": actor_input.model_dump(),
            }

            await Actor.set_value("RUN_SUMMARY", summary)
            await Actor.set_status_message(f"Completed successfully - {total_jobs} comprehensive jobs processed")
            logger.info("Actor completed successfully")

        except Exception as e:
            logger.error(f"Actor execution failed: {e}", exc_info=True)
            await Actor.set_status_message(f"Failed: {str(e)}")
            await Actor.fail()

        finally:
            # Clean up resources
            await service.cleanup()

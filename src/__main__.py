import asyncio
import random
import signal
import sys
import logging
import time
from datetime import datetime
from typing import Optional

from botasaurus_driver.exceptions import CloudflareDetectionException

from .core.service import UpworkJobService
from .main import prepare_scraper_environment, run_scraper_iteration

# Set up logging for the entry point
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global variables for graceful shutdown
_main_task: Optional[asyncio.Task] = None
_shutdown_requested = False


def signal_handler(signum, frame):
    """Handle shutdown signals gracefully."""
    global _shutdown_requested
    logger.info(f"Received signal {signum}, initiating graceful shutdown...")
    _shutdown_requested = True

    # Cancel the main task if it's running
    if _main_task and not _main_task.done():
        logger.info("Cancelling main task...")
        _main_task.cancel()

    logger.info("Shutdown signal processed")


# Set up signal handlers
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)


async def run_main_with_shutdown():
    """Run scraping iterations while keeping a persistent browser session."""
    global _main_task

    actor_input = None
    data_store = None
    service: Optional[UpworkJobService] = None

    try:
        actor_input, data_store = prepare_scraper_environment()
        service = UpworkJobService(actor_input, data_store)

        logger.info("Persistent scraping loop started; browser session will be reused")

        while not _shutdown_requested:
            _main_task = asyncio.create_task(
                run_scraper_iteration(service, actor_input, data_store)
            )
            try:
                await _main_task
            except CloudflareDetectionException:
                logger.error("Cloudflare detection exception during iteration")
            except asyncio.CancelledError:
                logger.info("Main task was cancelled - cleanup completed")
                raise
            except Exception as e:
                logger.error("Scraper execution failed: %s", e, exc_info=True)
                if data_store is not None:
                    error_summary = {
                        "error": str(e),
                        "error_type": type(e).__name__,
                        "processed_at": datetime.now().isoformat(),
                        "total_jobs_processed": service.total_jobs_processed if service else 0,
                    }
                    try:
                        await data_store.set_value("ERROR_SUMMARY", error_summary)
                    except Exception as summary_error:
                        logger.error(
                            "Failed to store error summary: %s", summary_error
                        )
                raise
            finally:
                _main_task = None

            if _shutdown_requested:
                break

            delay_seconds = random.randint(40, 60)
            logger.info(
                "Sleeping %s seconds before restarting iteration", delay_seconds
            )
            await _sleep_with_shutdown_check(delay_seconds)

    except asyncio.CancelledError:
        logger.info("Main runner cancelled - exiting")
        raise
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt")
        return 130
    except Exception as e:
        logger.error("Script failed with error: %s", e, exc_info=True)
        return 1
    finally:
        if service is not None:
            try:
                await service.cleanup()
            except Exception as cleanup_error:
                logger.error("Service cleanup failed: %s", cleanup_error)

        if data_store is not None:
            try:
                data_store.close()
            except Exception as store_error:
                logger.error("Data store cleanup failed: %s", store_error)

    return 0


async def _sleep_with_shutdown_check(delay_seconds: int) -> None:
    """Sleep in short intervals so shutdown signals can end the wait early."""
    end_time = time.monotonic() + delay_seconds
    while not _shutdown_requested:
        remaining = end_time - time.monotonic()
        if remaining <= 0:
            break
        await asyncio.sleep(min(1.0, remaining))


try:
    # Execute the Actor entry point with proper cancellation handling
    exit_code = asyncio.run(run_main_with_shutdown())
    sys.exit(exit_code)

except KeyboardInterrupt:
    logger.info("Script interrupted by user")
    sys.exit(130)  # Standard SIGINT exit code

except asyncio.CancelledError:
    logger.info("Script was cancelled - exiting gracefully")
    sys.exit(130)

except Exception as e:
    logger.error(f"Unexpected error in main runner: {e}", exc_info=True)
    sys.exit(1)

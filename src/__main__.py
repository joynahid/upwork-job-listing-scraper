import asyncio
import random
import signal
import sys
import logging
import time
from typing import Optional

from .main import main

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
    """Run main with proper async cancellation handling."""
    global _main_task

    try:
        while not _shutdown_requested:
            _main_task = asyncio.create_task(main())
            try:
                await _main_task
            except asyncio.CancelledError:
                logger.info("Main task was cancelled - cleanup completed")
                raise
            finally:
                _main_task = None

            if _shutdown_requested:
                break

            delay_seconds = random.randint(40, 60)
            logger.info("Sleeping %s seconds before restarting main", delay_seconds)
            await _sleep_with_shutdown_check(delay_seconds)
    except asyncio.CancelledError:
        logger.info("Main runner cancelled - exiting")
        raise
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt")
        return 130
    except Exception as e:
        logger.error(f"Script failed with error: {e}", exc_info=True)
        return 1


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

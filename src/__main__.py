import asyncio
import signal
import sys
import logging
import threading
from typing import Optional

from .main import main

# Set up logging for the entry point
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global variables for graceful shutdown
_main_task: Optional[asyncio.Task] = None
_shutdown_event = threading.Event()


def signal_handler(signum, frame):
    """Handle shutdown signals gracefully."""
    logger.info(f"Received signal {signum}, initiating graceful shutdown...")
    _shutdown_event.set()

    # Cancel the main task if it's running
    if _main_task and not _main_task.done():
        logger.info("Cancelling main task...")
        _main_task.cancel()

    # Give a moment for cleanup, then exit
    logger.info("Shutdown signal processed")


# Set up signal handlers
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)


async def run_main_with_shutdown():
    """Run main with proper shutdown handling."""
    global _main_task

    try:
        _main_task = asyncio.create_task(main())
        await _main_task
        return 0
    except asyncio.CancelledError:
        logger.info("Main task was cancelled - performing cleanup...")
        return 130  # Standard exit code for SIGINT
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt")
        return 130
    except Exception as e:
        logger.error(f"Script failed with error: {e}", exc_info=True)
        return 1


try:
    # Execute the Actor entry point with a timeout and shutdown handling
    exit_code = asyncio.run(
        asyncio.wait_for(run_main_with_shutdown(), timeout=1800)  # 30 minute timeout
    )
    sys.exit(exit_code)

except asyncio.TimeoutError:
    logger.error("Script timed out after 30 minutes")
    sys.exit(124)  # Standard timeout exit code

except KeyboardInterrupt:
    logger.info("Script interrupted by user")
    sys.exit(130)  # Standard SIGINT exit code

except Exception as e:
    logger.error(f"Unexpected error in main runner: {e}", exc_info=True)
    sys.exit(1)

import asyncio
import signal
import sys
import logging

from .main import main

# Set up logging for the entry point
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def signal_handler(signum, frame):
    """Handle termination signals gracefully."""
    logger.info(f"Received signal {signum}, terminating...")
    sys.exit(0)

# Set up signal handlers
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

try:
    # Execute the Actor entry point with a timeout
    asyncio.run(asyncio.wait_for(main(), timeout=1800))  # 30 minute timeout
except asyncio.TimeoutError:
    logger.error("Script timed out after 30 minutes")
    sys.exit(1)
except KeyboardInterrupt:
    logger.info("Script interrupted by user")
    sys.exit(0)
except Exception as e:
    logger.error(f"Script failed with error: {e}")
    sys.exit(1)

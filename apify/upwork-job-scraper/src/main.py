"""Upwork Job Scraper API Wrapper for Apify.

This Actor connects to a Go API backend to fetch Upwork job listings in real-time.
It processes jobs one by one and saves them to the Apify dataset with proper formatting.
Supports both standard mode and Apify Standby mode for HTTP API access.
"""

from __future__ import annotations

import asyncio
import threading
from http.server import ThreadingHTTPServer

from apify import Actor

try:
    # Try relative imports first (when run as module)
    from .api_wrapper import UpworkJobAPIWrapper
    from .config import ActorConfig
    from .http_handler import UpworkJobStandbyHandler
    from .job_processor import JobProcessor
except ImportError:
    # Fall back to absolute imports (when run directly)
    from api_wrapper import UpworkJobAPIWrapper
    from config import ActorConfig
    from http_handler import UpworkJobStandbyHandler
    from job_processor import JobProcessor


async def run_standard_mode() -> None:
    """Run the Actor in standard mode (original functionality)."""
    # Get Actor input and create configuration
    actor_input = await Actor.get_input() or {}
    config = ActorConfig(actor_input)
    
    # Validate configuration
    if not config.validate():
        Actor.log.error("❌ API key is required. Please set 'API_KEY' environment variable.")
        await Actor.exit()
        return

    Actor.log.info("🚀 Starting Upwork Job Scraper API Wrapper (Standard Mode)")
    config.log_config()

    # Initialize components
    api_wrapper = UpworkJobAPIWrapper(config.api_endpoint, config.api_key, config.debug_mode)
    job_processor = JobProcessor(api_wrapper)

    try:
        # Process jobs
        await job_processor.process_jobs_batch(config.max_jobs, config.filters, config.debug_mode)

    except Exception as e:
        Actor.log.error(f"❌ Actor execution failed: {e}")
        raise

    finally:
        # Clean up
        await api_wrapper.close()
        Actor.log.info("🧹 Cleanup completed")


async def run_standby_mode() -> None:
    """Run the Actor in Apify Standby mode as an HTTP server."""
    # Create configuration from environment variables only
    config = ActorConfig()

    # Validate configuration
    if not config.validate():
        Actor.log.error("❌ API key is required. Please set 'API_KEY' environment variable.")
        await Actor.exit()
        return

    Actor.log.info("🚀 Starting Upwork Job Scraper API Wrapper (Standby Mode)")
    Actor.log.info(f"📊 API Endpoint: {config.api_endpoint}")
    Actor.log.info(f"📊 Debug Mode: {config.debug_mode}")
    standby_port = getattr(Actor.config, "web_server_port", Actor.config.standby_port)
    Actor.log.info(f"🌐 Standby Port: {standby_port}")

    # Initialize components
    api_wrapper = UpworkJobAPIWrapper(config.api_endpoint, config.api_key, config.debug_mode)
    job_processor = JobProcessor(api_wrapper)

    try:
        # Create a handler factory that passes the job_processor to each request
        event_loop = asyncio.get_running_loop()

        def handler_factory(*args, **kwargs):
            return UpworkJobStandbyHandler(job_processor, event_loop, *args, **kwargs)

        http_server: ThreadingHTTPServer | None = None
        server_thread: threading.Thread | None = None

        try:
            http_server = ThreadingHTTPServer(('', standby_port), handler_factory)

            Actor.log.info(f"🌐 HTTP server started on port {standby_port}")
            Actor.log.info("📋 Ready to handle requests and readiness probes")

            server_thread = threading.Thread(
                target=http_server.serve_forever,
                name="apify-standby-http-server",
                daemon=True,
            )
            server_thread.start()

            stop_event = asyncio.Event()

            try:
                await stop_event.wait()
            except asyncio.CancelledError:
                Actor.log.info("🛑 Standby mode shutdown signal received")
                raise
        finally:
            if http_server is not None:
                http_server.shutdown()
                http_server.server_close()
            if server_thread is not None and server_thread.is_alive():
                server_thread.join(timeout=5)

    except Exception as e:
        Actor.log.error(f"❌ Standby server failed: {e}")
        raise
    finally:
        # Clean up
        await api_wrapper.close()
        Actor.log.info("🧹 Cleanup completed")


async def main() -> None:
    """Main entry point for the Apify Actor.
    
    Supports both standard mode and Apify Standby mode based on Actor.config.meta_origin.
    """
    async with Actor:
        # Check if Actor was started in Standby mode
        if Actor.config.meta_origin == 'STANDBY':
            Actor.log.info("🔄 Detected Standby mode - starting HTTP server")
            await run_standby_mode()
        else:
            Actor.log.info("🔄 Detected Standard mode - running one-time job scraping")
            await run_standard_mode()


if __name__ == "__main__":
    asyncio.run(main())

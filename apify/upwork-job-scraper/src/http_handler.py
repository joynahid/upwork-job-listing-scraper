"""HTTP request handler for Apify Standby mode."""

import asyncio
import json
from http.server import BaseHTTPRequestHandler
from typing import Any, Dict
from urllib.parse import urlparse

from apify import Actor

try:
    from .job_processor import JobProcessor
    from .utils import ParameterParser
except ImportError:
    from job_processor import JobProcessor
    from utils import ParameterParser


class UpworkJobStandbyHandler(BaseHTTPRequestHandler):
    """HTTP request handler for Apify Standby mode."""

    def __init__(
        self,
        job_processor: JobProcessor,
        event_loop: asyncio.AbstractEventLoop,
        *args,
        **kwargs,
    ):
        """Initialize the handler with job processor."""
        self.job_processor = job_processor
        self.event_loop = event_loop
        super().__init__(*args, **kwargs)
    
    def do_GET(self) -> None:
        """Handle GET requests for job scraping and readiness probes."""
        # Handle Apify standby readiness probe
        if 'x-apify-container-server-readiness-probe' in self.headers:
            Actor.log.info('📋 Readiness probe received')
            self._send_response(200, 'text/plain', b'ok')
            return
        
        try:
            # Parse URL and query parameters
            parsed_url = urlparse(self.path)
            params = ParameterParser.parse_query_params(parsed_url.query)
            
            Actor.log.info(f'🌐 HTTP request received: maxJobs={params["max_jobs"]}, filters={params["filters"]}')
            
            # Run the job scraping on the main asyncio event loop
            future = asyncio.run_coroutine_threadsafe(
                self.job_processor.process_jobs_batch(
                    params["max_jobs"],
                    params["filters"],
                    params["debug_mode"],
                ),
                self.event_loop,
            )

            result = future.result()

            # Send successful response
            response_data = {
                "success": True,
                "message": f"Successfully processed {result['processed_count']} jobs",
                "data": result,
            }

            self._send_json_response(200, response_data)

        except ValueError as e:
            Actor.log.warning(f"⚠️ Bad request: {e}")

            error_response = {
                "success": False,
                "error": str(e),
                "error_type": type(e).__name__
            }

            self._send_json_response(400, error_response)

        except Exception as e:
            Actor.log.error(f"❌ Error handling HTTP request: {e}")

            error_response = {
                "success": False,
                "error": str(e),
                "error_type": type(e).__name__
            }

            self._send_json_response(500, error_response)
    
    def _send_response(self, status_code: int, content_type: str, content: bytes) -> None:
        """Send HTTP response with given status, content type and content."""
        self.send_response(status_code)
        self.send_header('Content-Type', content_type)
        self.end_headers()
        self.wfile.write(content)
    
    def _send_json_response(self, status_code: int, data: Dict[str, Any]) -> None:
        """Send JSON HTTP response."""
        content = json.dumps(data, indent=2).encode()
        self._send_response(status_code, 'application/json', content)
    
    def log_message(self, format, *args):
        """Override to use Apify logging instead of stderr."""
        Actor.log.info(f"🌐 HTTP: {format % args}")

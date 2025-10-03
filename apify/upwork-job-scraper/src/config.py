"""Configuration management for the Upwork Job Scraper Actor."""

import os
from typing import Any, Dict, Optional

from apify import Actor

try:
    from .url_parser import UpworkURLParser
except ImportError:
    from url_parser import UpworkURLParser


class ActorConfig:
    """Configuration management for the Upwork Job Scraper Actor."""
    
    def __init__(self, actor_input: Optional[Dict[str, Any]] = None):
        """Initialize configuration from environment variables and actor input."""
        self._has_actor_input = actor_input is not None
        self.actor_input = actor_input or {}
        
        # Environment variables
        self.api_key = os.getenv("API_KEY")
        self.api_endpoint = os.getenv("API_ENDPOINT", "https://upworkjobscraperapi.nahidhq.com")
        self.debug_mode = os.getenv("DEBUG_MODE", "false").lower() == "true"
        
        # Actor input parameters
        self.upwork_url = self.actor_input.get("upworkUrl", "")
        self.max_jobs = self.actor_input.get("maxJobs", 20)
        
        # Build request payload for the Go API
        self.filters = self._build_request_payload()

    def _build_request_payload(self) -> Dict[str, Any]:
        """Build the request payload expected by the Go API."""
        if not self.upwork_url:
            return {}

        if self.debug_mode:
            try:
                parsed_filters = UpworkURLParser.parse_url(self.upwork_url)
                Actor.log.info(
                    f"✅ Parsed Upwork URL into {len(parsed_filters)} derived filters"
                )
                Actor.log.debug(f"   Derived filters: {parsed_filters}")
            except Exception as exc:  # pragma: no cover - best effort logging
                Actor.log.warning(f"⚠️ Unable to parse Upwork URL for diagnostics: {exc}")

        return {"upwork_url": self.upwork_url}
    
    def validate(self) -> bool:
        """Validate required configuration."""
        if not self.api_key:
            Actor.log.error("❌ API key is required")
            return False

        if self._has_actor_input and not self.upwork_url:
            Actor.log.error("❌ Upwork URL is required")
            return False

        return True
    
    def log_config(self) -> None:
        """Log configuration details."""
        Actor.log.info("📊 Configuration:")
        Actor.log.info(f"   Upwork URL: {self.upwork_url}")
        Actor.log.info(f"   Max Jobs: {self.max_jobs}")
        Actor.log.info(f"   Debug Mode: {self.debug_mode}")
        
        if self.debug_mode:
            Actor.log.info(f"   Request Payload: {self.filters}")

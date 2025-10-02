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
        self.actor_input = actor_input or {}
        
        # Environment variables
        self.api_key = os.getenv("API_KEY")
        self.api_endpoint = os.getenv("API_ENDPOINT", "https://upworkjobscraperapi.nahidhq.com")
        self.debug_mode = os.getenv("DEBUG_MODE", "false").lower() == "true"
        
        # Actor input parameters
        self.upwork_url = self.actor_input.get("upworkUrl", "")
        self.max_jobs = self.actor_input.get("maxJobs", 20)
        
        # Parse Upwork URL to extract filters
        self.filters = self._parse_upwork_url()
    
    def _parse_upwork_url(self) -> Dict[str, Any]:
        """
        Parse Upwork URL and extract filters in Go API format.
        
        Returns:
            Dictionary of filters formatted for Go API
        """
        if not self.upwork_url:
            if self.debug_mode:
                Actor.log.warning("⚠️ No Upwork URL provided, using default filters")
            return {}
        
        try:
            filters = UpworkURLParser.parse_url(self.upwork_url)
            if self.debug_mode:
                Actor.log.info(f"✅ Parsed {len(filters)} filters from URL")
            return filters
        except Exception as e:
            Actor.log.error(f"❌ Failed to parse Upwork URL: {e}")
            return {}
    
    def validate(self) -> bool:
        """Validate required configuration."""
        if not self.api_key:
            Actor.log.error("❌ API key is required")
            return False
        
        if not self.upwork_url:
            Actor.log.warning("⚠️ No Upwork URL provided - will fetch jobs without filters")
        
        return True
    
    def log_config(self) -> None:
        """Log configuration details."""
        Actor.log.info("📊 Configuration:")
        Actor.log.info(f"   Upwork URL: {self.upwork_url}")
        Actor.log.info(f"   Max Jobs: {self.max_jobs}")
        Actor.log.info(f"   Debug Mode: {self.debug_mode}")
        
        if self.debug_mode:
            Actor.log.info(f"   Parsed Filters: {self.filters}")

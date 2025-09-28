"""Configuration management for the Upwork Job Scraper Actor."""

import os
from typing import Any, Dict, Optional

from apify import Actor


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
        self.max_jobs = self.actor_input.get("maxJobs", 20)
        
        # Build filters from actor input
        self.filters = self._build_filters()
    
    def _build_filters(self) -> Dict[str, Any]:
        """Build filters object from actor input fields."""
        filters = {}
        filter_fields = [
            "paymentVerified", "category", "categoryGroup", "status", 
            "jobType", "contractorTier", "country", "tags", 
            "postedAfter", "postedBefore", "budgetMin", "budgetMax", "sort"
        ]
        
        for field in filter_fields:
            value = self.actor_input.get(field)
            if value is not None and value != "":
                if field == "tags" and isinstance(value, str):
                    # Convert comma-separated string to list for tags
                    filters[field] = [tag.strip() for tag in value.split(",") if tag.strip()]
                else:
                    filters[field] = value
        
        return filters
    
    def validate(self) -> bool:
        """Validate required configuration."""
        return bool(self.api_key)
    
    def log_config(self) -> None:
        """Log configuration details."""
        if self.debug_mode:
            Actor.log.info("📊 Configuration:")
            Actor.log.info(f"   Max Jobs: {self.max_jobs}")
            Actor.log.info(f"   Debug Mode: {self.debug_mode}")
            Actor.log.info(f"   Filters: {self.filters}")

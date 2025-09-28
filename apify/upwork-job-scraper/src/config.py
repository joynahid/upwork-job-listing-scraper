"""Configuration management for the Upwork Job Scraper Actor."""

import os
from typing import Any, Dict, List, Optional

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
            "offset",
            "paymentVerified",
            "category",
            "categoryGroup",
            "status",
            "jobType",
            "contractorTier",
            "country",
            "tags",
            "skills",
            "postedAfter",
            "postedBefore",
            "lastVisitedAfter",
            "budgetMin",
            "budgetMax",
            "hourlyMin",
            "hourlyMax",
            "durationLabel",
            "engagement",
            "buyerTotalSpentMin",
            "buyerTotalSpentMax",
            "buyerTotalAssignmentsMin",
            "buyerTotalAssignmentsMax",
            "buyerTotalJobsWithHiresMin",
            "buyerTotalJobsWithHiresMax",
            "sort",
        ]
        
        for field in filter_fields:
            value = self.actor_input.get(field)
            if value is None:
                continue

            if isinstance(value, str) and value.strip() == "":
                continue

            if field == "offset" and isinstance(value, int) and value <= 0:
                continue

            if field in {"tags", "skills"}:
                tokens: List[str] = []
                if isinstance(value, str):
                    tokens = [token.strip() for token in value.split(",") if token.strip()]
                elif isinstance(value, list):
                    tokens = [str(token).strip() for token in value if str(token).strip()]
                if tokens:
                    filters[field] = tokens
                continue

            if field == "country" and isinstance(value, str):
                filters[field] = value.upper()
                continue

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

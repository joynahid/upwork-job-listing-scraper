"""Utility functions for the Upwork Job Scraper Actor."""

from typing import Any, Dict
from urllib.parse import parse_qs


class ParameterParser:
    """Utility class for parsing parameters from different sources."""
    
    @staticmethod
    def parse_query_params(query_string: str) -> Dict[str, Any]:
        """Parse query parameters into a structured format."""
        query_params = parse_qs(query_string)
        
        # Extract basic parameters
        max_jobs_raw = query_params.get('maxJobs', [20])[0]
        try:
            max_jobs = int(max_jobs_raw)
        except (TypeError, ValueError):
            max_jobs = 20
        debug_mode = query_params.get('debug', ['false'])[0].lower() == 'true'

        upwork_url = query_params.get('upworkUrl', [''])[0] or query_params.get('upwork_url', [''])[0]

        filters: Dict[str, Any] = {}
        if upwork_url:
            filters['upwork_url'] = upwork_url

        if not filters:
            raise ValueError("Query parameter 'upworkUrl' (or 'upwork_url') is required")

        return {
            "max_jobs": max_jobs,
            "debug_mode": debug_mode,
            "filters": filters
        }

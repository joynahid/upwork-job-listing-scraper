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
        max_jobs = int(query_params.get('maxJobs', [20])[0])
        debug_mode = query_params.get('debug', ['false'])[0].lower() == 'true'
        
        # Build filters from query parameters
        filters = {}
        filter_fields = [
            "paymentVerified", "category", "categoryGroup", "status", 
            "jobType", "contractorTier", "country", "tags", 
            "postedAfter", "postedBefore", "budgetMin", "budgetMax", "sort"
        ]
        
        for field in filter_fields:
            if field in query_params:
                value = query_params[field][0]
                if value and value != "":
                    if field == "tags":
                        # Convert comma-separated string to list for tags
                        filters[field] = [tag.strip() for tag in value.split(",") if tag.strip()]
                    else:
                        filters[field] = value
        
        return {
            "max_jobs": max_jobs,
            "debug_mode": debug_mode,
            "filters": filters
        }

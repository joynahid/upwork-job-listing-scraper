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
            if field in query_params:
                value = query_params[field][0]
                if value and value != "":
                    if field in {"tags", "skills"}:
                        # Convert comma-separated string to list for tags or skills
                        filters[field] = [token.strip() for token in value.split(",") if token.strip()]
                    elif field == "offset":
                        try:
                            filters[field] = int(value)
                        except ValueError:
                            continue
                    else:
                        filters[field] = value

        return {
            "max_jobs": max_jobs,
            "debug_mode": debug_mode,
            "filters": filters
        }

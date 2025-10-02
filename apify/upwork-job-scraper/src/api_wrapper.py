"""API wrapper for communicating with the Upwork Job Go API."""

from typing import Any, Dict

import httpx
from apify import Actor


class UpworkJobAPIWrapper:
    """Wrapper class for the Upwork Job Go API."""

    def __init__(self, api_endpoint: str, api_key: str, debug_mode: bool = False):
        """Initialize the API wrapper.

        Args:
            api_endpoint: The Go API endpoint URL
            api_key: API key for authentication
            debug_mode: Enable debug logging
        """
        self.api_endpoint = api_endpoint.rstrip("/")
        self.api_key = api_key
        self.debug_mode = debug_mode

        # Create HTTP client with timeout
        self.client = httpx.AsyncClient(
            timeout=30.0, headers={"X-API-KEY": self.api_key}
        )

    async def fetch_jobs(self, max_jobs: int, filters: Dict[str, Any] = None) -> Dict[str, Any]:
        """Fetch jobs from the Go API.

        Args:
            max_jobs: Maximum number of jobs to fetch
            filters: Optional filters to apply to the job search (already in Go API format)

        Returns:
            API response containing job data

        Raises:
            httpx.HTTPError: If API request fails
        """
        jobs_url = f"{self.api_endpoint}/jobs"
        params = {"limit": max_jobs} if max_jobs else {}
        
        # Add filters to params if provided
        # Filters are already in Go API format from the URL parser
        if filters:
            for filter_key, filter_value in filters.items():
                if filter_value is not None:
                    # Handle comma-separated values
                    if isinstance(filter_value, list):
                        params[filter_key] = ",".join(str(item) for item in filter_value)
                    else:
                        params[filter_key] = filter_value

        if self.debug_mode:
            Actor.log.info(f"🔍 Fetching jobs from: {jobs_url} (limit: {max_jobs})")
            if params:
                Actor.log.info(f"🎯 Parameters: {params}")

        try:
            response = await self.client.get(jobs_url, params=params)
            response.raise_for_status()

            data = response.json()

            if not data.get("success", False):
                raise Exception(
                    f"API returned error: {data.get('message', 'Unknown error')}"
                )

            if self.debug_mode:
                Actor.log.info(f"✅ Successfully fetched {data.get('count', 0)} jobs")

            return data

        except httpx.HTTPError as e:
            Actor.log.error(f"❌ HTTP error fetching jobs: {e}")
            raise
        except Exception as e:
            Actor.log.error(f"❌ Error fetching jobs: {e}")
            raise

    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()

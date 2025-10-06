"""HTTP client for posting scraped jobs to the Go API."""

import logging
import os
from typing import Any

import httpx

logger = logging.getLogger(__name__)


class UpworkJobAPIClient:
    """Client for posting scraped job data to the Go API."""

    def __init__(self):
        """Initialize API client with configuration from environment."""
        self.base_url = os.getenv("UPWORK_API_URL", "https://upworkapi.upfindr.app")
        self.api_key = os.getenv("API_KEY", "5d1bc44510881442")

        if not self.api_key:
            raise ValueError("API_KEY environment variable is required")

        self.headers = {
            "X-API-KEY": self.api_key,
            "Content-Type": "application/json",
        }

        # Configure httpx client with retries
        self.client = httpx.AsyncClient(
            base_url=self.base_url,
            headers=self.headers,
            timeout=30.0,
        )

        logger.info("API client initialized: base_url=%s", self.base_url)

    async def ingest_jobs(self, jobs: list[dict[str, Any]]) -> dict[str, Any]:
        """
        Post a batch of scraped jobs to the API.

        Args:
            jobs: List of job dictionaries to ingest

        Returns:
            API response as dict

        Raises:
            httpx.HTTPError: If the API request fails
        """
        if not jobs:
            logger.warning("Empty jobs list provided to ingest_jobs")
            return {"success": False, "message": "No jobs to ingest"}

        logger.info("Posting %d jobs to API at %s/ingest/jobs", len(jobs), self.base_url)

        try:
            response = await self.client.post(
                "/ingest/jobs",
                json=jobs,
            )
            response.raise_for_status()

            result = response.json()
            logger.info(
                "API response: success=%s, message=%s, count=%s",
                result.get("success"),
                result.get("message"),
                result.get("count"),
            )

            return result

        except httpx.HTTPStatusError as e:
            logger.error(
                "API request failed with status %d: %s",
                e.response.status_code,
                e.response.text,
            )
            raise
        except httpx.RequestError as e:
            logger.error("API request error: %s", str(e))
            raise

    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()
        logger.debug("API client closed")

    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.close()
        return False

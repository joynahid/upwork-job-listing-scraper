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
            filters: Optional filters to apply to the job search

        Returns:
            API response containing job data

        Raises:
            httpx.HTTPError: If API request fails
        """
        jobs_url = f"{self.api_endpoint}/jobs"
        params = {"limit": max_jobs} if max_jobs else {}
        
        # Add filters to params if provided
        if filters:
            # Map filter names to API parameter names
            filter_mapping = {
                "offset": "offset",
                "paymentVerified": "payment_verified",
                "categoryGroup": "category_group",
                "jobType": "job_type",
                "contractorTier": "contractor_tier",
                "postedAfter": "posted_after",
                "postedBefore": "posted_before",
                "lastVisitedAfter": "last_visited_after",
                "budgetMin": "budget_min",
                "budgetMax": "budget_max",
                "hourlyMin": "hourly_min",
                "hourlyMax": "hourly_max",
                "durationLabel": "duration_label",
                "buyerTotalSpentMin": "buyer.total_spent_min",
                "buyerTotalSpentMax": "buyer.total_spent_max",
                "buyerTotalAssignmentsMin": "buyer.total_assignments_min",
                "buyerTotalAssignmentsMax": "buyer.total_assignments_max",
                "buyerTotalJobsWithHiresMin": "buyer.total_jobs_with_hires_min",
                "buyerTotalJobsWithHiresMax": "buyer.total_jobs_with_hires_max",
                "isContractToHire": "is_contract_to_hire",
                "numberOfPositionsMin": "number_of_positions_min",
                "numberOfPositionsMax": "number_of_positions_max",
                "wasRenewed": "was_renewed",
                "hideBudget": "hide_budget",
                "proposalsTier": "proposals_tier",
                "minJobSuccessScore": "min_job_success_score",
                "minOdeskHours": "min_odesk_hours",
                "prefEnglishSkill": "pref_english_skill",
                "risingTalent": "rising_talent",
                "shouldHavePortfolio": "should_have_portfolio",
                "minHoursWeek": "min_hours_week",
            }
            
            for filter_key, filter_value in filters.items():
                if filter_value is not None:
                    api_param = filter_mapping.get(filter_key, filter_key)
                    if filter_key in {"tags", "skills"} and isinstance(filter_value, list):
                        params[api_param] = ",".join(filter_value)
                    elif isinstance(filter_value, list):
                        params[api_param] = ",".join(str(item) for item in filter_value)
                    else:
                        params[api_param] = filter_value

        if self.debug_mode:
            Actor.log.info(f"🔍 Fetching jobs from: {jobs_url} (limit: {max_jobs})")
            if params:
                Actor.log.info(f"🎯 Filters: {params}")

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

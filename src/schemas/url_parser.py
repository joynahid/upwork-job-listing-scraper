"""Parse Upwork search URLs and extract query parameters."""

from __future__ import annotations

import logging
from urllib.parse import parse_qs, urlparse

from .input import (
    BudgetType,
    ExperienceLevel,
    JobType,
    LocationFilter,
    SearchParameters,
    SortOrder,
)

logger = logging.getLogger(__name__)


class UpworkURLParser:
    """Parse Upwork search URLs into SearchParameters."""

    @staticmethod
    def parse_url(url: str) -> SearchParameters:
        """
        Parse an Upwork search URL and extract search parameters.

        Example URL:
        https://www.upwork.com/nx/search/jobs/?amount=500-&client_hires=1-9,10-&hourly_rate=25-&location=Americas,Europe&payment_verified=1&sort=recency&t=0,1&q=python%20automation

        Args:
            url: Upwork search URL

        Returns:
            SearchParameters object with extracted filters
        """
        parsed_url = urlparse(url)
        query_params = parse_qs(parsed_url.query)

        logger.debug(f"Parsing URL: {url}")
        logger.debug(f"Query params: {query_params}")

        params = {}

        # Parse keywords (q parameter)
        if 'q' in query_params:
            params['keywords'] = query_params['q'][0]

        # Parse budget (amount parameter) - format: "min-max" or "min-"
        if 'amount' in query_params:
            amount_str = query_params['amount'][0]
            min_budget, max_budget = UpworkURLParser._parse_range(amount_str)
            if min_budget is not None:
                params['min_budget'] = min_budget
            if max_budget is not None:
                params['max_budget'] = max_budget

        # Parse hourly rate - format: "min-max" or "min-"
        if 'hourly_rate' in query_params:
            rate_str = query_params['hourly_rate'][0]
            min_rate, max_rate = UpworkURLParser._parse_range(rate_str)
            if min_rate is not None:
                params['min_hourly_rate'] = min_rate
            if max_rate is not None:
                params['max_hourly_rate'] = max_rate

        # Parse client hires - format: "min-max" or "1-9,10-"
        if 'client_hires' in query_params:
            hires_str = query_params['client_hires'][0]
            # Handle comma-separated ranges (e.g., "1-9,10-")
            if ',' in hires_str:
                # Take the first range
                hires_str = hires_str.split(',')[0]
            min_hires, max_hires = UpworkURLParser._parse_range(hires_str)
            if min_hires is not None:
                params['min_client_hires'] = min_hires
            if max_hires is not None:
                params['max_client_hires'] = max_hires

        # Parse location - format: "Americas,Europe" or "United States"
        if 'location' in query_params:
            location_str = query_params['location'][0]
            # Take the first location if multiple
            location = location_str.split(',')[0]
            params['location'] = UpworkURLParser._parse_location(location)

        # Parse payment verification - format: "1" (true) or absent (false)
        if 'payment_verified' in query_params:
            params['payment_verified'] = query_params['payment_verified'][0] == '1'

        # Parse sort order
        if 'sort' in query_params:
            sort_str = query_params['sort'][0]
            params['sort_by'] = UpworkURLParser._parse_sort_order(sort_str)

        # Parse job type (t parameter) - format: "0" (hourly), "1" (fixed), "0,1" (both)
        if 't' in query_params:
            type_str = query_params['t'][0]
            params['job_type'] = UpworkURLParser._parse_job_type(type_str)

        # Parse experience level (contractor_tier parameter)
        if 'contractor_tier' in query_params:
            tier_str = query_params['contractor_tier'][0]
            params['experience_level'] = UpworkURLParser._parse_experience_level(tier_str)

        logger.info(f"Parsed parameters: {params}")

        return SearchParameters(**params)

    @staticmethod
    def _parse_range(range_str: str) -> tuple[int | None, int | None]:
        """
        Parse a range string like "500-1000" or "25-" into min/max values.

        Args:
            range_str: Range string from URL parameter

        Returns:
            Tuple of (min_value, max_value), either can be None
        """
        if not range_str or '-' not in range_str:
            return None, None

        parts = range_str.split('-')

        try:
            min_val = int(parts[0]) if parts[0] else None
        except (ValueError, IndexError):
            min_val = None

        try:
            max_val = int(parts[1]) if len(parts) > 1 and parts[1] else None
        except (ValueError, IndexError):
            max_val = None

        return min_val, max_val

    @staticmethod
    def _parse_location(location_str: str) -> LocationFilter:
        """Parse location string into LocationFilter enum."""
        location_map = {
            'Americas': LocationFilter.AMERICAS,
            'Europe': LocationFilter.EUROPE,
            'Asia': LocationFilter.ASIA,
            'Oceania': LocationFilter.OCEANIA,
            'Africa': LocationFilter.AFRICA,
            'United States': LocationFilter.US_ONLY,
        }
        return location_map.get(location_str, LocationFilter.WORLDWIDE)

    @staticmethod
    def _parse_sort_order(sort_str: str) -> SortOrder:
        """Parse sort order string into SortOrder enum."""
        sort_map = {
            'recency': SortOrder.RECENCY,
            'relevance': SortOrder.RELEVANCE,
            'client_rating': SortOrder.CLIENT_RATING,
            'budget': SortOrder.BUDGET,
        }
        return sort_map.get(sort_str, SortOrder.RECENCY)

    @staticmethod
    def _parse_job_type(type_str: str) -> JobType:
        """Parse job type string into JobType enum."""
        if type_str == '0':
            return JobType.HOURLY
        elif type_str == '1':
            return JobType.FIXED
        else:
            return JobType.ANY

    @staticmethod
    def _parse_experience_level(tier_str: str) -> ExperienceLevel:
        """Parse experience level string into ExperienceLevel enum."""
        tier_map = {
            '1': ExperienceLevel.ENTRY,
            '2': ExperienceLevel.INTERMEDIATE,
            '3': ExperienceLevel.EXPERT,
        }
        return tier_map.get(tier_str, ExperienceLevel.ANY)

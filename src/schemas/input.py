"""Pydantic schemas for actor input validation."""

from enum import Enum
from urllib.parse import urlparse

from pydantic import BaseModel, ConfigDict, Field, HttpUrl, field_validator, model_validator


class BudgetType(str, Enum):
    """Budget type options."""
    FIXED = "fixed"
    HOURLY = "hourly"
    ANY = "any"


class ExperienceLevel(str, Enum):
    """Experience level options."""
    ENTRY = "entry-level"
    INTERMEDIATE = "intermediate"
    EXPERT = "expert"
    ANY = "any"


class LocationFilter(str, Enum):
    """Location filter options."""
    WORLDWIDE = "worldwide"
    AMERICAS = "Americas"
    EUROPE = "Europe"
    ASIA = "Asia"
    OCEANIA = "Oceania"
    AFRICA = "Africa"
    US_ONLY = "United States"


class JobType(str, Enum):
    """Job type options."""
    HOURLY = "hourly"
    FIXED = "fixed-price"
    ANY = "any"


class SortOrder(str, Enum):
    """Sort order options."""
    RECENCY = "recency"
    RELEVANCE = "relevance"
    CLIENT_RATING = "client_rating"
    BUDGET = "budget"


class SearchParameters(BaseModel):
    """Search parameters for Upwork job queries."""

    model_config = ConfigDict(
        alias_generator=lambda field_name: ''.join(word.capitalize() if i > 0 else word for i, word in enumerate(field_name.split('_'))),
        populate_by_name=True
    )

    keywords: str | None = Field(
        None,
        description="Search keywords (e.g., 'python automation', 'video processing')"
    )

    min_budget: int | None = Field(
        None,
        ge=0,
        le=1000000,
        description="Minimum budget amount in USD"
    )

    max_budget: int | None = Field(
        None,
        ge=0,
        le=1000000,
        description="Maximum budget amount in USD"
    )

    min_hourly_rate: int | None = Field(
        None,
        ge=0,
        le=200,
        description="Minimum hourly rate in USD"
    )

    max_hourly_rate: int | None = Field(
        None,
        ge=0,
        le=200,
        description="Maximum hourly rate in USD"
    )

    budget_type: BudgetType = Field(
        BudgetType.ANY,
        description="Type of budget filter"
    )

    experience_level: ExperienceLevel = Field(
        ExperienceLevel.ANY,
        description="Required experience level"
    )

    location: LocationFilter = Field(
        LocationFilter.WORLDWIDE,
        description="Geographic location filter"
    )

    job_type: JobType = Field(
        JobType.ANY,
        description="Type of job (hourly vs fixed-price)"
    )

    payment_verified: bool = Field(
        True,
        description="Only include payment verified clients"
    )

    min_client_hires: int | None = Field(
        None,
        ge=0,
        le=100,
        description="Minimum number of client hires"
    )

    max_client_hires: int | None = Field(
        None,
        ge=0,
        le=100,
        description="Maximum number of client hires"
    )

    sort_by: SortOrder = Field(
        SortOrder.RECENCY,
        description="Sort order for results"
    )

    posted_within_days: int | None = Field(
        None,
        ge=1,
        le=30,
        description="Only jobs posted within X days"
    )

    @field_validator('max_budget')
    @classmethod
    def validate_max_budget(cls, v, info):
        """Validate max budget is greater than min budget."""
        if v is not None and info.data.get('min_budget') is not None:
            if v < info.data['min_budget']:
                raise ValueError('max_budget must be greater than min_budget')
        return v

    @field_validator('max_hourly_rate')
    @classmethod
    def validate_max_hourly_rate(cls, v, info):
        """Validate max hourly rate is greater than min hourly rate."""
        if v is not None and info.data.get('min_hourly_rate') is not None:
            if v < info.data['min_hourly_rate']:
                raise ValueError('max_hourly_rate must be greater than min_hourly_rate')
        return v

    @field_validator('max_client_hires')
    @classmethod
    def validate_max_client_hires(cls, v, info):
        """Validate max client hires is greater than min client hires."""
        if v is not None and info.data.get('min_client_hires') is not None:
            if v < info.data['min_client_hires']:
                raise ValueError('max_client_hires must be greater than min_client_hires')
        return v


class ActorInput(BaseModel):
    """Complete actor input schema with validation."""

    model_config = ConfigDict(
        alias_generator=lambda field_name: ''.join(word.capitalize() if i > 0 else word for i, word in enumerate(field_name.split('_'))),
        populate_by_name=True
    )

    # Search configuration
    search_parameters: SearchParameters | None = Field(
        None,
        description="Structured search parameters for building Upwork queries"
    )

    custom_search_urls: list[HttpUrl] | None = Field(
        default_factory=list,
        description="Custom Upwork search URLs to scrape (takes precedence over search_parameters)"
    )

    # Processing options
    extract_details: bool = Field(
        False,
        description="Extract detailed job information using AI parsing (requires OpenAI API key)"
    )

    max_jobs: int = Field(
        50,
        ge=1,
        le=1000,
        description="Maximum number of jobs to process"
    )

    # Rate limiting
    delay_min: float = Field(
        2.0,
        ge=1.0,
        le=60.0,
        description="Minimum delay between requests in seconds"
    )

    delay_max: float = Field(
        5.0,
        ge=1.0,
        le=60.0,
        description="Maximum delay between requests in seconds"
    )

    # Debug options
    take_screenshots: bool = Field(
        False,
        description="Take screenshots of search pages for debugging"
    )

    debug_mode: bool = Field(
        False,
        description="Enable debug logging and verbose output"
    )

    # Output options
    output_format: str = Field(
        "basic",
        pattern="^(basic|enhanced)$",
        description="Output format: 'basic' for search results only, 'enhanced' for detailed extraction"
    )

    include_raw_data: bool = Field(
        True,
        description="Include raw scraped data alongside parsed data"
    )

    @field_validator('delay_max')
    @classmethod
    def validate_delay_max(cls, v, info):
        """Validate max delay is greater than min delay."""
        if info.data.get('delay_min') is not None and v < info.data['delay_min']:
            raise ValueError('delay_max must be greater than or equal to delay_min')
        return v

    @field_validator('custom_search_urls')
    @classmethod
    def validate_upwork_urls(cls, v):
        """Validate that URLs are from Upwork domain."""
        if v:
            for url in v:
                parsed = urlparse(str(url))
                if parsed.netloc not in ['www.upwork.com', 'upwork.com']:
                    raise ValueError(f'URL must be from upwork.com domain: {url}')
        return v

    @model_validator(mode='after')
    def validate_search_input(self):
        """Validate that either search_parameters or custom_search_urls is provided."""
        if not self.search_parameters and not self.custom_search_urls:
            # Provide default search parameters if nothing is specified
            self.search_parameters = SearchParameters(
                keywords="python automation",
                min_hourly_rate=25,
                payment_verified=True
            )
        return self

    def build_search_urls(self) -> list[str]:
        """Build Upwork search URLs from parameters."""
        if self.custom_search_urls:
            return [str(url) for url in self.custom_search_urls]

        if not self.search_parameters:
            return self._get_default_urls()

        return self._build_urls_from_parameters()

    def _get_default_urls(self) -> list[str]:
        """Get default search URLs."""
        return [
            "https://www.upwork.com/nx/search/jobs/?amount=500-&client_hires=1-9,10-&hourly_rate=25-&location=Americas,Antarctica,Europe&payment_verified=1&sort=recency&t=0,1&q=(python%20AND%20automation)"
        ]

    def _build_urls_from_parameters(self) -> list[str]:
        """Build URLs from search parameters."""
        params = self.search_parameters
        base_url = "https://www.upwork.com/nx/search/jobs/?"

        query_parts = []

        # Budget filters
        if params.min_budget is not None or params.max_budget is not None:
            min_budget = params.min_budget or 0
            max_budget = params.max_budget or ""
            query_parts.append(f"amount={min_budget}-{max_budget}")

        # Hourly rate filters
        if params.min_hourly_rate is not None or params.max_hourly_rate is not None:
            min_rate = params.min_hourly_rate or 0
            max_rate = params.max_hourly_rate or ""
            query_parts.append(f"hourly_rate={min_rate}-{max_rate}")

        # Client hires filter
        if params.min_client_hires is not None or params.max_client_hires is not None:
            min_hires = params.min_client_hires or 1
            max_hires = params.max_client_hires or ""
            if max_hires:
                query_parts.append(f"client_hires={min_hires}-{max_hires}")
            else:
                query_parts.append(f"client_hires={min_hires}-")

        # Location filter
        if params.location != LocationFilter.WORLDWIDE:
            location_map = {
                LocationFilter.AMERICAS: "Americas",
                LocationFilter.EUROPE: "Europe",
                LocationFilter.ASIA: "Asia",
                LocationFilter.OCEANIA: "Oceania",
                LocationFilter.AFRICA: "Africa",
                LocationFilter.US_ONLY: "United States"
            }
            query_parts.append(f"location={location_map[params.location]}")

        # Payment verification
        if params.payment_verified:
            query_parts.append("payment_verified=1")

        # Sort order
        sort_map = {
            SortOrder.RECENCY: "recency",
            SortOrder.RELEVANCE: "relevance",
            SortOrder.CLIENT_RATING: "client_rating",
            SortOrder.BUDGET: "budget"
        }
        query_parts.append(f"sort={sort_map[params.sort_by]}")

        # Job type
        if params.job_type != JobType.ANY:
            type_map = {
                JobType.HOURLY: "t=0",
                JobType.FIXED: "t=1"
            }
            query_parts.append(type_map[params.job_type])
        else:
            query_parts.append("t=0,1")  # Both types

        # Keywords query
        if params.keywords:
            # URL encode the keywords
            import urllib.parse
            encoded_keywords = urllib.parse.quote(params.keywords)
            query_parts.append(f"q={encoded_keywords}")

        url = base_url + "&".join(query_parts)
        return [url]

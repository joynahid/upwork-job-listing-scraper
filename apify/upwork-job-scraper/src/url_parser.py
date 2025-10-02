"""Parse Upwork search URLs and convert to Go API parameters."""

from typing import Any, Dict, Optional
from urllib.parse import parse_qs, urlparse
import logging

logger = logging.getLogger(__name__)


class UpworkURLParser:
    """Parse Upwork search URLs into Go API compatible parameters."""

    # Mapping from Upwork URL params to Go API params
    PARAM_MAPPING = {
        # Search and filters
        'q': 'search',  # Search query
        'payment_verified': 'payment_verified',
        'category2_uid': 'category',
        'subcategory2_uid': 'category_group',
        
        # Job type and status
        'job_type': 'job_type',
        'contractor_tier': 'contractor_tier',
        'job_status': 'status',
        
        # Location
        'client_country': 'country',
        'location': 'country',
        
        # Skills and tags
        'skills': 'skills',
        'tags': 'tags',
        
        # Budget and rates
        'amount': 'budget',  # Fixed budget (will be split into min/max)
        'hourly_rate': 'hourly',  # Hourly rate (will be split into min/max)
        
        # Time filters
        'created_time': 'posted_after',
        
        # Client filters
        'client_hires': 'client_hires',
        'client_total_spent': 'client_spent',
        'client_total_feedback': 'client_feedback',
        
        # Job characteristics
        'duration_v3': 'duration_label',
        'workload': 'workload',
        'contract_to_hire': 'is_contract_to_hire',
        
        # Sorting
        'sort': 'sort',
    }

    @staticmethod
    def parse_url(url: str) -> Dict[str, Any]:
        """
        Parse Upwork search URL and extract parameters in Go API format.
        
        Args:
            url: Full Upwork search URL
            
        Returns:
            Dictionary of parameters formatted for Go API
            
        Example:
            Input: https://www.upwork.com/nx/search/jobs/?q=python&hourly_rate=50-100&payment_verified=1
            Output: {'search': 'python', 'hourly_min': 50, 'hourly_max': 100, 'payment_verified': True}
        """
        try:
            parsed_url = urlparse(url)
            query_params = parse_qs(parsed_url.query)
            
            logger.info(f"Parsing Upwork URL: {url}")
            logger.debug(f"Extracted query params: {query_params}")
            
            api_params = {}
            
            # Process each parameter
            for upwork_param, values in query_params.items():
                if not values or not values[0]:
                    continue
                
                value = values[0]  # Take first value
                
                # Handle each parameter type
                if upwork_param == 'q':
                    # Search query
                    api_params['search'] = value
                    
                elif upwork_param == 'payment_verified':
                    # Boolean: "1" means true
                    api_params['payment_verified'] = value == '1'
                    
                elif upwork_param in ['job_type', 't']:
                    # Job type mapping
                    api_params['job_type'] = UpworkURLParser._parse_job_type(value)
                    
                elif upwork_param == 'contractor_tier':
                    # Experience level
                    api_params['contractor_tier'] = UpworkURLParser._parse_contractor_tier(value)
                    
                elif upwork_param == 'job_status':
                    # Job status
                    api_params['status'] = UpworkURLParser._parse_status(value)
                    
                elif upwork_param in ['client_country', 'location']:
                    # Country code
                    api_params['country'] = value.upper()
                    
                elif upwork_param == 'skills':
                    # Skills (comma-separated)
                    api_params['skills'] = value
                    
                elif upwork_param == 'tags':
                    # Tags (comma-separated)
                    api_params['tags'] = value
                    
                elif upwork_param == 'amount':
                    # Fixed budget range: "500-1000" or "500-"
                    min_val, max_val = UpworkURLParser._parse_range(value)
                    if min_val is not None:
                        api_params['budget_min'] = min_val
                    if max_val is not None:
                        api_params['budget_max'] = max_val
                        
                elif upwork_param == 'hourly_rate':
                    # Hourly rate range: "25-100" or "50-"
                    min_val, max_val = UpworkURLParser._parse_range(value)
                    if min_val is not None:
                        api_params['hourly_min'] = min_val
                    if max_val is not None:
                        api_params['hourly_max'] = max_val
                        
                elif upwork_param == 'client_hires':
                    # Client hires range: "1-9" or "10-"
                    min_val, max_val = UpworkURLParser._parse_range(value)
                    if min_val is not None:
                        api_params['buyer.total_jobs_with_hires_min'] = int(min_val)
                    if max_val is not None:
                        api_params['buyer.total_jobs_with_hires_max'] = int(max_val)
                        
                elif upwork_param == 'client_total_spent':
                    # Client total spent range
                    min_val, max_val = UpworkURLParser._parse_range(value)
                    if min_val is not None:
                        api_params['buyer.total_spent_min'] = min_val
                    if max_val is not None:
                        api_params['buyer.total_spent_max'] = max_val
                        
                elif upwork_param == 'duration_v3':
                    # Duration label
                    api_params['duration_label'] = UpworkURLParser._parse_duration(value)
                    
                elif upwork_param == 'workload':
                    # Workload
                    api_params['workload'] = value
                    
                elif upwork_param == 'contract_to_hire':
                    # Contract to hire
                    api_params['is_contract_to_hire'] = value == '1' or value.lower() == 'true'
                    
                elif upwork_param == 'sort':
                    # Sort order
                    api_params['sort'] = UpworkURLParser._parse_sort(value)
                    
                elif upwork_param == 'category2_uid':
                    # Category
                    api_params['category'] = value
                    
                elif upwork_param == 'subcategory2_uid':
                    # Category group
                    api_params['category_group'] = value
            
            logger.info(f"Parsed {len(api_params)} parameters for Go API")
            logger.debug(f"Go API params: {api_params}")
            
            return api_params
            
        except Exception as e:
            logger.error(f"Error parsing Upwork URL: {e}")
            raise ValueError(f"Failed to parse Upwork URL: {e}")
    
    @staticmethod
    def _parse_range(range_str: str) -> tuple[Optional[float], Optional[float]]:
        """
        Parse a range string like "500-1000" or "25-" into min/max values.
        
        Args:
            range_str: Range string (e.g., "500-1000", "25-", "-1000")
            
        Returns:
            Tuple of (min_value, max_value), either can be None
        """
        if not range_str or '-' not in range_str:
            return None, None
        
        parts = range_str.split('-', 1)
        
        min_val = None
        max_val = None
        
        try:
            if parts[0].strip():
                min_val = float(parts[0].strip())
        except (ValueError, IndexError):
            pass
        
        try:
            if len(parts) > 1 and parts[1].strip():
                max_val = float(parts[1].strip())
        except (ValueError, IndexError):
            pass
        
        return min_val, max_val
    
    @staticmethod
    def _parse_job_type(value: str) -> str:
        """
        Parse job type from Upwork URL param to Go API format.
        
        Upwork uses: 't' parameter with values like "0" (hourly), "1" (fixed), "0,1" (any)
        Or 'job_type' parameter with values like "hourly", "fixed"
        
        Args:
            value: Job type value from URL
            
        Returns:
            Go API compatible job type string
        """
        value_lower = value.lower().strip()
        
        # Numeric codes
        if value_lower == '0':
            return 'hourly'
        elif value_lower == '1':
            return 'fixed-price'
        elif value_lower in ['0,1', '1,0']:
            return ''  # Any type
        
        # String values
        if 'hourly' in value_lower:
            return 'hourly'
        elif 'fixed' in value_lower or 'price' in value_lower:
            return 'fixed-price'
        
        return ''
    
    @staticmethod
    def _parse_contractor_tier(value: str) -> str:
        """
        Parse contractor tier from Upwork URL to Go API format.
        
        Upwork uses: "1" (entry), "2" (intermediate), "3" (expert), "1,2" (entry+intermediate)
        
        Args:
            value: Contractor tier value from URL
            
        Returns:
            Go API compatible contractor tier string
        """
        value_lower = value.lower().strip()
        
        # If multiple tiers, take the first one
        if ',' in value_lower:
            value_lower = value_lower.split(',')[0]
        
        tier_map = {
            '1': 'entry',
            '2': 'intermediate',
            '3': 'expert',
            'entry': 'entry',
            'intermediate': 'intermediate',
            'expert': 'expert',
        }
        
        return tier_map.get(value_lower, '')
    
    @staticmethod
    def _parse_status(value: str) -> str:
        """Parse job status."""
        value_lower = value.lower().strip()
        
        status_map = {
            'open': 'open',
            '1': 'open',
            'closed': 'closed',
            '2': 'closed',
        }
        
        return status_map.get(value_lower, '')
    
    @staticmethod
    def _parse_duration(value: str) -> str:
        """Parse duration label."""
        duration_map = {
            'week': 'Less than 1 month',
            'month': '1 to 3 months',
            'semester': '3 to 6 months',
            'ongoing': 'More than 6 months',
        }
        
        return duration_map.get(value.lower(), value)
    
    @staticmethod
    def _parse_sort(value: str) -> str:
        """
        Parse sort order from Upwork URL to Go API format.
        
        Upwork uses: 'recency', 'relevance', 'client_rating', 'duration', etc.
        Go API uses: 'publish_time_desc', 'publish_time_asc', 'last_visited_desc', etc.
        
        Args:
            value: Sort value from URL
            
        Returns:
            Go API compatible sort string
        """
        sort_map = {
            'recency': 'publish_time_desc',
            'relevance': 'publish_time_desc',  # Default to recency
            'client_rating': 'last_visited_desc',
            'duration': 'publish_time_desc',
        }
        
        return sort_map.get(value.lower(), 'publish_time_desc')


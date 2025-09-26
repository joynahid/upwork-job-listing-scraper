import logging
from datetime import datetime

from pydantic import BaseModel, Field

from .postgres_backend import PostgreSQLJobTracker, create_postgres_tracker

logger = logging.getLogger(__name__)


class FailedJobExtraction(BaseModel):
    """Model to track failed job extractions."""
    url: str
    error_message: str
    attempted_at: datetime = Field(default_factory=datetime.now)
    raw_data: dict | None = None  # Store whatever data was extracted before failure


class JobData(BaseModel):
    job_id: str | None = None
    title: str
    description: str
    budget: str | None = None
    hourly_rate: str | None = None
    skills: str | None = None
    posted_date: str
    proposals_count: str | None = None
    last_viewed_by_client: str | None = None
    hires: str | None = None
    interviewing: str | None = None
    invites_sent: str | None = None
    unanswered_invites: str | None = None
    experience_level: str
    job_type: str | None = None
    duration: str | None = None
    project_type: str | None = None  
    work_hours: str | None = None
    skills: list[str] | None = None
    client_location: str | None = None
    client_local_time: str | None = None
    client_industry: str | None = None
    client_company_size: str | None = None
    member_since: str
    total_spent: str | None = None
    total_hires: str | None = None
    total_active: str | None = None
    total_client_hours: str | None = None
    url: str | None = None
    last_visited_at: datetime = Field(default_factory=datetime.now)

    def set_url(self, url: str) -> None:
        self.url = url
        self.infer_job_id()

    def infer_job_id(self) -> str:
        assert self.url is not None

        # Remove query parameters first
        url_without_params = self.url.split("?")[0]
        
        # Split by "/" and get the last non-empty part
        parts = [part for part in url_without_params.split("/") if part]
        
        if parts:
            # The job ID is typically the last part, which looks like a tilde followed by numbers
            last_part = parts[-1]
            
            # If the last part starts with ~, it's likely the job ID
            if last_part.startswith("~"):
                self.job_id = last_part
            else:
                # Fallback to old logic
                self.job_id = last_part
        else:
            self.job_id = "unknown"
        
        return self.job_id


class RealtimeJobDatabase:
    """PostgreSQL-based job database for persistent job tracking."""

    def __init__(self, connection_string: str | None = None, stale_threshold_seconds: int = 10):
        """Initialize RealtimeJobDatabase with PostgreSQL backend.

        Args:
            connection_string: PostgreSQL connection string (if None, uses environment variables)
            stale_threshold_seconds: How old a job entry should be to be considered stale
        """
        self.stale_threshold_seconds = stale_threshold_seconds

        # Initialize PostgreSQL backend - this is our primary and only storage
        self.postgres_tracker: PostgreSQLJobTracker = create_postgres_tracker(
            connection_string, stale_threshold_seconds
        )

        logger.info(
            "RealtimeJobDatabase initialized with PostgreSQL backend",
        )

    def should_process(self, job_id: str) -> bool:
        """Check if a job should be processed using PostgreSQL.

        Args:
            job_id: Unique job identifier

        Returns:
            True if the job should be processed, False otherwise
        """
        return self.postgres_tracker.should_process(job_id)

    def do_seen(self, job_data: JobData) -> None:
        """Mark a job as seen/processed using PostgreSQL.

        Args:
            job_data: Job data to mark as seen
        """
        if job_data.job_id is None:
            raise ValueError("job_data must have a job_id")
            
        job_dict = job_data.model_dump()
        self.postgres_tracker.do_seen(job_dict)

    def get_job_history(self, job_id: str, limit: int = 10) -> list[dict]:
        """Get processing history for a job using PostgreSQL.

        Args:
            job_id: Unique job identifier
            limit: Maximum number of history entries to return

        Returns:
            List of job data dictionaries in reverse chronological order
        """
        return self.postgres_tracker.get_job_history(job_id, limit)

    def get_stats(self) -> dict:
        """Get PostgreSQL database statistics.

        Returns:
            Dictionary with database statistics
        """
        return self.postgres_tracker.get_stats()

    def record_failed_extraction(self, failed_job: FailedJobExtraction) -> None:
        """Record a failed job extraction attempt.
        
        Args:
            failed_job: FailedJobExtraction instance with failure details
        """
        self.postgres_tracker.record_failed_job(
            job_url=failed_job.url,
            error_message=failed_job.error_message,
            raw_data=failed_job.raw_data
        )

    def get_failed_extractions(self, limit: int = 50) -> list[dict]:
        """Get recent failed job extraction attempts.
        
        Args:
            limit: Maximum number of failed jobs to return
            
        Returns:
            List of failed job dictionaries, most recent first
        """
        return self.postgres_tracker.get_failed_jobs(limit)

    def cleanup_failed_extractions(self, days_to_keep: int = 30) -> int:
        """Clean up old failed job entries.
        
        Args:
            days_to_keep: Number of days of failed job history to keep
            
        Returns:
            Number of entries deleted
        """
        return self.postgres_tracker.cleanup_failed_jobs(days_to_keep)

    def cleanup_old_history(self, days_to_keep: int = 7) -> int:
        """Clean up old history entries to save space using PostgreSQL.

        Args:
            days_to_keep: Number of days of history to keep

        Returns:
            Number of entries deleted
        """
        return self.postgres_tracker.cleanup_old_history(days_to_keep)

    def close(self) -> None:
        """Close PostgreSQL connection with error handling."""
        try:
            if hasattr(self, 'postgres_tracker') and self.postgres_tracker:
                self.postgres_tracker.close()
                logger.info("RealtimeJobDatabase closed successfully")
        except Exception as e:
            logger.error("Error closing RealtimeJobDatabase: %s", e)

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit with error handling."""
        try:
            self.close()
        except Exception as e:
            logger.error("Error in RealtimeJobDatabase context manager exit: %s", e)
            # Don't suppress the original exception
            return False

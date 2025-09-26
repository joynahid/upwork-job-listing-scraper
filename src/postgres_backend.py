"""PostgreSQL backend for persistent job tracking with real-time updates."""

import json
import logging
import os
import uuid
from datetime import datetime, timedelta
from typing import Any, Dict, Optional

from sqlalchemy import Column, String, DateTime, Text, create_engine, desc
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.dialects.postgresql import UUID

logger = logging.getLogger(__name__)

Base = declarative_base()


class JobEntry(Base):
    """SQLAlchemy model for job entries."""

    __tablename__ = "job_entries"

    job_id = Column(String, primary_key=True)
    entry_type = Column(String, primary_key=True)  # 'latest' or 'history'
    timestamp = Column(String, primary_key=True)  # For history entries
    data = Column(Text)  # JSON data
    last_visited_at = Column(DateTime)


class FailedJobEntry(Base):
    """SQLAlchemy model for failed job extraction entries."""

    __tablename__ = "failed_job_entries"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    job_url = Column(String, nullable=False)
    job_id = Column(String, nullable=True)  # Extracted from URL if possible
    error_message = Column(Text, nullable=False)
    attempted_at = Column(DateTime, nullable=False)
    raw_data = Column(Text, nullable=True)  # JSON data of whatever was extracted


class PostgreSQLJobTracker:
    """PostgreSQL-based job tracker for real-time updates."""

    def __init__(self, connection_string: str, stale_threshold_seconds: int = 10):
        """Initialize PostgreSQL job tracker.

        Args:
            connection_string: PostgreSQL connection string
            stale_threshold_seconds: How old a job entry should be to be considered stale
        """
        self.connection_string = connection_string
        self.stale_threshold = timedelta(seconds=stale_threshold_seconds)
        self._initialize_db()

    def _initialize_db(self) -> None:
        """Initialize PostgreSQL database with proper setup for real-time updates."""
        try:
            # Create SQLAlchemy engine with optimized settings for real-time updates
            self.engine = create_engine(
                self.connection_string,
                echo=False,
                pool_pre_ping=True,
                pool_recycle=3600,  # Recycle connections every hour
                pool_size=10,
                max_overflow=20,
                connect_args={
                    "connect_timeout": 10,
                    "application_name": "upwork_job_scraper"
                }
            )

            # Create tables
            Base.metadata.create_all(self.engine)

            # Create session factory with autocommit for real-time updates
            self.SessionLocal = sessionmaker(autocommit=False, autoflush=True, bind=self.engine)

            logger.info("PostgreSQL database initialized successfully")

        except Exception as exc:
            logger.error("Failed to initialize PostgreSQL database: %s", exc)
            raise

    def should_process(self, job_id: str) -> bool:
        """Check if a job should be processed based on staleness.

        Args:
            job_id: Unique job identifier

        Returns:
            True if the job should be processed, False otherwise
        """
        try:
            with self.SessionLocal() as session:
                # Check for latest entry
                latest_entry = (
                    session.query(JobEntry)
                    .filter_by(job_id=job_id, entry_type="latest", timestamp="")
                    .first()
                )

                if not latest_entry:
                    logger.debug("Job %s not found in database - should process", job_id)
                    return True

                # Check if entry is stale
                time_diff = datetime.now() - latest_entry.last_visited_at
                is_stale = time_diff > self.stale_threshold

                if is_stale:
                    logger.debug(
                        "Job %s is stale (last seen %s ago) - should process",
                        job_id,
                        time_diff,
                    )
                else:
                    logger.debug(
                        "Job %s is fresh (last seen %s ago) - skip processing",
                        job_id,
                        time_diff,
                    )

                return is_stale

        except Exception as exc:
            logger.error("Error checking if job %s should be processed: %s", job_id, exc)
            return True  # Default to processing on error

    def do_seen(self, job_data: Dict[str, Any]) -> None:
        """Mark a job as seen/processed with immediate commit.

        Args:
            job_data: Dictionary containing job information including job_id
        """
        job_id = job_data.get("job_id")
        if not job_id:
            logger.error("Job data missing job_id, cannot mark as seen")
            return

        logger.info("💾 Marking job %s as SEEN in PostgreSQL database", job_id[:20] + "..." if len(job_id) > 20 else job_id)

        try:
            # Add timestamp if not present and ensure it's serializable
            if "last_visited_at" not in job_data:
                job_data["last_visited_at"] = datetime.now().isoformat()
            elif isinstance(job_data["last_visited_at"], datetime):
                job_data["last_visited_at"] = job_data["last_visited_at"].isoformat()

            # Serialize data once
            serialized_data = json.dumps(job_data, default=str)

            with self.SessionLocal() as session:
                # Store/update as latest entry
                latest_entry = (
                    session.query(JobEntry)
                    .filter_by(
                        job_id=job_id,
                        entry_type="latest",
                        timestamp="",  # Latest entries have empty timestamp
                    )
                    .first()
                )

                if latest_entry:
                    latest_entry.data = serialized_data
                    latest_entry.last_visited_at = datetime.now()
                else:
                    latest_entry = JobEntry(
                        job_id=job_id,
                        entry_type="latest",
                        timestamp="",
                        data=serialized_data,
                        last_visited_at=datetime.now(),
                    )
                    session.add(latest_entry)

                # Also store with timestamp for history tracking
                timestamp = datetime.now().isoformat().replace(":", "-").replace(".", "-")
                history_entry = JobEntry(
                    job_id=job_id,
                    entry_type="history",
                    timestamp=timestamp,
                    data=serialized_data,
                    last_visited_at=datetime.now(),
                )
                session.add(history_entry)

                # Commit immediately for real-time updates
                session.commit()
                
                logger.info("✅ Job %s committed to PostgreSQL in real-time", job_id[:20] + "..." if len(job_id) > 20 else job_id)

        except Exception as exc:
            logger.error("Error marking job %s as seen: %s", job_id, exc)

    def get_job_history(self, job_id: str, limit: int = 10) -> list[Dict[str, Any]]:
        """Get processing history for a job.

        Args:
            job_id: Unique job identifier
            limit: Maximum number of history entries to return

        Returns:
            List of job data dictionaries in reverse chronological order
        """
        try:
            with self.SessionLocal() as session:
                entries = (
                    session.query(JobEntry.data)
                    .filter_by(job_id=job_id, entry_type="history")
                    .order_by(desc(JobEntry.last_visited_at))
                    .limit(limit)
                    .all()
                )

                history = []
                for (data_str,) in entries:
                    try:
                        job_data = json.loads(data_str)
                        history.append(job_data)
                    except json.JSONDecodeError as e:
                        logger.warning("Failed to parse job data for %s: %s", job_id, e)

                return history

        except Exception as exc:
            logger.error("Error getting job history for %s: %s", job_id, exc)
            return []

    def get_stats(self) -> Dict[str, Any]:
        """Get database statistics.

        Returns:
            Dictionary with database statistics
        """
        try:
            with self.SessionLocal() as session:
                # Count latest entries (unique jobs)
                unique_jobs = session.query(JobEntry).filter_by(entry_type="latest").count()

                # Count total history entries
                total_history = session.query(JobEntry).filter_by(entry_type="history").count()

                # Count failed jobs
                failed_jobs = session.query(FailedJobEntry).count()

                # Get recent activity (last 24 hours)
                yesterday = datetime.now() - timedelta(days=1)
                recent_jobs = (
                    session.query(JobEntry)
                    .filter_by(entry_type="latest")
                    .filter(JobEntry.last_visited_at >= yesterday)
                    .count()
                )

                return {
                    "unique_jobs": unique_jobs,
                    "total_history_entries": total_history,
                    "failed_jobs": failed_jobs,
                    "recent_jobs_24h": recent_jobs,
                    "database_type": "postgresql",
                    "last_updated": datetime.now().isoformat(),
                }

        except Exception as exc:
            logger.error("Error getting database stats: %s", exc)
            return {"error": str(exc)}

    def cleanup_old_history(self, days_to_keep: int = 7) -> int:
        """Clean up old history entries to save space.

        Args:
            days_to_keep: Number of days of history to keep

        Returns:
            Number of entries deleted
        """
        try:
            cutoff_date = datetime.now() - timedelta(days=days_to_keep)

            with self.SessionLocal() as session:
                # Delete old history entries
                deleted_count = (
                    session.query(JobEntry)
                    .filter(
                        JobEntry.entry_type == "history", 
                        JobEntry.last_visited_at < cutoff_date
                    )
                    .delete()
                )

                session.commit()

                logger.info("Cleaned up %d old history entries", deleted_count)
                return deleted_count

        except Exception as exc:
            logger.error("Error cleaning up old history: %s", exc)
            return 0

    def record_failed_job(self, job_url: str, error_message: str, raw_data: Optional[Dict[str, Any]] = None) -> None:
        """Record a failed job extraction attempt.
        
        Args:
            job_url: The URL of the job that failed to extract
            error_message: Description of the error
            raw_data: Any partial data that was extracted before failure
        """
        # Extract job_id from URL if possible
        job_id = None
        try:
            if job_url:
                # Remove query parameters first
                url_without_params = job_url.split("?")[0]
                
                # Split and get last part
                parts = [part for part in url_without_params.split("/") if part]
                if parts:
                    job_id = parts[-1]
        except Exception:
            pass
            
        try:
            with self.SessionLocal() as session:
                failed_entry = FailedJobEntry(
                    job_url=job_url,
                    job_id=job_id,
                    error_message=error_message,
                    attempted_at=datetime.now(),
                    raw_data=json.dumps(raw_data) if raw_data else None
                )
                
                session.add(failed_entry)
                
                # Commit immediately for real-time updates
                session.commit()
                
                logger.info("✅ Failed job extraction recorded in real-time for URL: %s", job_url)
            
        except Exception as exc:
            logger.error("Error recording failed job: %s", exc)

    def get_failed_jobs(self, limit: int = 50) -> list[Dict[str, Any]]:
        """Get recent failed job extraction attempts.
        
        Args:
            limit: Maximum number of failed jobs to return
            
        Returns:
            List of failed job dictionaries, most recent first
        """
        try:
            with self.SessionLocal() as session:
                failed_entries = (
                    session.query(FailedJobEntry)
                    .order_by(desc(FailedJobEntry.attempted_at))
                    .limit(limit)
                    .all()
                )
                
                failed_jobs = []
                for entry in failed_entries:
                    failed_job = {
                        "id": str(entry.id),
                        "job_url": entry.job_url,
                        "job_id": entry.job_id,
                        "error_message": entry.error_message,
                        "attempted_at": entry.attempted_at.isoformat(),
                        "raw_data": json.loads(entry.raw_data) if entry.raw_data else None
                    }
                    failed_jobs.append(failed_job)
                
                return failed_jobs
            
        except Exception as exc:
            logger.error("Error getting failed jobs: %s", exc)
            return []

    def cleanup_failed_jobs(self, days_to_keep: int = 30) -> int:
        """Clean up old failed job entries.
        
        Args:
            days_to_keep: Number of days of failed job history to keep
            
        Returns:
            Number of entries deleted
        """
        try:
            cutoff_date = datetime.now() - timedelta(days=days_to_keep)
            
            with self.SessionLocal() as session:
                deleted_count = (
                    session.query(FailedJobEntry)
                    .filter(FailedJobEntry.attempted_at < cutoff_date)
                    .delete()
                )
                
                session.commit()
                
                logger.info("Cleaned up %d old failed job entries", deleted_count)
                return deleted_count
            
        except Exception as exc:
            logger.error("Error cleaning up failed jobs: %s", exc)
            return 0

    def close(self) -> None:
        """Close the database connection with comprehensive error handling."""
        try:
            # Close any open sessions first
            if hasattr(self, "SessionLocal") and self.SessionLocal:
                try:
                    # Close any remaining sessions
                    self.SessionLocal.close_all()
                except Exception as session_exc:
                    logger.debug("Error closing sessions: %s", session_exc)
            
            # Dispose of the engine
            if hasattr(self, "engine") and self.engine:
                self.engine.dispose()
                logger.info("PostgreSQL database connection closed")
                
        except Exception as exc:
            logger.error("Error closing PostgreSQL database: %s", exc)
        finally:
            # Ensure attributes are cleared even if cleanup fails
            if hasattr(self, "engine"):
                self.engine = None
            if hasattr(self, "SessionLocal"):
                self.SessionLocal = None

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit with error handling."""
        try:
            self.close()
        except Exception as e:
            logger.error("Error in PostgreSQLJobTracker context manager exit: %s", e)
            # Don't suppress the original exception
            return False


def create_postgres_tracker(
    connection_string: Optional[str] = None, stale_threshold_seconds: int = 10
) -> PostgreSQLJobTracker:
    """Create a PostgreSQL job tracker instance.

    Args:
        connection_string: PostgreSQL connection string. If None, uses environment variables
        stale_threshold_seconds: How old a job entry should be to be considered stale

    Returns:
        Initialized PostgreSQLJobTracker instance
    """
    if connection_string is None:
        # Build connection string from environment variables
        host = os.getenv("POSTGRES_HOST", "localhost")
        port = os.getenv("POSTGRES_PORT", "5432")
        database = os.getenv("POSTGRES_DB", "upwork_jobs")
        username = os.getenv("POSTGRES_USER", "postgres")
        password = os.getenv("POSTGRES_PASSWORD", "postgres")
        
        connection_string = f"postgresql://{username}:{password}@{host}:{port}/{database}"

    return PostgreSQLJobTracker(connection_string, stale_threshold_seconds)

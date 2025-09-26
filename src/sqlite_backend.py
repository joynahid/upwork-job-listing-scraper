"""SQLite backend for persistent job tracking (RocksDB interface wrapper)."""

import json
import logging
import os
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, Optional

from sqlalchemy import Column, String, DateTime, Text, create_engine, desc
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

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

    id = Column(String, primary_key=True)  # UUID or auto-increment
    job_url = Column(String, nullable=False)
    job_id = Column(String, nullable=True)  # Extracted from URL if possible
    error_message = Column(Text, nullable=False)
    attempted_at = Column(DateTime, nullable=False)
    raw_data = Column(Text, nullable=True)  # JSON data of whatever was extracted


class SQLiteJobTracker:
    """SQLite-based job tracker that mimics RocksDB interface."""

    def __init__(self, db_path: str | Path, stale_threshold_seconds: int = 10):
        """Initialize SQLite job tracker.

        Args:
            db_path: Path to SQLite database file/directory
            stale_threshold_seconds: How old a job entry should be to be considered stale
        """
        self.db_path = Path(db_path)
        self.stale_threshold = timedelta(seconds=stale_threshold_seconds)

        # Ensure the directory exists
        if self.db_path.suffix != ".db":
            # If it's a directory, create the db file inside it
            self.db_path.mkdir(parents=True, exist_ok=True)
            self.db_file = self.db_path / "jobs.db"
        else:
            # If it's a file path, ensure parent directory exists
            self.db_path.parent.mkdir(parents=True, exist_ok=True)
            self.db_file = self.db_path

        self._initialize_db()

    def _initialize_db(self) -> None:
        """Initialize SQLite database with proper setup."""
        try:
            # Create SQLAlchemy engine
            self.engine = create_engine(f"sqlite:///{self.db_file}", echo=False)

            # Create tables
            Base.metadata.create_all(self.engine)

            # Create session factory
            self.SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=self.engine)

            logger.info("SQLite database initialized successfully at %s", self.db_file)

        except Exception as exc:
            logger.error("Failed to initialize SQLite database at %s: %s", self.db_file, exc)
            raise

    def should_process(self, job_id: str) -> bool:
        """Check if a job should be processed.

        A job should be processed if:
        1. It hasn't been seen before, OR
        2. The last entry for this job is older than the stale threshold

        Args:
            job_id: Unique job identifier

        Returns:
            True if the job should be processed, False otherwise
        """
        try:
            with self.SessionLocal() as session:
                # Get the latest entry for this job
                entry = (
                    session.query(JobEntry).filter_by(job_id=job_id, entry_type="latest").first()
                )

                if entry is None:
                    # Job not seen before
                    logger.info("🆕 Job %s not seen before, should process", job_id[:20] + "..." if len(job_id) > 20 else job_id)
                    return True

                # Parse the stored data to get last_visited_at
                try:
                    job_data = json.loads(entry.data)
                    last_visited_str = job_data.get("last_visited_at")

                    if not last_visited_str:
                        logger.warning("Job %s missing last_visited_at, treating as new", job_id)
                        return True

                    # Parse the datetime
                    last_visited = datetime.fromisoformat(last_visited_str.replace("Z", "+00:00"))
                    if last_visited.tzinfo:
                        last_visited = last_visited.replace(tzinfo=None)

                    # Check if it's stale
                    now = datetime.now()
                    is_stale = (now - last_visited) > self.stale_threshold
                    time_diff = now - last_visited

                    if is_stale:
                        logger.info(
                            "🕒 Job %s last visited %s ago, will process (threshold=%s)",
                            job_id[:20] + "..." if len(job_id) > 20 else job_id,
                            time_diff,
                            self.stale_threshold,
                        )
                    else:
                        logger.info(
                            "👀 Job %s already seen %s ago, skipping (not stale, threshold=%s)",
                            job_id[:20] + "..." if len(job_id) > 20 else job_id,
                            time_diff,
                            self.stale_threshold,
                        )

                    return is_stale

                except (json.JSONDecodeError, ValueError) as exc:
                    logger.error("Failed to parse job data for %s: %s", job_id, exc)
                    return True  # Treat as new if we can't parse

        except Exception as exc:
            logger.error("Error checking if job %s should be processed: %s", job_id, exc)
            return True  # Default to processing on error

    def do_seen(self, job_data: Dict[str, Any]) -> None:
        """Mark a job as seen/processed.

        Args:
            job_data: Dictionary containing job information including job_id
        """
        job_id = job_data.get("job_id")
        if not job_id:
            logger.error("Job data missing job_id, cannot mark as seen")
            return

        logger.info("💾 Marking job %s as SEEN in database", job_id[:20] + "..." if len(job_id) > 20 else job_id)

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

                session.commit()
                logger.debug("Marked job %s as seen", job_id)

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
                    .order_by(desc(JobEntry.timestamp))
                    .limit(limit)
                    .all()
                )

                history = []
                for (data,) in entries:  # Unpack tuple from query result
                    try:
                        job_data = json.loads(data)
                        history.append(job_data)
                    except json.JSONDecodeError as exc:
                        logger.warning("Failed to parse history entry for job %s: %s", job_id, exc)
                        continue

                return history

        except Exception as exc:
            logger.error("Error getting history for job %s: %s", job_id, exc)
            return []

    def get_stats(self) -> Dict[str, Any]:
        """Get database statistics.

        Returns:
            Dictionary with database statistics
        """
        try:
            with self.SessionLocal() as session:
                # Count total jobs (latest entries)
                job_count = session.query(JobEntry).filter_by(entry_type="latest").count()

                # Count total history entries
                history_count = session.query(JobEntry).filter_by(entry_type="history").count()

                stats = {
                    "total_jobs": job_count,
                    "total_history_entries": history_count,
                    "db_path": str(self.db_file),
                    "db_type": "sqlite",
                    "stale_threshold_seconds": self.stale_threshold.total_seconds(),
                }

                return stats

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
                        JobEntry.entry_type == "history", JobEntry.last_visited_at < cutoff_date
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
        import uuid
        
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
                    id=str(uuid.uuid4()),
                    job_url=job_url,
                    job_id=job_id,
                    error_message=error_message,
                    attempted_at=datetime.now(),
                    raw_data=json.dumps(raw_data) if raw_data else None
                )
                
                session.add(failed_entry)
                session.commit()
                
                logger.info("Recorded failed job extraction for URL: %s", job_url)
            
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
                failed_entries = session.query(FailedJobEntry)\
                    .order_by(desc(FailedJobEntry.attempted_at))\
                    .limit(limit)\
                    .all()
                
                result = []
                for entry in failed_entries:
                    failed_data = {
                        "id": entry.id,
                        "job_url": entry.job_url,
                        "job_id": entry.job_id,
                        "error_message": entry.error_message,
                        "attempted_at": entry.attempted_at.isoformat(),
                        "raw_data": json.loads(entry.raw_data) if entry.raw_data else None
                    }
                    result.append(failed_data)
                
                return result
            
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
                deleted_count = session.query(FailedJobEntry)\
                    .filter(FailedJobEntry.attempted_at < cutoff_date)\
                    .delete()
                
                session.commit()
                
                logger.info("Cleaned up %d old failed job entries", deleted_count)
                return deleted_count
            
        except Exception as exc:
            logger.error("Error cleaning up failed jobs: %s", exc)
            return 0

    def close(self) -> None:
        """Close the database connection."""
        try:
            if hasattr(self, "engine"):
                self.engine.dispose()
                logger.info("SQLite database connection closed")
        except Exception as exc:
            logger.error("Error closing SQLite database: %s", exc)

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()


def create_sqlite_tracker(
    db_path: Optional[str] = None, stale_threshold_seconds: int = 10
) -> SQLiteJobTracker:
    """Factory function to create SQLite job tracker.

    Args:
        db_path: Path to database directory/file. If None, uses environment variable
                 ROCKSDB_DATA_DIR or defaults to ./sqlite_data
        stale_threshold_seconds: How old a job entry should be to be considered stale

    Returns:
        Initialized SQLiteJobTracker instance
    """
    if db_path is None:
        # Use the same env var name for compatibility
        db_path = os.getenv("ROCKSDB_DATA_DIR", "./sqlite_data")

    return SQLiteJobTracker(db_path, stale_threshold_seconds)

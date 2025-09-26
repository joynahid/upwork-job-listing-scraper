"""RocksDB backend for persistent job tracking."""

import json
import logging
import os
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any

import rocksdb

logger = logging.getLogger(__name__)


class RocksDBJobTracker:
    """RocksDB-based job tracker for persistent storage of job processing state."""

    def __init__(self, db_path: str | Path, stale_threshold_seconds: int = 10):
        """Initialize RocksDB job tracker.

        Args:
            db_path: Path to RocksDB database directory
            stale_threshold_seconds: How old a job entry should be to be considered stale
        """
        self.db_path = Path(db_path)
        self.stale_threshold = timedelta(seconds=stale_threshold_seconds)
        self._db: rocksdb.DB | None = None
        self._opts: rocksdb.Options | None = None

        self._initialize_db()

    def _initialize_db(self) -> None:
        """Initialize RocksDB database with proper options."""
        try:
            # Create database directory if it doesn't exist
            self.db_path.mkdir(parents=True, exist_ok=True)

            # Configure RocksDB options
            self._opts = rocksdb.Options()
            self._opts.create_if_missing = True
            self._opts.error_if_exists = False

            # Performance optimizations
            self._opts.write_buffer_size = 64 * 1024 * 1024  # 64MB
            self._opts.max_write_buffer_number = 3
            self._opts.target_file_size_base = 64 * 1024 * 1024  # 64MB

            # Compression
            self._opts.compression = rocksdb.CompressionType.snappy_compression

            # Open the database
            self._db = rocksdb.DB(str(self.db_path), self._opts)

            logger.info("RocksDB initialized successfully at %s", self.db_path)

        except Exception as exc:
            logger.error("Failed to initialize RocksDB at %s: %s", self.db_path, exc)
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
        if not self._db:
            logger.error("RocksDB not initialized")
            return False

        try:
            # Get the latest entry for this job
            key = f"job:{job_id}:latest"
            value = self._db.get(key.encode())

            if value is None:
                # Job not seen before
                logger.debug("Job %s not seen before, should process", job_id)
                return True

            # Parse the stored data
            try:
                job_data = json.loads(value.decode())
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

                logger.debug(
                    "Job %s last visited at %s, is_stale=%s (threshold=%s)",
                    job_id, last_visited, is_stale, self.stale_threshold
                )

                return is_stale

            except (json.JSONDecodeError, ValueError) as exc:
                logger.error("Failed to parse job data for %s: %s", job_id, exc)
                return True  # Treat as new if we can't parse

        except Exception as exc:
            logger.error("Error checking if job %s should be processed: %s", job_id, exc)
            return True  # Default to processing on error

    def do_seen(self, job_data: dict[str, Any]) -> None:
        """Mark a job as seen/processed.

        Args:
            job_data: Dictionary containing job information including job_id
        """
        if not self._db:
            logger.error("RocksDB not initialized")
            return

        job_id = job_data.get("job_id")
        if not job_id:
            logger.error("Job data missing job_id, cannot mark as seen")
            return

        try:
            # Add timestamp if not present
            if "last_visited_at" not in job_data:
                job_data["last_visited_at"] = datetime.now().isoformat()

            # Store as latest entry
            latest_key = f"job:{job_id}:latest"
            self._db.put(latest_key.encode(), json.dumps(job_data).encode())

            # Also store with timestamp for history tracking
            timestamp = datetime.now().isoformat().replace(":", "-").replace(".", "-")
            history_key = f"job:{job_id}:history:{timestamp}"
            self._db.put(history_key.encode(), json.dumps(job_data).encode())

            logger.debug("Marked job %s as seen", job_id)

        except Exception as exc:
            logger.error("Error marking job %s as seen: %s", job_id, exc)

    def get_job_history(self, job_id: str, limit: int = 10) -> list[dict[str, Any]]:
        """Get processing history for a job.

        Args:
            job_id: Unique job identifier
            limit: Maximum number of history entries to return

        Returns:
            List of job data dictionaries in reverse chronological order
        """
        if not self._db:
            logger.error("RocksDB not initialized")
            return []

        try:
            history = []
            prefix = f"job:{job_id}:history:".encode()

            # Use iterator to get all history entries
            it = self._db.iteritems()
            it.seek(prefix)

            for key, value in it:
                if not key.startswith(prefix):
                    break

                try:
                    job_data = json.loads(value.decode())
                    history.append(job_data)

                    if len(history) >= limit:
                        break

                except json.JSONDecodeError as exc:
                    logger.warning("Failed to parse history entry %s: %s", key, exc)
                    continue

            # Sort by timestamp (most recent first)
            history.sort(
                key=lambda x: x.get("last_visited_at", ""),
                reverse=True
            )

            return history

        except Exception as exc:
            logger.error("Error getting history for job %s: %s", job_id, exc)
            return []

    def get_stats(self) -> dict[str, Any]:
        """Get database statistics.

        Returns:
            Dictionary with database statistics
        """
        if not self._db:
            return {"error": "RocksDB not initialized"}

        try:
            stats = {}

            # Count total jobs
            job_count = 0
            history_count = 0

            it = self._db.iteritems()
            for key, _ in it:
                key_str = key.decode()
                if ":latest" in key_str:
                    job_count += 1
                elif ":history:" in key_str:
                    history_count += 1

            stats["total_jobs"] = job_count
            stats["total_history_entries"] = history_count
            stats["db_path"] = str(self.db_path)

            # Get RocksDB internal stats if available
            try:
                db_stats = self._db.get_property(b"rocksdb.stats").decode()
                stats["rocksdb_stats"] = db_stats
            except Exception:
                pass  # Stats not available in all RocksDB versions

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
        if not self._db:
            logger.error("RocksDB not initialized")
            return 0

        try:
            cutoff_date = datetime.now() - timedelta(days=days_to_keep)
            cutoff_str = cutoff_date.isoformat()

            deleted_count = 0
            keys_to_delete = []

            # Find old history entries
            it = self._db.iteritems()
            for key, value in it:
                key_str = key.decode()
                if ":history:" not in key_str:
                    continue

                try:
                    job_data = json.loads(value.decode())
                    last_visited = job_data.get("last_visited_at", "")

                    if last_visited < cutoff_str:
                        keys_to_delete.append(key)

                except json.JSONDecodeError:
                    # Delete corrupted entries too
                    keys_to_delete.append(key)

            # Delete old entries
            for key in keys_to_delete:
                self._db.delete(key)
                deleted_count += 1

            logger.info("Cleaned up %d old history entries", deleted_count)
            return deleted_count

        except Exception as exc:
            logger.error("Error cleaning up old history: %s", exc)
            return 0

    def close(self) -> None:
        """Close the database connection."""
        if self._db:
            try:
                # RocksDB Python doesn't have an explicit close method
                # The database is automatically closed when the object is garbage collected
                self._db = None
                logger.info("RocksDB connection closed")
            except Exception as exc:
                logger.error("Error closing RocksDB: %s", exc)

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()


def create_rocksdb_tracker(
    db_path: str | None = None,
    stale_threshold_seconds: int = 10
) -> RocksDBJobTracker:
    """Factory function to create RocksDB job tracker.
    
    Args:
        db_path: Path to database directory. If None, uses environment variable
                 ROCKSDB_DATA_DIR or defaults to ./rocksdb_data
        stale_threshold_seconds: How old a job entry should be to be considered stale
        
    Returns:
        Initialized RocksDBJobTracker instance
    """
    if db_path is None:
        db_path = os.getenv("ROCKSDB_DATA_DIR", "./rocksdb_data")

    return RocksDBJobTracker(db_path, stale_threshold_seconds)

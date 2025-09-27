"""Firebase provider for direct Firebase Realtime Database access."""

import os
import logging
from typing import Optional

import firebase_admin
from firebase_admin import credentials, db, firestore_async
from google.cloud import firestore


logger = logging.getLogger(__name__)


class FirebaseProvider:
    """Simple Firebase Realtime Database provider."""

    def __init__(
        self,
        service_account_path: Optional[str] = None,
    ):
        """Initialize Firebase provider.

        Args:
            database_url: Firebase Realtime Database URL (if None, uses FIREBASE_DATABASE_URL env var)
            service_account_path: Path to service account JSON (if None, uses FIREBASE_SERVICE_ACCOUNT_PATH env var or default credentials)
        """
        self.service_account_path = service_account_path or os.getenv(
            "FIREBASE_SERVICE_ACCOUNT_PATH"
        )

        self._initialize_firebase()

    def _initialize_firebase(self) -> None:
        """Initialize Firebase Admin SDK."""
        try:
            # Check if Firebase app is already initialized
            try:
                app = firebase_admin.get_app()
                logger.info("Using existing Firebase app instance")
            except ValueError:
                # Initialize new Firebase app
                if self.service_account_path and os.path.exists(
                    self.service_account_path
                ):
                    cred = credentials.Certificate(self.service_account_path)
                    logger.info(
                        f"Using service account from: {self.service_account_path}"
                    )
                else:
                    # Use default credentials (GOOGLE_APPLICATION_CREDENTIALS)
                    cred = credentials.ApplicationDefault()
                    logger.info("Using default application credentials")

                app = firebase_admin.initialize_app(
                    cred
                )
                logger.info("Firebase app initialized successfully")


        except Exception as e:
            logger.error(f"Failed to initialize Firebase: {e}")
            raise

    @property
    def firestore(self) -> firestore.AsyncClient:
        """Get Firebase Firestore client."""
        return firestore_async.client()


def get_firebase() -> FirebaseProvider:
    """Factory function to get a configured Firebase provider.

    Returns:
        FirebaseProvider: Configured Firebase provider instance
    """
    return FirebaseProvider()


def get_firebase_with_config(
    service_account_path: Optional[str] = None
) -> FirebaseProvider:
    """Factory function to get a Firebase provider with specific configuration.

    Args:
        service_account_path: Path to service account JSON file

    Returns:
        FirebaseProvider: Configured Firebase provider instance
    """
    return FirebaseProvider(service_account_path)

"""Entry point for the Upwork Job Scraper Apify Actor."""

from .main import main
import asyncio

if __name__ == "__main__":
    asyncio.run(main())
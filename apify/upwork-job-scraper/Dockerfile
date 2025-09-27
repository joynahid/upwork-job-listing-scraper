# Use the official Apify Python base image
FROM apify/actor-python:3.11

# Copy requirements.txt first for better Docker layer caching
COPY requirements.txt ./

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the source code
COPY . ./

# Set the entry point to run the main.py file
CMD python3 src/main.py

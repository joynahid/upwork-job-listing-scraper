# Use Ubuntu LTS as base image - let uv handle Python
FROM ubuntu:22.04

# Set timezone and locale to avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Copy uv from the official Docker image (best practice)
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# Install system dependencies (including Python for uv)
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    curl \
    gnupg \
    ca-certificates \
    xvfb \
    x11vnc \
    fluxbox \
    dbus-x11 \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcairo-gobject2 \
    libdrm2 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libxss1 \
    libxtst6 \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js (required for Botasaurus)
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
 && apt-get install -y nodejs \
 && echo "Node.js version:" \
 && node --version \
 && echo "npm version:" \
 && npm --version \
 && npm config set cache /app/.npm \
 && npm install -g proxy-chain \
 && mkdir -p /app/.npm \
 && chown -R 1000:1000 /app/.npm

# Install Google Chrome (with dependency fix)
RUN apt-get update && apt-get install -y wget
RUN wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN apt-get update && apt-get install -y -f ./google-chrome-stable_current_amd64.deb

# Set working directory
WORKDIR /app

# Optimize uv settings for Docker builds
ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy
ENV UV_CACHE_DIR=/root/.cache/uv

# Copy project configuration files for dependency installation
COPY pyproject.toml ./
COPY uv.lock ./

# Install dependencies only (optimized for caching)
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-install-project --no-dev

# Copy the project source code
COPY . .

# Install the project itself (fast since deps are cached)
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev

# Compile Python code to check for syntax errors
RUN python3 -m compileall -q src/

# Create storage directory
RUN mkdir -p /app/storage

# Create runtime cache directory and change ownership of entire app directory
RUN mkdir -p /app/.cache/uv && chown -R 1000:1000 /app

# Set runtime UV environment variables for non-root user
ENV UV_CACHE_DIR=/app/.cache/uv
ENV UV_NO_CACHE=0

# Set npm environment variables for non-root user
ENV NPM_CONFIG_CACHE=/app/.npm
ENV NPM_CONFIG_PREFIX=/app/.npm-global

# Chrome/Browser environment variables for Docker
ENV DISPLAY=:99
ENV CHROME_BIN=/usr/bin/google-chrome
ENV CHROME_PATH=/usr/bin/google-chrome
ENV CHROMIUM_PATH=/usr/bin/google-chrome
ENV GOOGLE_CHROME_BIN=/usr/bin/google-chrome

# Create startup script for Chrome in Docker
RUN echo '#!/bin/bash\n\
# Start Xvfb for headless display\n\
Xvfb :99 -screen 0 1920x1080x24 -ac +extension GLX +render -noreset &\n\
export DISPLAY=:99\n\
\n\
# Wait for display to be ready\n\
sleep 2\n\
\n\
# Start the application\n\
UV_CACHE_DIR=/app/.cache/uv uv run python -m src\n\
' > /app/start.sh && chmod +x /app/start.sh

# Run the startup script
CMD ["/app/start.sh"]

# Apify GitHub Workflow Setup

This document explains how to configure the GitHub workflow for automatic deployment to Apify.

## Required GitHub Secrets

To enable automatic deployment, you need to add the following secret to your GitHub repository:

### APIFY_API_KEY

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `APIFY_API_KEY`
5. Value: `your-secret-api-key-here` (replace with your actual Apify API token)

## How to Get Your Apify API Token

1. Log in to your [Apify Console](https://console.apify.com/)
2. Go to **Settings** → **Integrations**
3. Find the **API tokens** section
4. Copy your existing token or create a new one
5. Use this token as the value for `APIFY_API_KEY` in GitHub secrets

## Workflow Triggers

The workflow will automatically run when:

- **Push to main branch**: Any changes to files in `apify/upwork-job-scraper/` directory
- **Pull request to main**: For testing deployments before merging
- **Manual trigger**: You can manually run the workflow from the Actions tab

## Workflow Steps

1. **Checkout code**: Downloads the repository code
2. **Setup Node.js**: Installs Node.js runtime for Apify CLI
3. **Install Apify CLI**: Installs the official Apify command-line tool
4. **Deploy**: Authenticates with Apify and pushes the actor

## Troubleshooting

### Common Issues

- **Authentication failed**: Check that your `APIFY_API_KEY` secret is correctly set
- **Actor not found**: Ensure the actor exists in your Apify account
- **Build failed**: Check the actor's Dockerfile and dependencies

### Checking Deployment Status

1. Go to **Actions** tab in your GitHub repository
2. Click on the latest workflow run
3. Check the logs for any errors or success messages
4. Verify deployment in your [Apify Console](https://console.apify.com/)

## Actor Configuration

The actor is configured with:
- **Name**: `live-upwork-job-scraper`
- **Environment Variables**:
  - `API_ENDPOINT`: https://upworkjobscraperapi.nahidhq.com
  - `API_KEY`: Uses the `@apiKey` reference from Apify secrets

Make sure your Apify account has the necessary API key configured for the actor to work properly.

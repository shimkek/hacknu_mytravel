# GitHub Actions Deployment Setup

This document explains how to set up the CI/CD pipeline for the BI Telegram Bot project.

## Required GitHub Secrets

You need to configure the following secrets in your GitHub repository:

### Docker Hub Credentials
- `DOCKER_USERNAME`: Your Docker Hub username 
- `DOCKER_PASSWORD`: Your Docker Hub password 

### VPS Connection Details
- `VPS_HOST`: Your VPS hostname or IP address (e.g., srv985126.hstgr.cloud)
- `VPS_USERNAME`: SSH username for your VPS (e.g., root or your user)
- `VPS_SSH_KEY`: Private SSH key for connecting to your VPS

### Telegram Bot Configuration
- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token

## How to Configure GitHub Secrets

1. Go to your GitHub repository
2. Click on **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret** for each secret above
4. Add the name and value for each secret

## SSH Key Setup

### Generate SSH Key (if you don't have one)
```bash
ssh-keygen -t rsa -b 4096 -C "github-actions@yourdomain.com"
```

### Copy Public Key to VPS
```bash
ssh-copy-id -i ~/.ssh/id_rsa.pub username@your-vps-host
```

### Add Private Key to GitHub Secrets
```bash
# Copy the private key content
cat ~/.ssh/id_rsa
```
Copy the entire output and paste it as the `VPS_SSH_KEY` secret in GitHub.

## VPS Prerequisites

Your VPS should have:
1. **Docker and Docker Compose installed**
2. **SSH access configured**
3. **Project directory created**: `~/bi_telegram_bot/`
4. **Database initialization files**: Copy the `init-db/` folder to your VPS

### Install Docker on VPS (Ubuntu/Debian)
```bash
# Update package index
sudo apt update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose-plugin

# Add your user to docker group
sudo usermod -aG docker $USER
```

### Prepare VPS Directory Structure
```bash
# Create project directory
mkdir -p ~/bi_telegram_bot

# Create necessary subdirectories
cd ~/bi_telegram_bot
mkdir -p photos init-db

# Copy database initialization files (do this manually or via scp)
# scp -r ./init-db/ username@your-vps:/home/username/bi_telegram_bot/
```

## Workflow Explanation

The deployment workflow consists of 4 main jobs:

### 1. **Test Job**
- Runs on every push and pull request
- Checks Python syntax and code quality
- Installs dependencies and runs linting
- Validates that both Python files can be compiled

### 2. **Build and Push Job**
- Runs only on pushes to `main` branch
- Builds Docker images for both services using a matrix strategy
- Pushes images to Docker Hub with multiple tags:
  - `latest` (for main branch)
  - `sha-<commit-hash>`
  - Branch name
- Uses Docker layer caching for faster builds
- Supports multi-platform builds (amd64, arm64)

### 3. **Deploy Job**
- Runs only after successful test and build jobs
- Connects to VPS via SSH
- Copies deployment files (docker-compose.yml, .env)
- Pulls latest Docker images
- Stops and recreates application containers (keeps database running)
- Verifies deployment success
- Cleans up old Docker images

### 4. **Test Local Job** (Pull Requests Only)
- Builds images locally for testing
- Validates that images can be built and started
- Runs basic import tests

## Docker Images

The workflow creates one Docker image:
- `yerassyl20036/bi-telegram-bot:latest` - The Telegram bot service

The deployment also includes:
- Minio S3-compatible storage service for photos (configured once, persistent across deployments)

## Deployment Process

1. **Trigger**: Push to `main` branch
2. **Test**: Code quality checks and syntax validation
3. **Build**: Create and push Docker images to Docker Hub
4. **Deploy**: 
   - Connect to VPS via SSH
   - Pull latest images
   - Stop old application containers
   - Start new containers with updated images
   - Verify deployment
   - Clean up old images

## Database Handling

- The database container is **NOT** recreated during deployment
- Only application containers (telegram-bot) are updated
- Minio and database containers persist across deployments
- Database and Minio initialization only happens on first deployment

## Testing the Deployment

### Local Testing
You can test the Docker builds locally:
```bash
# Build image locally
docker build -f Dockerfile.bot -t test-telegram-bot .

# Test that it starts
docker run --rm test-telegram-bot python -c "import telegram_bot; print('Bot OK')"
```

### Monitor Deployment
After pushing to main, you can:
1. Check the Actions tab in GitHub to see the workflow progress
2. SSH into your VPS to check container status:
```bash
cd ~/bi_telegram_bot
docker-compose ps
docker-compose logs telegram-bot
docker-compose logs minio
```

## Troubleshooting

### Common Issues

1. **SSH Connection Failed**
   - Verify VPS_HOST, VPS_USERNAME, and VPS_SSH_KEY secrets
   - Test SSH connection manually: `ssh username@host`

2. **Docker Login Failed**
   - Check DOCKER_USERNAME and DOCKER_PASSWORD secrets
   - Verify Docker Hub credentials

3. **Container Won't Start**
   - Check environment variables in .env file
   - Verify database is running: `docker-compose ps postgres`
   - Check logs: `docker-compose logs service-name`

4. **Permission Denied**
   - Ensure user is in docker group on VPS
   - Check file permissions in project directory

### Manual Deployment
If you need to deploy manually:
```bash
# On your VPS
cd ~/bi_telegram_bot
docker-compose pull
docker-compose up -d --no-deps telegram-bot
```

## Security Notes

- SSH keys are automatically cleaned up after deployment
- Environment variables are handled securely through GitHub Secrets
- Database credentials are not exposed in logs
- Docker Hub authentication is done securely

## Monitoring

After deployment, verify services are running:
```bash
# Check service status
curl http://your-vps/  # Minio health check (port 80)
curl http://your-vps:9001/  # Minio console access

# Check logs
docker-compose logs -f telegram-bot
docker-compose logs -f minio
```

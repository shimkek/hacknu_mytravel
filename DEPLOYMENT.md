# VPS Deployment Guide

## Quick Setup on VPS

### 1. Initial Setup (run once)
```bash
# Install Docker and Docker Compose if not already installed
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
# Logout and login again

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Clone the repository
git clone https://github.com/Yerassyl20036/hacknu_mytravel.git
cd hacknu_mytravel
```

### 2. Fix Repository Issues (if needed)

If you get "divergent branches" error, it means there are conflicts between master and main branches.

**On your local machine:**
```bash
# Fix the repository structure
./fix-repo.sh
```

**On VPS if you get branch errors:**
```bash
# Delete the directory and clone fresh
rm -rf hacknu_mytravel
git clone https://github.com/Yerassyl20036/hacknu_mytravel.git
cd hacknu_mytravel
```

### 3. Deploy/Update (run anytime)
```bash
# Simple one-command deployment
./deploy.sh
```

This script will:
- Automatically detect and switch to the correct branch (master or main)
- Pull latest changes from GitHub
- Stop old containers
- Build new containers locally
- Start all services
- Check everything is working
- Show you monitoring commands

### 3. Monitor
```bash
# View all logs
docker-compose logs -f

# View specific parser
docker-compose logs -f parser_2gis

# Connect to database
docker-compose exec postgres psql -U postgres -d mytravel_db
```

### 4. Database Queries
```sql
-- See all accommodations
SELECT * FROM accommodations;

-- See recent parser activity
SELECT * FROM parsing_logs ORDER BY started_at DESC LIMIT 10;

-- Count by source
SELECT source_website, COUNT(*) FROM accommodations GROUP BY source_website;
```

## Why This Approach?

- **No Docker Registry**: Builds locally on VPS
- **No CI/CD**: Simple manual deployment
- **Port 5433**: Won't conflict with existing PostgreSQL
- **Independent Services**: Each parser runs separately
- **Easy Updates**: Just run `./deploy.sh` again
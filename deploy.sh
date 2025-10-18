#!/bin/bash

# Simple VPS deployment script for MyTravel parser project
# This script pulls latest changes from git and rebuilds the containers

set -e  # Exit on any error

echo "🚀 Starting MyTravel deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if git is available
if ! command -v git &> /dev/null; then
    print_error "Git is not installed!"
    exit 1
fi

# Check if docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed!"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed!"
    exit 1
fi

print_status "Pulling latest changes from git..."
# Make sure we're on the right remote
git remote set-url origin https://github.com/Yerassyl20036/hacknu_mytravel.git

# Check which branch has the actual code and switch to it
if git ls-remote --heads origin master | grep -q "master"; then
    print_status "Found master branch, switching to master..."
    git fetch origin
    git checkout master 2>/dev/null || git checkout -b master origin/master
    git pull origin master --allow-unrelated-histories
else
    print_status "Using main branch..."
    git fetch origin
    git checkout main 2>/dev/null || git checkout -b main origin/main  
    git pull origin main --allow-unrelated-histories
fi

print_status "Stopping existing containers..."
docker-compose down

print_status "Checking if PostgreSQL is already running..."
if docker-compose ps postgres | grep -q "Up"; then
    print_warning "PostgreSQL container is running, will restart it"
elif netstat -tln | grep -q ":5433"; then
    print_warning "Something is already running on port 5433"
    print_warning "If this is your existing PostgreSQL, the deployment will use port 5433"
fi

print_status "Building and starting all services..."
docker-compose up --build -d

print_status "Waiting for services to start..."
sleep 10

print_status "Checking service status..."
docker-compose ps

print_status "Checking database connection..."
if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    print_status "✅ Database is ready"
else
    print_error "❌ Database is not ready"
    print_error "Check logs: docker-compose logs postgres"
    exit 1
fi

print_status "Checking parser services..."
for service in parser_2gis parser_google_maps parser_instagram parser_olx parser_yandex; do
    if docker-compose ps $service | grep -q "Up"; then
        print_status "✅ $service is running"
    else
        print_warning "⚠️  $service is not running"
        print_warning "Check logs: docker-compose logs $service"
    fi
done

print_status "🎉 Deployment completed!"
print_status ""
print_status "Useful commands:"
print_status "  View all logs:        docker-compose logs -f"
print_status "  View specific parser: docker-compose logs -f parser_2gis"
print_status "  Connect to database:  docker-compose exec postgres psql -U postgres -d mytravel_db"
print_status "  Stop all services:    docker-compose down"
print_status "  Restart service:      docker-compose restart parser_2gis"
print_status ""
print_status "Database is available on: localhost:5433"
print_status "To check parsed data: SELECT * FROM accommodations;"
print_status "To check parser logs: SELECT * FROM parsing_logs ORDER BY started_at DESC;"
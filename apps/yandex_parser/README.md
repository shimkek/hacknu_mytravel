# Yandex Maps Accommodation Parser

A comprehensive Go-based parser for scraping accommodation data from Yandex Maps for Kazakhstan tourism.

## Features

- 🏨 Supports 7 accommodation types (Hotels, Guesthouses, Sanatoriums, Camping, etc.)
- 🏙️ Covers 40+ Kazakhstan cities (major cities and tourist destinations)
- 📊 Realistic pagination and data generation with optional real Selenium scraping
- 💾 PostgreSQL database integration with production-ready schema
- 🔄 Automatic duplicate detection and updates
- 📈 Comprehensive logging and statistics
- 🌐 Selenium WebDriver support with ChromeDriver

## Quick Start

### Option 1: Automated Setup (Recommended)

```bash
# Run the setup script to install ChromeDriver and dependencies
./setup.sh

# Run the parser
./yandex_parser
```

### Option 2: Manual Setup

1. **Install ChromeDriver**:
   - Download from [ChromeDriver releases](https://chromedriver.chromium.org/)
   - Place `chromedriver` executable in project root or system PATH

2. **Build and run**:
```bash
go build -o yandex_parser main.go
./yandex_parser
```

## Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mytravel_db
DB_SSLMODE=disable

# Selenium Configuration (optional)
CHROMEDRIVER_PATH=./chromedriver
CHROME_BIN=/usr/bin/chromium-browser  # For Docker
```

## Docker Deployment

The Docker image includes ChromeDriver and all dependencies:

```bash
# Build image
docker build -t yandex-parser .

# Run container
docker run --network hacknu_mytravel_default \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  yandex-parser
```

## Output

The parser will populate the `accommodations` table with realistic data for Kazakhstan tourism, including:

- Accommodation details (name, type, location)
- Contact information (phone, email, website)
- Business details (rooms, capacity, pricing)
- Reviews and ratings
- Amenities and photos
- Geographic coordinates

## Statistics

Expected output: ~2000-3000 accommodation records across Kazakhstan cities.

## Selenium Support

- **Real scraping**: When ChromeDriver is available and working
- **Mock data fallback**: Generates realistic data when Selenium fails
- **Headless mode**: Optimized for server environments
- **Docker ready**: Pre-configured Chrome and ChromeDriver in container
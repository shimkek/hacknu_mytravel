# MyTravel - Simple Microservices Accommodation Parser

A simple microservices system that parses accommodation data from multiple websites and stores it in a centralized PostgreSQL database.

## Architecture

- **Database**: Simple PostgreSQL database with one init script
- **Parsers**: 5 independent Go microservices, each in its own Docker container
- **Data Sources**: 2GIS, Google Maps, Instagram, OLX, and Yandex
- **Deployment**: Simple docker-compose with custom PostgreSQL port (5433)

## Project Structure

```
├── apps/                          # Individual parser microservices
│   ├── 2gis_parser/
│   │   ├── main.go               # Simple counter that logs to DB
│   │   ├── go.mod                # Individual dependencies
│   │   └── Dockerfile            # Simple Docker build
│   ├── google_maps_parser/
│   ├── instagram_parser/
│   ├── olx_parser/
│   └── yandex_parser/
├── infrastructure/
│   └── database/
│       └── init.sql              # Simple database initialization
├── docker-compose.yml            # Simple multi-service deployment
├── deploy.sh                     # VPS deployment script
└── .env                          # Environment configuration
```

## Quick VPS Deployment

### 1. On your VPS, clone the repo:
```bash
git clone <your-repository-url>
cd hacknu_mytravel
```

### 2. Run the deployment script:
```bash
./deploy.sh
```

This script will:
- Pull latest changes from git
- Stop any existing containers
- Build and start all services
- Check that everything is working
- Show you useful commands

### 3. Check that it's working:
```bash
# View logs from all parsers
docker-compose logs -f

# Connect to database and see the data
docker-compose exec postgres psql -U postgres -d mytravel_db
```

In the database, run:
```sql
-- See all accommodations being created
SELECT * FROM accommodations;

-- See parser activity
SELECT * FROM parsing_logs ORDER BY started_at DESC;

-- Count accommodations by source
SELECT source_website, COUNT(*) FROM accommodations GROUP BY source_website;
```

## What Each Parser Does

Each parser is a simple Go application that:
- Connects to the PostgreSQL database
- Runs in a loop every 30-60 seconds
- Logs its activity to the `parsing_logs` table
- Creates test accommodations to demonstrate it's working
- Each parser has a different interval to show they're independent

**Current Implementation**: Simple counters for proof of concept
**Next Step**: Replace with actual website scraping logic

## Database Schema

**accommodations table**: Stores all accommodation data exactly as specified by project requirements

### Required Fields (Project Specification):
- **Название объекта** → `name` (VARCHAR 500) - Name of accommodation
- **Координаты (GPS)** → `latitude`, `longitude` (DECIMAL) - GPS coordinates  
- **Адрес** → `address` (TEXT) - Full address
- **Тип размещения (категория)** → `accommodation_type` (ENUM) - Category/type
- **Контактные данные** → `phone`, `email`, `social_media_links` (JSONB) - Contact info
- **Сайт/страница в соцсети** → `website_url`, `social_media_page` - Website and social pages
- **Описание услуг** → `service_description` (TEXT) - Service description
- **Количество номеров/мест** → `room_count`, `capacity` (INTEGER) - Rooms and capacity
- **Ценовой диапазон** → `price_range_min`, `price_range_max` (DECIMAL) - Price range
- **Фотографии** → `photos` (JSONB) - Photo links array
- **Отзывы и рейтинги** → `rating`, `review_count`, `reviews` (JSONB) - Reviews and ratings
- **Инфраструктура** → `amenities` (JSONB) - WiFi, parking, kitchen, etc.
- **Статус проверки** → `verification_status` (ENUM) - new/verified/in_development
- **Дата последнего обновления** → `last_updated` (TIMESTAMP) - Auto-updated on changes

### Technical Fields:
- `source_website` - Which parser found this data
- `external_id` - ID from source website  
- `created_at` - When record was created
- `deleted_at` - Soft delete timestamp

**parsing_logs table**: Tracks parser activity and statistics

## Manual Commands

### View specific parser logs:
```bash
docker-compose logs -f parser_2gis
docker-compose logs -f parser_google_maps
docker-compose logs -f parser_instagram
docker-compose logs -f parser_olx
docker-compose logs -f parser_yandex
```

### Restart a specific parser:
```bash
docker-compose restart parser_2gis
```

### Connect to database:
```bash
docker-compose exec postgres psql -U postgres -d mytravel_db
```

### Stop everything:
```bash
docker-compose down
```

### Rebuild and restart:
```bash
docker-compose down
docker-compose up --build -d
```

## Local Development

### Run individual parser locally:
```bash
cd apps/2gis_parser
go mod download
go run main.go
```

### Start just the database:
```bash
docker-compose up postgres -d
```

## Environment Variables

Edit `.env` file to change database settings:
- `DB_HOST`: Database host (default: localhost)  
- `DB_PORT`: Database port (default: 5433)
- `DB_USER`, `DB_PASSWORD`, `DB_NAME`: Database credentials

## Next Steps for Production

1. **Replace counters with real parsers**: Add actual website scraping logic
2. **Add rate limiting**: Respect website scraping limits
3. **Error handling**: Add retry logic and better error recovery
4. **Data validation**: Validate scraped data before insertion
5. **Monitoring**: Add health checks and metrics

## Why This Structure?

- **Simple**: No complex migrations, no shared internal packages
- **Independent**: Each parser has its own go.mod and can be developed separately
- **VPS-friendly**: Uses port 5433 to avoid conflicts with existing PostgreSQL
- **Easy deployment**: One script handles everything
- **Local build**: No need for Docker registry, builds locally on VPS
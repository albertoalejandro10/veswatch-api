# VESWatch API

A lightweight Go service that provides Venezuelan exchange rate information by scraping the BCV (Banco Central de Venezuela) official rate and fetching Binance P2P USDT/VES rates.

## Legal Disclaimer

> "VESWatch provides reference exchange rates obtained from public sources. This information is not official financial advice."

## Features

- 🇻🇪 **BCV Rate Scraping** - Scrapes official USD rate from bcv.org.ve using Colly
- 💱 **Binance P2P** - Fetches USDT/VES market rates from Binance P2P
- 📊 **Breach Calculation** - Calculates percentage difference between rates
- ⏰ **Smart Scheduling** - BCV updates daily (Mon-Fri), Binance every 5 minutes
- 🚀 **Fly.io Ready** - Docker-based deployment configuration included

## API Endpoints

### `GET /rates`

Returns current exchange rates:

```json
{
  "bcv": 45.82,
  "binance": 46.31,
  "breach": 1.07,
  "updatedAt": "2026-01-15T11:00:00-04:00"
}
```

### `GET /health`

Health check endpoint:

```json
{
  "status": "ok"
}
```

### `GET /`

API information:

```json
{
  "name": "VESWatch API",
  "version": "1.0.0",
  "endpoints": "/rates"
}
```

## Local Development

### Prerequisites

- Go 1.23 or later
- Git

### Running Locally

1. **Clone and navigate to the api directory:**

```bash
cd api
```

2. **Install dependencies:**

```bash
go mod tidy
```

3. **Run the server:**

```bash
go run ./cmd/server
```

4. **Test the endpoint:**

```bash
curl http://localhost:8080/rates
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `TZ` | System | Timezone for scheduling |

## Deployment to Fly.io

### Prerequisites

- [Fly CLI](https://fly.io/docs/hands-on/install-flyctl/) installed
- Fly.io account

### Deploy

1. **Login to Fly.io:**

```bash
fly auth login
```

2. **Create the app (first time only):**

```bash
cd api
fly launch --no-deploy
```

3. **Deploy:**

```bash
fly deploy
```

4. **Check status:**

```bash
fly status
```

5. **View logs:**

```bash
fly logs
```

6. **Test the deployed API:**

```bash
curl https://veswatch-api.fly.dev/rates
```

## Project Structure

```
api/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── http/
│   │   └── handlers.go       # HTTP handlers
│   ├── rates/
│   │   ├── model.go          # Data models
│   │   └── service.go        # Rate service
│   ├── scheduler/
│   │   └── scheduler.go      # Job scheduler
│   └── scraper/
│       ├── bcv.go            # BCV scraper (Colly)
│       └── binance.go        # Binance P2P fetcher
├── Dockerfile                # Multi-stage Docker build
├── fly.toml                  # Fly.io configuration
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## Architecture

### Data Flow

```
┌─────────────┐     ┌─────────────┐
│   BCV.org   │     │  Binance    │
│  (scrape)   │     │  P2P API    │
└──────┬──────┘     └──────┬──────┘
       │                   │
       ▼                   ▼
┌─────────────────────────────────┐
│          Rate Service           │
│   (in-memory storage + calc)    │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│         HTTP Handler            │
│         GET /rates              │
└─────────────────────────────────┘
```

### Scheduling

- **BCV**: Once daily at 11:30 AM Venezuela time (Mon-Fri only)
- **Binance**: Every 5 minutes

### Reliability

- Failed scrapes preserve the last known value
- No panics on external failures
- All errors are logged

## Technologies

- **Go 1.23** - Latest stable Go
- **gocolly/colly** - Web scraping framework
- **Standard library** - HTTP server, JSON encoding
- **Docker** - Multi-stage builds
- **Fly.io** - Edge deployment platform

## License

MIT

---

Built with 💛 for Venezuela 🇻🇪

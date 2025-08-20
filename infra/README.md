# Audio Book AI - Infrastructure

This directory contains the Docker Compose configuration and infrastructure setup for the Audio Book AI application.

## Architecture Overview

The application uses a microservices architecture with the following components:

- **Supabase**: Cloud-based Auth + PostgreSQL + Storage (not dockerized)
- **Redis**: Job queue for background processing
- **Go API**: REST API service
- **Go Worker**: Background job processor for summarization using Gemini API
- **Go Transcriber**: Rev.ai-based audio transcription worker
- **Go AI Orchestrator**: AI workflow orchestration using Gemini API
- **Next.js Web App**: Frontend application

## Services Overview

The Docker Compose infrastructure includes:

- **redis**: Job queue for background processing
- **api**: Go REST API service
- **worker**: Go worker for summarize jobs using Gemini
- **transcriber**: Go transcriber using Rev.ai
- **ai_orchestrator**: Go orchestrator worker using Gemini
- **web**: Next.js web application

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- At least 4GB of available RAM
- 10GB of available disk space
- Supabase project set up
- Gemini API key
- Rev.ai API key

### Setup Instructions

1. **Clone and navigate to the project:**

   ```bash
   cd audio-book-ai
   ```

2. **Copy environment variables:**

   ```bash
   cp infra/env.example infra/.env
   ```

3. **Edit the environment file:**

   ```bash
   # Edit infra/.env and add your Supabase credentials and API keys
   nano infra/.env
   ```

4. **Start all services:**

   ```bash
   cd infra
   docker-compose up -d
   ```

5. **Check service status:**
   ```bash
   docker-compose ps
   ```

## Service Details

### Redis (redis)

- **Port**: 6379
- **Purpose**: Job queue for background processing
- **Volume**: `redisdata`

### Go API (api)

- **Port**: 8080
- **Purpose**: REST API service
- **Dependencies**: redis

### Go Worker (worker)

- **Purpose**: Background job processing for summarization
- **Dependencies**: redis
- **AI Provider**: Gemini API

### Go Transcriber (transcriber)

- **Purpose**: Audio transcription using Rev.ai
- **Dependencies**: redis
- **AI Provider**: Rev.ai API

### Go AI Orchestrator (ai_orchestrator)

- **Purpose**: AI workflow orchestration for tags and embeddings
- **Dependencies**: redis
- **AI Provider**: Gemini API

### Next.js Web App (web)

- **Port**: 3000
- **Purpose**: Frontend application
- **Dependencies**: api

## Environment Variables

Key environment variables needed:

### Supabase Configuration

- `SUPABASE_URL`: Your Supabase project URL
- `SUPABASE_PUBLISHABLE_KEY`: Supabase publishable key (formerly anon key)
- `SUPABASE_SECRET_KEY`: Supabase secret key (formerly service role key)
- `SUPABASE_STORAGE_BUCKET`: Storage bucket name (default: "audio")

**📁 Storage Setup**: See [STORAGE_SETUP.md](./STORAGE_SETUP.md) for detailed instructions on configuring Supabase storage with S3-compatible endpoints.

### Redis Configuration

- `REDIS_URL`: Redis connection string
- `JOBS_PREFIX`: Job queue prefix (default: "audiobooks")

### Gemini API Configuration

- `GEMINI_API_KEY`: Your Gemini API key
- `GEMINI_URL`: Gemini API base URL
- `GEMINI_MODEL`: Gemini model to use (default: "gemini-2.0-flash-exp")

### Rev.ai Configuration

- `REV_AI_API_KEY`: Your Rev.ai API key
- `REV_AI_URL`: Rev.ai API base URL

## Development Workflow

### Starting Development Environment

```bash
cd infra
docker-compose up -d
```

### Stopping Services

```bash
docker-compose down
```

### Rebuilding Services

```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f worker
docker-compose logs -f transcriber
docker-compose logs -f web
```

## Database Management

Since you're using Supabase, database operations are handled through:

1. **Supabase Dashboard**: For schema management and migrations
2. **Direct Connection**: Optional direct PostgreSQL connection for advanced operations

## Troubleshooting

### Common Issues

**Service won't start:**

```bash
# Check logs
docker-compose logs [service-name]

# Check if ports are in use
netstat -tulpn | grep :8080
```

**API key issues:**

```bash
# Check if API keys are set correctly
docker-compose logs worker
docker-compose logs transcriber
docker-compose logs ai_orchestrator
```

**Storage issues:**

```bash
# Check storage configuration
docker-compose logs api

# Test storage setup
cd api && go run test_storage.go
```

See [STORAGE_SETUP.md](./STORAGE_SETUP.md) for detailed storage troubleshooting.

**Redis connection issues:**

```bash
# Check if Redis is running
docker-compose ps redis

# Check Redis logs
docker-compose logs redis
```

### Useful Commands

```bash
# Clean up everything
docker-compose down -v --remove-orphans

# Rebuild specific service
docker-compose build api

# Execute commands in running container
docker-compose exec api /bin/sh
docker-compose exec worker /bin/sh

# View resource usage
docker stats
```

## File Structure

```
infra/
├── docker-compose.yml    # Main Docker Compose configuration
├── env.example          # Environment variables template
├── .env                # Environment variables (create from template)
├── Makefile            # Common operations and commands
└── README.md           # This file
```

## Support

For issues related to the infrastructure setup, please check:

1. Docker and Docker Compose versions
2. Available system resources
3. Port conflicts
4. Supabase configuration
5. Environment variable configuration
6. API key configuration for Gemini and Rev.ai

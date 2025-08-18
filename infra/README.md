# Audio Book AI - Infrastructure

This directory contains the Docker Compose configuration and infrastructure setup for the Audio Book AI application.

## Architecture Overview

The application uses a microservices architecture with the following components:

- **Supabase**: Cloud-based Auth + PostgreSQL + Storage (not dockerized)
- **Redis**: Job queue for background processing
- **Ollama**: Local LLM serving for AI operations
- **Go API**: REST API service
- **Go Worker**: Background job processor for summarization
- **Python Transcriber**: Whisper-based audio transcription worker
- **Go AI Orchestrator**: AI workflow orchestration
- **Next.js Web App**: Frontend application

## Services Overview

The Docker Compose infrastructure includes:

- **redis**: Job queue for background processing
- **ollama**: LLM serving (llama3.1, nomic-embed-text)
- **api**: Go REST API service
- **worker**: Go worker for summarize jobs
- **transcriber**: Python Whisper worker
- **ai_orchestrator**: Go orchestrator worker
- **web**: Next.js web application

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- At least 8GB of available RAM (for Ollama models)
- 20GB of available disk space
- Supabase project set up

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
   # Edit infra/.env and add your Supabase credentials and configuration
   nano infra/.env
   ```

4. **Start all services:**

   ```bash
   cd infra
   docker-compose up -d
   ```

5. **Pull Ollama models:**

   ```bash
   make pull-models
   ```

6. **Check service status:**
   ```bash
   docker-compose ps
   ```

## Service Details

### Redis (redis)

- **Port**: 6379
- **Purpose**: Job queue for background processing
- **Volume**: `redisdata`

### Ollama (ollama)

- **Port**: 11434
- **Purpose**: Local LLM serving
- **Models**: llama3.1, nomic-embed-text
- **Volume**: `ollama`

### Go API (api)

- **Port**: 8080
- **Purpose**: REST API service
- **Dependencies**: redis, ollama

### Go Worker (worker)

- **Purpose**: Background job processing for summarization
- **Dependencies**: redis, ollama

### Python Transcriber (transcriber)

- **Purpose**: Audio transcription using Whisper
- **Dependencies**: redis

### Go AI Orchestrator (ai_orchestrator)

- **Purpose**: AI workflow orchestration
- **Dependencies**: redis, ollama

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

### Redis Configuration

- `REDIS_URL`: Redis connection string
- `JOBS_PREFIX`: Job queue prefix (default: "audiobooks")

### Ollama Configuration

- `OLLAMA_URL`: Ollama service URL
- `AI_SUMMARY_MODEL`: Model for summarization (default: "llama3.1")
- `AI_EMBED_MODEL`: Model for embeddings (default: "nomic-embed-text")

### Whisper Configuration

- `WHISPER_MODEL`: Whisper model size (default: "base")
- `WHISPER_LANGUAGE`: Language detection (default: "auto")

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

### Pulling Ollama Models

```bash
make pull-models
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

**Ollama model issues:**

```bash
# Check if models are downloaded
docker-compose exec ollama ollama list

# Pull missing models
make pull-models
```

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
2. Available system resources (especially RAM for Ollama)
3. Port conflicts
4. Supabase configuration
5. Environment variable configuration

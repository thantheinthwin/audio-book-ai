.PHONY: up down pull-models migrate logs status clean rebuild build-with-retry

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Pull Ollama models
pull-models:
	docker-compose exec ollama ollama pull llama3.1
	docker-compose exec ollama ollama pull nomic-embed-text

# Run database migrations (optional - can use Supabase dashboard)
migrate:
	@echo "Database migrations should be handled through Supabase dashboard"
	@echo "Or use direct connection if needed"

# Show logs for all services
logs:
	docker-compose logs -f

# Show status of all services
status:
	docker-compose ps

# Clean up everything (containers, networks, volumes)
clean:
	docker-compose down -v --remove-orphans

# Rebuild all services
rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Build with retry logic for network issues
build-with-retry:
	@echo "Building services with retry logic..."
	docker-compose down
	@echo "Building transcriber service first (most likely to fail)..."
	docker-compose build --no-cache transcriber || (echo "Retrying transcriber build..." && docker-compose build --no-cache transcriber)
	@echo "Building remaining services..."
	docker-compose build --no-cache
	docker-compose up -d

# Show logs for specific services
logs-api:
	docker-compose logs -f api

logs-worker:
	docker-compose logs -f worker

logs-transcriber:
	docker-compose logs -f transcriber

logs-web:
	docker-compose logs -f web

logs-ollama:
	docker-compose logs -f ollama

logs-redis:
	docker-compose logs -f redis

# Restart specific services
restart-api:
	docker-compose restart api

restart-worker:
	docker-compose restart worker

restart-transcriber:
	docker-compose restart transcriber

restart-web:
	docker-compose restart web

# Shell access to containers
shell-api:
	docker-compose exec api /bin/sh

shell-worker:
	docker-compose exec worker /bin/sh

shell-transcriber:
	docker-compose exec transcriber /bin/bash

shell-web:
	docker-compose exec web /bin/sh

# Check Ollama models
models:
	docker-compose exec ollama ollama list

# Help
help:
	@echo "Available commands:"
	@echo "  up              - Start all services"
	@echo "  down            - Stop all services"
	@echo "  pull-models     - Pull Ollama models"
	@echo "  migrate         - Run database migrations (use Supabase dashboard)"
	@echo "  logs            - Show logs for all services"
	@echo "  status          - Show status of all services"
	@echo "  clean           - Clean up everything"
	@echo "  rebuild         - Rebuild all services"
	@echo "  build-with-retry - Rebuild with retry logic for network issues"
	@echo ""
	@echo "Service-specific logs:"
	@echo "  logs-api        - Show API logs"
	@echo "  logs-worker     - Show worker logs"
	@echo "  logs-transcriber - Show transcriber logs"
	@echo "  logs-web        - Show web app logs"
	@echo "  logs-ollama     - Show Ollama logs"
	@echo "  logs-redis      - Show Redis logs"
	@echo ""
	@echo "Service restart:"
	@echo "  restart-api     - Restart API service"
	@echo "  restart-worker  - Restart worker service"
	@echo "  restart-transcriber - Restart transcriber service"
	@echo "  restart-web     - Restart web service"
	@echo ""
	@echo "Shell access:"
	@echo "  shell-api       - Access API container shell"
	@echo "  shell-worker    - Access worker container shell"
	@echo "  shell-transcriber - Access transcriber container shell"
	@echo "  shell-web       - Access web container shell"
	@echo ""
	@echo "Other:"
	@echo "  models          - List Ollama models"
	@echo "  help            - Show this help"

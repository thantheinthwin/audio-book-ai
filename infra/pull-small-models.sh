#!/bin/bash

echo "Pulling small AI models for faster processing..."

# Pull small Whisper model (tiny)
echo "Pulling Whisper tiny model..."
docker-compose exec ollama ollama pull whisper-tiny

# Pull small Llama2 model (7B)
echo "Pulling Llama2 7B model..."
docker-compose exec ollama ollama pull llama2:7b

# Pull small embedding model
echo "Pulling Nomic embed text model..."
docker-compose exec ollama ollama pull nomic-embed-text

echo "All small models pulled successfully!"
echo ""
echo "Available models:"
docker-compose exec ollama ollama list

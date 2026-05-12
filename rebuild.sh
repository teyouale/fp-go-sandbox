#!/bin/bash

echo "🛑 Stopping containers..."
docker-compose down

echo "🗑️  Removing old images..."
docker rmi codapi-server:latest fpgo-sandbox:latest 2>/dev/null || true

echo "🔨 Building new images..."
docker-compose build --no-cache --pull

echo "🚀 Starting services..."
docker-compose up

echo "✅ Done! Codapi should be running on http://localhost:1313"

# Made with Bob

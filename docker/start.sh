#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "================================================"
echo "  AI Platform - Docker Build & Start"
echo "================================================"
echo ""

# Check for Docker
if ! command -v docker &>/dev/null; then
    echo "ERROR: Docker not found. Please install Docker first."
    exit 1
fi

# Check for docker compose
if ! docker compose version &>/dev/null; then
    echo "ERROR: docker compose not found. Please install Docker Compose first."
    exit 1
fi

echo "[1/2] Building Docker images..."
docker compose build
echo ""

echo "[2/2] Starting services..."
docker compose up -d
echo ""

echo "================================================"
echo "  Services starting on:"
echo "  - API Gateway:    http://localhost:8080"
echo "  - AI Gateway:     http://localhost:8081"
echo "  - Web Frontend:   http://localhost"
echo "  - PostgreSQL:     localhost:5432"
echo ""
echo "  Run 'docker compose logs -f' to follow logs"
echo "  Run 'docker compose down' to stop all services"
echo "================================================"

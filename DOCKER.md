# Docker & Deployment Guide

VPN Management API with complete Docker and GitHub Actions CI/CD setup.

## 📦 Quick Start

### Prerequisites
- Docker & Docker Compose 2.0+
- Git
- (Optional) Make for convenience commands

### Run with Docker Compose

```bash
# Start all services (API + PostgreSQL)
docker-compose up -d

# Check logs
docker-compose logs -f vpn-api

# Access API
curl http://localhost:8080/health

# Stop services
docker-compose down
```

API will be available at: `http://localhost:8080`
PostgreSQL at: `localhost:5432`

## 🛠️ Using Make Commands

If you have `make` installed, use these convenience commands:

```bash
# View all available commands
make help

# Development workflow
make dev           # Build, start, and run migrations
make docker-logs   # View container logs
make docker-down   # Stop all containers
make docker-clean  # Remove containers and volumes

# Local development (without Docker)
make build         # Build binary locally
make run           # Run locally
make test          # Run tests
make lint          # Run linters
make fmt           # Format code
```

## 🐳 Docker Configuration

### Dockerfile Features
- **Multi-stage build**: Optimized image size (~50MB)
- **Security**: Non-root user execution
- **Alpine Linux**: Minimal attack surface
- **Health checks**: Automatic container health monitoring
- **Distroless-ready**: Can be swapped to distroless for smaller images

### docker-compose.yml Services

#### PostgreSQL
```yaml
Service: postgres
Port: 5432
User: vpn_user (configurable)
Password: vpn_password (configurable)
Database: vpn_db (configurable)
Volumes: postgres_data (persistent)
```

#### VPN API
```yaml
Service: vpn-api
Port: 8080
Depends on: postgres (with health check)
Health Check: Available
Environment: Fully configurable via .env
```

## 🔧 Configuration

### Environment Variables

Create `.env` file in project root:

```bash
# Database Configuration
DB_USER=vpn_user
DB_PASSWORD=vpn_password
DB_NAME=vpn_db
DB_HOST=postgres
DB_PORT=5432

# Application
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=info

# WireGuard Agent
WG_AGENT_URL=http://wireguard-agent:9000
WG_AGENT_API_KEY=your-api-key
```

### Using different environment files

```bash
# Development
docker-compose --env-file .env.dev up -d

# Staging
docker-compose --env-file .env.staging up -d

# Production
docker-compose --env-file .env.prod up -d
```

## 📊 Database Migrations

### With Docker

```bash
# Run migrations when starting
docker-compose up -d
docker-compose exec vpn-api go run cmd/migrate/main.go up

# Rollback
docker-compose exec vpn-api go run cmd/migrate/main.go down
```

### With Make

```bash
make db-migrate-up
make db-migrate-down
make db-seed
```

## 🏗️ Building Custom Images

### Build with custom tag

```bash
docker build -t vpn-api:1.0.0 .
docker run -p 8080:8080 vpn-api:1.0.0
```

### Build with specific Go version

```bash
docker build --build-arg GO_VERSION=1.24 -t vpn-api:latest .
```

## 📝 Logs and Monitoring

### View logs

```bash
# API logs
docker-compose logs -f vpn-api

# Database logs
docker-compose logs -f postgres

# All logs
docker-compose logs -f
```

### Access API logs directory

```bash
# Logs are mounted to ./logs
tail -f logs/app.log
```

## 🔐 Security Best Practices

### Before Production

1. **Change default credentials**
   ```bash
   # Update in .env or docker-compose.yml
   DB_PASSWORD=<strong-random-password>
   WG_AGENT_API_KEY=<strong-random-key>
   ```

2. **Use secrets management**
   ```bash
   # Using Docker secrets
   docker secret create db_password db_password.txt
   ```

3. **Network isolation**
   ```yaml
   # docker-compose.yml already uses custom network
   networks:
     - vpn-network
   ```

4. **Regular updates**
   ```bash
   # Update base images
   docker pull postgres:16-alpine
   docker pull alpine:3.19
   ```

5. **Scan for vulnerabilities**
   ```bash
   # Using Trivy
   trivy image vpn-api:latest
   ```

## 📈 Production Deployment

### Multi-container setup example

```bash
# Build image
docker build -t registry.example.com/vpn-api:1.0.0 .

# Push to registry
docker push registry.example.com/vpn-api:1.0.0

# Deploy with Kubernetes (example)
kubectl apply -f k8s/deployment.yaml
```

### Docker Swarm deployment

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml vpn
```

### Environment-specific setup

```bash
# Production with healthchecks and restarts
docker run -d \
  --name vpn-api \
  --restart unless-stopped \
  --health-cmd="curl -f http://localhost:8080/health" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  -p 8080:8080 \
  -e DB_HOST=prod-postgres \
  -e APP_ENV=production \
  vpn-api:1.0.0
```

## 🚀 CI/CD Workflows

### Automated workflows included

1. **CI Pipeline** (`.github/workflows/ci.yml`)
   - Code formatting check
   - Linting (golangci-lint)
   - Unit tests
   - Docker build cache
   - Security scanning (Trivy)
   - Code quality analysis

2. **Deploy Pipeline** (`.github/workflows/deploy.yml`)
   - Automatic Docker image build & push
   - Push to GitHub Container Registry (GHCR)
   - Triggered on: push to main, tags, PRs

3. **Scheduled Jobs** (`.github/workflows/schedule.yml`)
   - Daily dependency updates check
   - Security audit
   - CodeQL analysis
   - License compliance
   - Performance benchmarks

### GitHub Container Registry

```bash
# Login to GHCR
echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin

# Tag and push
docker tag vpn-api:latest ghcr.io/your-org/vpn-api:latest
docker push ghcr.io/your-org/vpn-api:latest

# Pull from GHCR
docker pull ghcr.io/your-org/vpn-api:latest
```

## 🧪 Testing

### Run tests in container

```bash
docker-compose exec vpn-api go test -v ./...
```

### Run with coverage

```bash
docker-compose exec vpn-api go test -v -cover ./...
```

### Load testing

```bash
# Using Apache Bench
ab -n 1000 -c 10 http://localhost:8080/health

# Using Hey
hey -n 1000 -c 10 http://localhost:8080/health
```

## 🆘 Troubleshooting

### Container won't start

```bash
# Check logs
docker-compose logs vpn-api

# Inspect container
docker-compose ps
```

### Database connection issues

```bash
# Test connectivity
docker-compose exec vpn-api pg_isready -h postgres -U vpn_user

# Check database
docker-compose exec postgres psql -U vpn_user -d vpn_db -c "\dt"
```

### Port already in use

```bash
# Find process using port
lsof -i :8080

# Or use different port
docker-compose -f docker-compose.yml -p myapp up -d
```

### Rebuild from scratch

```bash
# Remove all and rebuild
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

## 📖 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Best practices for writing Dockerfiles](https://docs.docker.com/develop/dev-best-practices/dockerfile-best-practices/)

## 📄 License

This project uses Apache 2.0 License - see LICENSE file for details.

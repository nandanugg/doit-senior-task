# Distributed URL Shortener

A production-ready URL shortening service designed for high throughput, demonstrating cloud-native architecture patterns, operational excellence, and infrastructure as code.

**Key Features:**
- 🚀 High-performance URL shortening and redirection
- 📊 Click tracking and analytics
- ⏰ Configurable TTL with efficient expiration management
- 🔒 PII-compliant (no IP logging)
- 📈 Auto-scaling for 10,000+ requests/second
- ☁️ Cloud-native deployment (AWS ECS Fargate)
- 🏗️ Infrastructure as Code (Terraform)

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Gap Analysis](#2-gap-analysis)
3. [Capacity Planning](#3-capacity-planning)
4. [Service Level Management](#4-service-level-management)
5. [Future Enhancements](#5-future-enhancements)
6. [Quick Start](#quick-start)
7. [API Reference](#api-reference)
8. [Development](#development)

---

## 1. Architecture Overview

### Data Model

**Primary store (Redis AOF):**
```text
url_id_sequence            -> INCR for sequential IDs
url:{id}                   -> long_url (string, TTL enforced)
```

**Analytics store (PostgreSQL):**
```sql
CREATE TABLE url_analytics (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL,
    long_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    click_count BIGINT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMPTZ
);
```

**Key design decisions:**
1. **Sequential IDs**: Redis `INCR` is atomic and fast.
2. **Short codes**: Hex-encoded IDs with character masking for readability (see [ADR-001](docs/adr/001-hex-code-masking.md)).
3. **Persistence**: Redis AOF keeps URL mappings durable; PostgreSQL stores analytics (see [ADR-003](docs/adr/003-persistent-database-over-in-memory.md)).
4. **Thread-safe counters**: Click counts are incremented via PostgreSQL atomic updates (see [ADR-002](docs/adr/002-count-persistence.md)).

### High-Level System Architecture

```
Users → ALB → ECS (Go API)
               ├── Redis (AOF): URL mappings + TTL
               └── PostgreSQL: analytics + stats
```

### Request Flow

**Create (POST /s):**
1. Validate URL + TTL
2. `INCR` for ID, `SET url:{id} EX {ttl}`
3. Insert analytics row
4. Return short code

**Redirect (GET /s/{code}):**
1. Decode short code → ID
2. `GET url:{id}`
3. Async analytics update
4. Return 302

**Stats (GET /stats/{code}):**
1. Decode short code → ID
2. SELECT analytics row
3. Return JSON

**Observability:** every response includes `X-Processing-Time-Micros`.

---

## 2. Gap Analysis

### Why In-Memory Alone Fails in a Stateless Cloud
ECS tasks are ephemeral: scaling, deployments, or crashes wipe local memory. With multiple tasks, each instance would hold a different subset of URLs, causing inconsistent redirects and data loss.

### Managed Storage Choice
- **Redis (AOF)** provides shared, durable mappings with TTL-based cleanup.
- **PostgreSQL** preserves analytics and supports the `/stats` endpoint.

This keeps business logic storage-agnostic while satisfying durability and consistency in a stateless deployment.

---

## 3. Capacity Planning

### Storage (12-Month Projection)

**Assumptions:**
- 100M new URLs/month
- Avg long URL length: 200 chars
- TTL: 48 hours (configurable)

**Per-record estimate:**
```
Redis entry (key + value + overhead): ~280 bytes
```

**Active dataset (48h TTL):**
```
Daily creates: 3.33M
Active URLs:   6.66M
Active size:   ~1.9 GB
```

**Total storage (with AOF + replica):**
- Redis RAM: ~2 GB
- AOF disk: ~2-3 GB
- Replica: ~2 GB
- **Total: ~6-7 GB**

**Without TTL (archive scenario):**
- 1.2B URLs/year → ~336 GB RAM
- Not viable on a single Redis node; requires cold storage (S3/DB) or sharding.

### Scaling for 10,000 RPS Redirects

- Redis handles O(1) lookups; async analytics keeps redirects fast.
- ECS scales horizontally (2–10 tasks).
- **Strategy:** connection pooling + pipelined writes → add read replica → Redis Cluster sharding.

---

## 4. Service Level Management

### Capacity Envelope (Baseline Infra)
These SLOs are scoped to (see AWS sizing docs for [Fargate](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) and [ElastiCache node types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html)):
- **ECS Fargate:** .256 CPU / 512 MB per task, autoscaling 2–10 tasks.
- **Redis (ElastiCache):** single primary + replica, memory sized for the active dataset (~2 GB @ 48h TTL).
- **RDS PostgreSQL:** db.t3.micro for analytics writes only (off the hot path).
  - If RDS saturates, redirects still meet SLOs but `/stats` can lag.

**Assumption:** 1 active user ≈ 1 redirect/sec. With the current request path (Redis GET + minimal Go logic + async analytics), we budget **~1,000 rps per task** as a conservative envelope.
**This assumes instance-class network/IO limits are not saturated** for Redis and RDS.

**Guarantees within this envelope:**
- **Up to ~2,000 active users (2 tasks):** p95 < 50ms, 99.9% availability.
- **Up to ~10,000 active users (10 tasks):** p95 < 50ms, 99.9% availability.

Above this, SLOs require scaling Redis (cluster/sharding) or increasing task size.

**SLI 1: Redirect Availability**
- **SLO:** 99.9% over 30 days
- **Measure:** 302 / total redirects

**SLI 2: Redirect Latency**
- **SLO:** p95 < 50ms
- **Measure:** ALB TargetResponseTime p95

**On-call scenario:** Redis AOF rewrite or client pool exhaustion causes timeouts and elevated p95.
- Mitigate by scaling ECS, increasing Redis pool size, and tuning AOF rewrite thresholds.

---

## Quick Start

### Prerequisites

- **Go** 1.25+
- **PostgreSQL** 18+ (analytics)
- **Redis** 8+ (primary URL store)
- **Docker** (for containerized deployment)
- **Make** (for development workflows)

### Local Development

1. **Clone repository:**
   ```bash
   git clone <repository-url>
   cd url-shortener
   ```

2. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your PostgreSQL connection string (optional: REDIS_URL)
   ```

3. **Start PostgreSQL + Redis (Docker):**
   ```bash
   docker-compose up -d postgres redis
   ```

4. **Run migrations:**
   ```bash
   make migrate-up
   ```

5. **Start server:**
   ```bash
   make run
   # Server starts on http://localhost:8080
   ```

### Production Deployment

See [terraform/README.md](terraform/README.md) for complete AWS deployment guide.

Quick steps:
```bash
cd terraform
terraform init
terraform apply
```

---

## API Reference

### Create Short URL

**Endpoint:** `POST /s`

**Request:**
```json
{
  "long_url": "https://example.com/very/long/url",
  "ttl_seconds": 86400
}
```

**Response (200 OK):**
```json
{
  "short_code": "a3f7c2d",
  "short_url": "http://localhost:8080/s/a3f7c2d",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2026-01-11T10:00:00Z",
  "expires_at": "2026-01-12T10:00:00Z"
}
```

**Headers:**
- `X-Processing-Time-Micros`: Internal execution time in microseconds

### Redirect to Long URL

**Endpoint:** `GET /s/{short_code}`

**Response (302 Found):**
```
Location: https://example.com/very/long/url
X-Processing-Time-Micros: 12500
```

**Response (404 Not Found):**
URL not found or expired

### Get URL Statistics

**Endpoint:** `GET /stats/{short_code}`

**Response (200 OK):**
```json
{
  "short_code": "a3f7c2d",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2026-01-11T10:00:00Z",
  "expires_at": "2026-01-12T10:00:00Z",
  "click_count": 42,
  "last_accessed_at": "2026-01-11T15:30:00Z"
}
```

**Headers:**
- `X-Processing-Time-Micros`: Internal execution time in microseconds

---

## Development

### Run Tests

```bash
# Unit tests
make test

# Property-based tests
make test-property

# Integration tests (requires database)
make test-integration

# Load tests (requires k6)
make test-concurrent
```

### Linting and Formatting

```bash
# Format code
make fmt

# Run linter
make lint
```

### Generate Mocks

```bash
make mocks
```

---

## Architecture Decision Records (ADRs)

- [ADR-001: Hex Code Masking for ID Obfuscation](docs/adr/001-hex-code-masking.md)
- [ADR-002: Count Persistence Strategy](docs/adr/002-count-persistence.md)
- [ADR-003: Redis AOF Primary Storage](docs/adr/003-persistent-database-over-in-memory.md)

---

## License

MIT

---

## Support

For issues or questions, please create an issue in the GitHub repository.

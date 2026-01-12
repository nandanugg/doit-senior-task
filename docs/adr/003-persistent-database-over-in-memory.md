# ADR 003: Redis AOF Primary Storage Over PostgreSQL BRIN

**Status:** Accepted

**Date:** 2026-02-01

**Context:** Storage layer implementation

## Context

The requirements allow in-memory storage for the exercise, but the system must be production-ready and durable. We initially documented PostgreSQL with BRIN indexes as the primary store. In practice, our access pattern is dominated by **random key lookups** (short code → URL), and BRIN is a **lossy index** optimized for ordered scans, which adds CPU overhead for point lookups.

## Decision

Use **Redis with AOF persistence** as the primary store for URL mappings, and keep **PostgreSQL for analytics**.

**Redis responsibilities:**
- Store URL mappings as `url:{id} -> long_url`
- Generate sequential IDs with `INCR`
- Enforce TTL via `EXPIRE`
- Persist data via AOF (`appendfsync everysec`)

**PostgreSQL responsibilities:**
- Store click counts and audit fields (`url_analytics`)

## Rationale

- **Lookup efficiency**: O(1) key lookups fit the redirect path.
- **Durability**: AOF provides persistence without BRIN index overhead.
- **Atomic ID generation**: `INCR` is atomic across concurrent writers.
- **TTL-based retention**: Expiration is handled by Redis without background sweeps.

## Trade-offs

- **Memory-bound storage**: Redis scales with RAM; TTL keeps growth bounded.
- **Operational overhead**: AOF rewrite and replication require monitoring.
- **Limited ad-hoc queries**: Analytics must live in PostgreSQL or another store.

## Consequences

- Redirect latency improves due to Redis lookups.
- URL storage no longer depends on PostgreSQL indexes.
- Analytics remain durable and queryable in PostgreSQL.

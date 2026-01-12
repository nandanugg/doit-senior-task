# ADR 002: Count Persistence Strategy

**Status:** Accepted

**Date:** 2026-01-11

**Context:**
- modules/core/service/link_redirector.go:33-54
- modules/core/internal/repo/db/url_analytic_repo.go:60-68
- migrations/000001_init.up.sql:15

## Context

We track click counts under high concurrency while keeping redirect latency low. URL mappings live in Redis, while analytics (click counts, timestamps) are stored in PostgreSQL.

## Decision

We persist click counts in PostgreSQL and update them **asynchronously** on redirect:

```go
// modules/core/service/link_redirector.go:33-54
func (s *LinkRedirectorService) Redirect(ctx context.Context, shortCode string) (string, error) {
    id, err := lib.HexDecode(shortCode)
    if err != nil {
        return "", ErrNotFound
    }

    longURL, err := s.cacheRepo.Get(ctx, id)
    if err != nil {
        return "", ErrNotFound
    }

    now := time.Now()
    go func() {
        _ = s.analyticRepo.UpdateStat(context.Background(), id, now)
    }()

    return longURL, nil
}
```

```go
// modules/core/internal/repo/db/url_analytic_repo.go:60-68
func (r *PostgresURLAnalyticRepo) UpdateStat(ctx context.Context, urlID int64, now time.Time) error {
    _, err := r.db.ExecContext(
        ctx,
        `UPDATE url_analytics SET click_count = click_count + 1, last_accessed_at = $1 WHERE url_id = $2`,
        now,
        urlID,
    )
    return err
}
```

## Rationale

- **Low redirect latency**: Redirects are not blocked by database writes.
- **Atomic increments**: PostgreSQL handles concurrent updates safely.
- **Operational simplicity**: No extra queues or aggregation workers.

## Trade-offs

- **Eventual consistency**: Stats can lag behind redirects by a few seconds.
- **Possible drops**: If the async update fails, analytics may miss a click.
- **Write load**: Hot URLs can create write contention in PostgreSQL.

## Alternatives Considered

1. **Redis counters + periodic flush**
   - Fast increments but adds flush jobs, drift detection, and failure handling.
2. **Append-only click events**
   - Accurate and auditable, but requires aggregation pipelines and more storage.

## Consequences

- Redirects remain fast and consistent under burst load.
- Analytics accuracy is strong but not strictly guaranteed per-request.
- Future scaling may require batching or a dedicated analytics store.

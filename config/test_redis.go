package config

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

// TestRedis holds the test Redis client and cleanup function.
type TestRedis struct {
	Client  *redis.Client
	Cleanup func()
}

// SetupTestRedis creates a Redis client for testing and provides a cleanup function.
// It uses a test-specific database number to isolate tests.
func SetupTestRedis(t *testing.T) *TestRedis {
	t.Helper()

	// Get Redis URL from environment, default to localhost
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		dbIndex := testRedisDBIndex(t)
		redisURL = fmt.Sprintf("redis://localhost:6379/%d", dbIndex)
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("failed to parse redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	// Ping to verify connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("failed to ping redis: %v", err)
	}

	// Flush the test database to start clean
	if err := client.FlushDB(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("failed to flush redis test database: %v", err)
	}

	// Return TestRedis with cleanup function
	cleanup := func() {
		// Flush database before closing
		_ = client.FlushDB(context.Background())
		_ = client.Close()
	}

	return &TestRedis{
		Client:  client,
		Cleanup: cleanup,
	}
}

func testRedisDBIndex(t *testing.T) int {
	h := fnv.New32a()
	_, _ = fmt.Fprintf(h, "%s-%d", t.Name(), os.Getpid())
	return int(h.Sum32()%15) + 1
}

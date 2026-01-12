package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/nanda/doit/config"
	"github.com/nanda/doit/modules/core/lib"
	testutil "github.com/nanda/doit/test"
)

// TestContext holds shared state for integration tests.
// It provides access to the test server and stores response data
// that can be used across different step definitions.
type TestContext struct {
	T             *testing.T
	Server        *testutil.TestServer
	TestDB        *config.TestDB
	TestRedis     *config.TestRedis
	Client        *http.Client
	LastShortCode string
	LastResponse  *http.Response
	LastBody      []byte
}

// NewTestContext creates a new test context with default HTTP client.
func NewTestContext(t *testing.T) *TestContext {
	return &TestContext{
		T:      t,
		Client: &http.Client{},
	}
}

// Setup initializes the test database, Redis, and server.
func (tc *TestContext) Setup() error {
	tc.TestDB = config.SetupTestDB(tc.T)
	tc.TestRedis = config.SetupTestRedis(tc.T)
	tc.Server = testutil.NewTestServer(tc.TestDB.DB, tc.TestRedis.Client)
	return nil
}

// Cleanup releases resources after test completion.
func (tc *TestContext) Cleanup() {
	if tc.Server != nil {
		tc.Server.Close()
	}
	if tc.TestRedis != nil {
		tc.TestRedis.Cleanup()
	}
	if tc.TestDB != nil {
		tc.TestDB.Cleanup()
	}
}

// Reset clears state between scenarios while keeping server running.
func (tc *TestContext) Reset() {
	tc.LastShortCode = ""
	tc.LastResponse = nil
	tc.LastBody = nil
	// Reset client redirect behavior
	tc.Client.CheckRedirect = nil
}

// DisableRedirects configures the HTTP client to not follow redirects.
func (tc *TestContext) DisableRedirects() {
	tc.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

// ServerURL returns the base URL of the test server.
func (tc *TestContext) ServerURL() string {
	if tc.Server == nil {
		return ""
	}
	return tc.Server.URL()
}

// ExpireURL forces a URL to expire immediately in Redis for testing.
func (tc *TestContext) ExpireURL(shortCode string) error {
	if tc.TestRedis == nil {
		return nil
	}

	// Decode the short code to get the ID
	id, err := lib.HexDecode(shortCode)
	if err != nil {
		return fmt.Errorf("failed to decode short code: %w", err)
	}

	// Set the key to expire immediately (0 seconds TTL)
	key := fmt.Sprintf("url:%d", id)
	return tc.TestRedis.Client.Expire(context.Background(), key, 0).Err()
}

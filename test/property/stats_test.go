package property

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"github.com/nanda/doit/config"
	testutil "github.com/nanda/doit/test"
)

// TestProperty_StatisticsAlwaysAvailable verifies that statistics should
// always be available for created URLs.
func TestProperty_StatisticsAlwaysAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	server := testutil.NewTestServer(testDB.DB, testRedis.Client)
	defer server.Close()

	properties := gopter.NewProperties(DefaultTestParameters())

	properties.Property("statistics are always available", prop.ForAll(
		func(url string) bool {
			shortCode := CreateShortURL(t, server, url)

			// Get statistics
			resp, err := http.Get(server.URL() + "/stats/" + shortCode)
			if err != nil {
				t.Logf("Failed to get stats: %v", err)
				return false
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Logf("Expected 200 for stats, got %d", resp.StatusCode)
				return false
			}

			var stats struct {
				LongURL    string `json:"long_url"`
				ClickCount int64  `json:"click_count"`
				CreatedAt  string `json:"created_at"`
				ExpiresAt  string `json:"expires_at"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
				t.Logf("Failed to decode stats: %v", err)
				return false
			}

			// Verify stats contain expected data
			if stats.LongURL != url {
				t.Logf("Stats long_url mismatch: expected %s, got %s", url, stats.LongURL)
				return false
			}
			if stats.ClickCount < 0 {
				t.Logf("Click count should not be negative: %d", stats.ClickCount)
				return false
			}
			if stats.CreatedAt == "" {
				t.Log("created_at should not be empty")
				return false
			}
			if stats.ExpiresAt == "" {
				t.Log("expires_at should not be empty")
				return false
			}

			return true
		},
		GenValidURL(),
	))

	properties.TestingRun(t)
}

package property

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/nanda/doit/config"
	testutil "github.com/nanda/doit/test"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_ClickCountIncreasesMonotonically verifies that click count
// increases monotonically with each redirect.
func TestProperty_ClickCountIncreasesMonotonically(t *testing.T) {
	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	server := testutil.NewTestServer(testDB.DB, testRedis.Client)
	defer server.Close()

	client := NewNoRedirectClient()
	properties := gopter.NewProperties(ReducedTestParameters())

	properties.Property("click count increases monotonically", prop.ForAll(
		func(url string, numClicks int) bool {
			shortCode := CreateShortURL(t, server, url)

			for range numClicks {
				// Perform redirect (increments count) - don't follow redirects
				resp, err := client.Get(server.URL() + "/s/" + shortCode)
				if err != nil {
					t.Logf("Failed to redirect: %v", err)
					return false
				}
				_ = resp.Body.Close()
			}

			// Final count should equal number of clicks (eventual consistency)
			return waitForClickCountAtLeast(t, server, shortCode, int64(numClicks))
		},
		GenValidURL(),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

func waitForClickCountAtLeast(t *testing.T, server *testutil.TestServer, shortCode string, expected int64) bool {
	t.Helper()

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(server.URL() + "/stats/" + shortCode)
		if err != nil {
			t.Logf("Failed to get stats: %v", err)
			return false
		}

		var stats struct {
			ClickCount int64 `json:"click_count"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			_ = resp.Body.Close()
			t.Logf("Failed to decode stats: %v", err)
			return false
		}
		_ = resp.Body.Close()

		if stats.ClickCount >= expected {
			return true
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Logf("Click count did not reach expected value: expected=%d", expected)
	return false
}

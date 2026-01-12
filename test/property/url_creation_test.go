package property

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/nanda/doit/config"
	testutil "github.com/nanda/doit/test"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_CreatedURLsAreRetrievable verifies that any valid URL that is
// created should be retrievable via its short code.
func TestProperty_CreatedURLsAreRetrievable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	server := testutil.NewTestServer(testDB.DB, testRedis.Client)
	defer server.Close()

	client := NewNoRedirectClient()
	properties := gopter.NewProperties(DefaultTestParameters())

	properties.Property("created URLs are retrievable", prop.ForAll(
		func(url string) bool {
			// Create the short URL
			body := `{"long_url":"` + url + `"}`
			resp, err := http.Post(server.URL()+"/s", "application/json", strings.NewReader(body))
			if err != nil {
				t.Logf("Failed to create short URL: %v", err)
				return false
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Logf("Expected 200, got %d", resp.StatusCode)
				return false
			}

			var result struct {
				ShortCode string `json:"short_code"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Logf("Failed to decode response: %v", err)
				return false
			}

			// Verify the URL is retrievable
			resp2, err := client.Get(server.URL() + "/s/" + result.ShortCode)
			if err != nil {
				t.Logf("Failed to retrieve short URL: %v", err)
				return false
			}
			defer func() { _ = resp2.Body.Close() }()

			if resp2.StatusCode != http.StatusFound {
				t.Logf("Expected 302, got %d", resp2.StatusCode)
				return false
			}

			location := resp2.Header.Get("Location")
			return location == url
		},
		GenValidURL(),
	))

	properties.TestingRun(t)
}

// TestProperty_DifferentURLsProduceDifferentShortCodes verifies that URLs with
// different long URLs should produce different short codes.
func TestProperty_DifferentURLsProduceDifferentShortCodes(t *testing.T) {
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

	properties.Property("different URLs produce different short codes", prop.ForAll(
		func(url1, url2 string) bool {
			if url1 == url2 {
				return true // Skip if URLs are the same
			}

			shortCode1 := CreateShortURL(t, server, url1)
			shortCode2 := CreateShortURL(t, server, url2)

			return shortCode1 != shortCode2
		},
		GenValidURL(),
		GenValidURL(),
	))

	properties.TestingRun(t)
}

// TestProperty_ShortCodesAreURLSafe verifies that short codes contain only
// valid custom hex characters. The hex encoding maps: 0->g, 1->h, so valid
// chars are: 2-9, a-f, g, h.
func TestProperty_ShortCodesAreURLSafe(t *testing.T) {
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

	properties.Property("short codes contain only valid hex characters", prop.ForAll(
		func(url string) bool {
			shortCode := CreateShortURL(t, server, url)

			// Verify short code contains only custom hex characters
			for _, c := range shortCode {
				isDigit := c >= '2' && c <= '9'
				isHexLetter := c >= 'a' && c <= 'h'
				if !isDigit && !isHexLetter {
					t.Logf("Short code contains invalid character: %c in %s", c, shortCode)
					return false
				}
			}

			// Verify short code is not empty
			return len(shortCode) > 0
		},
		GenValidURL(),
	))

	properties.TestingRun(t)
}

// TestProperty_ShortCodeLengthIsReasonable verifies that short code length
// should be consistent and reasonable (between 1 and 16 characters).
func TestProperty_ShortCodeLengthIsReasonable(t *testing.T) {
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

	properties.Property("short code length is reasonable", prop.ForAll(
		func(url string) bool {
			shortCode := CreateShortURL(t, server, url)

			// Short codes should be between 1 and 16 characters (reasonable for hex encoding)
			return len(shortCode) >= 1 && len(shortCode) <= 16
		},
		GenValidURL(),
	))

	properties.TestingRun(t)
}

package property

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	testutil "github.com/nanda/doit/test"
)

// genLowerAlphaString generates a lowercase alphabetic string with length constraints.
func genLowerAlphaString(minLen, maxLen int) gopter.Gen {
	return gen.SliceOfN(maxLen, gen.AlphaLowerChar()).
		SuchThat(func(chars []rune) bool {
			return len(chars) >= minLen
		}).
		Map(func(chars []rune) string {
			return string(chars)
		})
}

// genAlphaNumString generates an alphanumeric string with length constraints.
func genAlphaNumString(minLen, maxLen int) gopter.Gen {
	return gen.SliceOfN(maxLen, gen.AlphaNumChar()).
		SuchThat(func(chars []rune) bool {
			return len(chars) >= minLen
		}).
		Map(func(chars []rune) string {
			s := string(chars)
			// Convert to lowercase for consistency
			return strings.ToLower(s)
		})
}

// GenValidURL generates a complete valid URL using gopter generators.
func GenValidURL() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("http", "https"),
		genLowerAlphaString(3, 15),
		gen.OneConstOf("com", "org", "net", "io", "dev"),
		gen.SliceOfN(3, genAlphaNumString(1, 10)),
	).Map(func(values []interface{}) string {
		scheme := values[0].(string)
		domain := values[1].(string)
		tld := values[2].(string)
		pathSegments := values[3].([]string)

		// Ensure domain is lowercase
		domain = strings.ToLower(domain)

		url := scheme + "://" + domain + "." + tld
		if len(pathSegments) > 0 {
			// Filter out empty segments
			var validSegments []string
			for _, seg := range pathSegments {
				if seg != "" {
					validSegments = append(validSegments, seg)
				}
			}
			if len(validSegments) > 0 {
				url += "/" + strings.Join(validSegments, "/")
			}
		}
		return url
	}).SuchThat(func(url string) bool {
		// Ensure URL is valid (has proper domain)
		return strings.Contains(url, "://") && len(url) > 10
	})
}

// CreateShortURL creates a short URL and returns the short code.
func CreateShortURL(t *testing.T, server *testutil.TestServer, url string) string {
	body := `{"long_url":"` + url + `"}`
	resp, err := http.Post(server.URL()+"/s", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var result struct {
		ShortCode string `json:"short_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return result.ShortCode
}

// NewNoRedirectClient creates an HTTP client that doesn't follow redirects.
func NewNoRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// DefaultTestParameters returns the default gopter test parameters.
func DefaultTestParameters() *gopter.TestParameters {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	return parameters
}

// ReducedTestParameters returns test parameters with fewer iterations.
// Useful for tests that involve multiple HTTP calls per iteration.
func ReducedTestParameters() *gopter.TestParameters {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	return parameters
}

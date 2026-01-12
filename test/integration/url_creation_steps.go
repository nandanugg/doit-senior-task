package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
)

// URLCreationSteps handles step definitions for URL creation scenarios.
type URLCreationSteps struct {
	ctx *TestContext
}

// NewURLCreationSteps creates a new instance with the given test context.
func NewURLCreationSteps(ctx *TestContext) *URLCreationSteps {
	return &URLCreationSteps{ctx: ctx}
}

// Register adds all URL creation step definitions to the scenario context.
func (s *URLCreationSteps) Register(sc *godog.ScenarioContext) {
	sc.Step(`^the service is running$`, s.theServiceIsRunning)
	sc.Step(`^I create a short URL for "([^"]*)"$`, s.iCreateAShortURLFor)
	sc.Step(`^I have created a short URL for "([^"]*)"$`, s.iHaveCreatedAShortURLFor)
	sc.Step(`^I should receive a short code$`, s.iShouldReceiveAShortCode)
	sc.Step(`^the short code should be valid$`, s.theShortCodeShouldBeValid)
}

func (s *URLCreationSteps) theServiceIsRunning() error {
	if s.ctx.Server == nil {
		return fmt.Errorf("server is not running")
	}
	return nil
}

func (s *URLCreationSteps) iCreateAShortURLFor(longURL string) error {
	body := fmt.Sprintf(`{"long_url":"%s"}`, longURL)
	resp, err := s.ctx.Client.Post(
		s.ctx.ServerURL()+"/s",
		"application/json",
		bytes.NewReader([]byte(body)),
	)
	if err != nil {
		return fmt.Errorf("failed to create short URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Save response for header checks
	s.ctx.LastResponse = resp

	// Read body for parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	s.ctx.LastBody = bodyBytes

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result struct {
		ShortCode string `json:"short_code"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	s.ctx.LastShortCode = result.ShortCode
	return nil
}

func (s *URLCreationSteps) iHaveCreatedAShortURLFor(longURL string) error {
	return s.iCreateAShortURLFor(longURL)
}

func (s *URLCreationSteps) iShouldReceiveAShortCode() error {
	if s.ctx.LastShortCode == "" {
		return fmt.Errorf("no short code received")
	}
	return nil
}

func (s *URLCreationSteps) theShortCodeShouldBeValid() error {
	// The hex encoding maps: 0->g, 1->h
	// Valid characters are: 2-9, a-f, g, h
	for _, c := range s.ctx.LastShortCode {
		isDigit := c >= '2' && c <= '9'
		isHexLetter := c >= 'a' && c <= 'h'
		if !isDigit && !isHexLetter {
			return fmt.Errorf("short code contains invalid character: %c in %s", c, s.ctx.LastShortCode)
		}
	}
	return nil
}

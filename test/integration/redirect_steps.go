package integration

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cucumber/godog"
)

// RedirectSteps handles step definitions for URL redirect scenarios.
type RedirectSteps struct {
	ctx *TestContext
}

// NewRedirectSteps creates a new instance with the given test context.
func NewRedirectSteps(ctx *TestContext) *RedirectSteps {
	return &RedirectSteps{ctx: ctx}
}

// Register adds all redirect step definitions to the scenario context.
func (s *RedirectSteps) Register(sc *godog.ScenarioContext) {
	sc.Step(`^I visit the short URL$`, s.iVisitTheShortURL)
	sc.Step(`^I visit the short URL (\d+) times$`, s.iVisitTheShortURLTimes)
	sc.Step(`^I should be redirected to "([^"]*)"$`, s.iShouldBeRedirectedTo)
}

func (s *RedirectSteps) iVisitTheShortURL() error {
	s.ctx.DisableRedirects()

	resp, err := s.ctx.Client.Get(s.ctx.ServerURL() + "/s/" + s.ctx.LastShortCode)
	if err != nil {
		return fmt.Errorf("failed to visit short URL: %w", err)
	}

	s.ctx.LastResponse = resp
	return nil
}

func (s *RedirectSteps) iVisitTheShortURLTimes(count int) error {
	s.ctx.DisableRedirects()

	var lastResp *http.Response
	for i := range count {
		resp, err := s.ctx.Client.Get(s.ctx.ServerURL() + "/s/" + s.ctx.LastShortCode)
		if err != nil {
			return fmt.Errorf("failed to visit short URL (attempt %d): %w", i+1, err)
		}

		if i == count-1 {
			// Save last response for header checks
			lastResp = resp
		} else {
			_ = resp.Body.Close()
		}
	}

	// Wait for async analytics updates to complete
	time.Sleep(100 * time.Millisecond)

	// Save the last response
	s.ctx.LastResponse = lastResp
	return nil
}

func (s *RedirectSteps) iShouldBeRedirectedTo(expectedURL string) error {
	if s.ctx.LastResponse == nil {
		return fmt.Errorf("no response available")
	}

	if s.ctx.LastResponse.StatusCode != http.StatusFound {
		return fmt.Errorf("expected status 302, got %d", s.ctx.LastResponse.StatusCode)
	}

	location := s.ctx.LastResponse.Header.Get("Location")
	if location != expectedURL {
		return fmt.Errorf("expected location %s, got %s", expectedURL, location)
	}

	_ = s.ctx.LastResponse.Body.Close()
	return nil
}

package integration

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cucumber/godog"
)

// StatsSteps handles step definitions for URL statistics scenarios.
type StatsSteps struct {
	ctx *TestContext
}

// NewStatsSteps creates a new instance with the given test context.
func NewStatsSteps(ctx *TestContext) *StatsSteps {
	return &StatsSteps{ctx: ctx}
}

// Register adds all statistics step definitions to the scenario context.
func (s *StatsSteps) Register(sc *godog.ScenarioContext) {
	sc.Step(`^I check the statistics$`, s.iCheckTheStatistics)
	sc.Step(`^the click count should be (\d+)$`, s.theClickCountShouldBe)
	sc.Step(`^I should see the long URL "([^"]*)"$`, s.iShouldSeeTheLongURL)
	sc.Step(`^I should see the click count$`, s.iShouldSeeTheClickCount)
	sc.Step(`^I should see the creation time$`, s.iShouldSeeTheCreationTime)
	sc.Step(`^I should see the expiration time$`, s.iShouldSeeTheExpirationTime)
}

func (s *StatsSteps) iCheckTheStatistics() error {
	resp, err := s.ctx.Client.Get(s.ctx.ServerURL() + "/stats/" + s.ctx.LastShortCode)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return fmt.Errorf("failed to read response body: %w", err)
	}
	_ = resp.Body.Close()

	s.ctx.LastBody = body
	s.ctx.LastResponse = resp
	return nil
}

func (s *StatsSteps) theClickCountShouldBe(expected int) error {
	var stats struct {
		ClickCount int64 `json:"click_count"`
	}
	if err := json.Unmarshal(s.ctx.LastBody, &stats); err != nil {
		return fmt.Errorf("failed to parse statistics: %w", err)
	}

	if stats.ClickCount != int64(expected) {
		return fmt.Errorf("expected click count %d, got %d", expected, stats.ClickCount)
	}
	return nil
}

func (s *StatsSteps) iShouldSeeTheLongURL(expectedURL string) error {
	var stats struct {
		LongURL string `json:"long_url"`
	}
	if err := json.Unmarshal(s.ctx.LastBody, &stats); err != nil {
		return fmt.Errorf("failed to parse statistics: %w", err)
	}

	if stats.LongURL != expectedURL {
		return fmt.Errorf("expected long URL %s, got %s", expectedURL, stats.LongURL)
	}
	return nil
}

func (s *StatsSteps) iShouldSeeTheClickCount() error {
	var stats struct {
		ClickCount int64 `json:"click_count"`
	}
	if err := json.Unmarshal(s.ctx.LastBody, &stats); err != nil {
		return fmt.Errorf("failed to parse statistics: %w", err)
	}

	if stats.ClickCount < 0 {
		return fmt.Errorf("click count is negative: %d", stats.ClickCount)
	}
	return nil
}

func (s *StatsSteps) iShouldSeeTheCreationTime() error {
	var stats struct {
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(s.ctx.LastBody, &stats); err != nil {
		return fmt.Errorf("failed to parse statistics: %w", err)
	}

	if stats.CreatedAt == "" {
		return fmt.Errorf("created_at is empty")
	}
	return nil
}

func (s *StatsSteps) iShouldSeeTheExpirationTime() error {
	var stats struct {
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.Unmarshal(s.ctx.LastBody, &stats); err != nil {
		return fmt.Errorf("failed to parse statistics: %w", err)
	}

	if stats.ExpiresAt == "" {
		return fmt.Errorf("expires_at is empty")
	}
	return nil
}

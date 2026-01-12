package integration

import (
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

// CommonSteps contains step definitions that are shared across scenarios.
type CommonSteps struct {
	ctx *TestContext
}

// NewCommonSteps creates a new CommonSteps instance.
func NewCommonSteps(ctx *TestContext) *CommonSteps {
	return &CommonSteps{ctx: ctx}
}

// Register adds all common step definitions to the scenario context.
func (s *CommonSteps) Register(sc *godog.ScenarioContext) {
	sc.Step(`^the response should have processing time header$`, s.theResponseShouldHaveProcessingTimeHeader)
}

// theResponseShouldHaveProcessingTimeHeader verifies that the X-Processing-Time-Micros
// header is present in the response and contains a valid integer value.
func (s *CommonSteps) theResponseShouldHaveProcessingTimeHeader() error {
	if s.ctx.LastResponse == nil {
		return fmt.Errorf("no response available")
	}

	header := s.ctx.LastResponse.Header.Get("X-Processing-Time-Micros")
	if header == "" {
		return fmt.Errorf("X-Processing-Time-Micros header is missing")
	}

	// Verify it's a valid integer
	value, err := strconv.ParseInt(header, 10, 64)
	if err != nil {
		return fmt.Errorf("X-Processing-Time-Micros header is not a valid integer: %s", header)
	}

	// Verify it's a positive number
	if value < 0 {
		return fmt.Errorf("X-Processing-Time-Micros should be positive, got: %d", value)
	}

	return nil
}

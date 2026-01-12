package integration

import (
	"errors"

	"github.com/cucumber/godog"
)

var ErrNoShortCode = errors.New("no short code available")

// RegisterExpirationSteps registers Gherkin steps for URL expiration testing.
func RegisterExpirationSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	sc.Step(`^I expire the last created short URL$`, ctx.iExpireTheLastCreatedShortURL)
}

func (tc *TestContext) iExpireTheLastCreatedShortURL() error {
	if tc.LastShortCode == "" {
		return ErrNoShortCode
	}

	return tc.ExpireURL(tc.LastShortCode)
}

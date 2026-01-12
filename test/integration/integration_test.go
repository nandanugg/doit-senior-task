package integration

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
)

// createScenarioInitializer returns a scenario initializer with access to testing.T.
func createScenarioInitializer(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		ctx := NewTestContext(t)

		// Register lifecycle hooks
		sc.Before(func(c context.Context, scenario *godog.Scenario) (context.Context, error) {
			if err := ctx.Setup(); err != nil {
				return c, err
			}
			return c, nil
		})

		sc.After(func(c context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
			ctx.Cleanup()
			return c, nil
		})

		// Register step definitions from each domain
		NewURLCreationSteps(ctx).Register(sc)
		NewRedirectSteps(ctx).Register(sc)
		NewStatsSteps(ctx).Register(sc)
		NewCommonSteps(ctx).Register(sc)
		RegisterExpirationSteps(sc, ctx)
	}
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: createScenarioInitializer(t),
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

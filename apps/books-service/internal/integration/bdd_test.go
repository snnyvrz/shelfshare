//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
)

func TestBDD(t *testing.T) {
	suite := godog.TestSuite{
		Name: "lending",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			state := &lendingScenario{}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				return ctx, state.reset()
			})
			state.register(sc)
		},
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"features"},
			Strict: true, Concurrency: 1, TestingT: t,
		},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("BDD suite exited with status %d", status)
	}
}

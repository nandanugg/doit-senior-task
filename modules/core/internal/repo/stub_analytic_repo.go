package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/nanda/doit/modules/core/entity"
)

// StubURLAnalyticRepo is a temporary stub implementation until database is added.
type StubURLAnalyticRepo struct{}

func NewStubURLAnalyticRepo() *StubURLAnalyticRepo {
	return &StubURLAnalyticRepo{}
}

func (s *StubURLAnalyticRepo) Create(ctx context.Context, analytic *entity.URLAnalytic) (int64, error) {
	// Stub: analytics not persisted yet
	return 0, nil
}

func (s *StubURLAnalyticRepo) GetByURLID(ctx context.Context, urlID int64) (*entity.URLAnalytic, error) {
	// Stub: analytics not available yet
	return nil, fmt.Errorf("analytics not available without database")
}

func (s *StubURLAnalyticRepo) UpdateStat(ctx context.Context, urlID int64, now time.Time) error {
	// Stub: analytics update ignored for now
	return nil
}

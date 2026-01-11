package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nanda/doit/modules/core/entity"
	"github.com/nanda/doit/modules/core/lib"
)

type LinkAnalyzer interface {
	Analyze(ctx context.Context, shortCode string) (*entity.URLAnalytic, error)
}

type LinkAnalyzerService struct {
	analyticRepo URLAnalyticRepo
}

func NewLinkAnalyzerService(analyticRepo URLAnalyticRepo) *LinkAnalyzerService {
	return &LinkAnalyzerService{
		analyticRepo: analyticRepo,
	}
}

func (s *LinkAnalyzerService) Analyze(ctx context.Context, shortCode string) (*entity.URLAnalytic, error) {
	id, err := lib.HexDecode(shortCode)
	if err != nil {
		return nil, ErrNotFound
	}

	analytic, err := s.analyticRepo.GetByURLID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return analytic, nil
}

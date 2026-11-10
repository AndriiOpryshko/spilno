package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Reviews serves the оператор's queue and review page and records checks of extracts.
type Reviews struct {
	repo ports.ReviewRepository
}

func NewReviews(repo ports.ReviewRepository) *Reviews { return &Reviews{repo: repo} }

// ReviewView is a review page: the detail and the advisory warnings of the latest check.
type ReviewView struct {
	ports.ReviewDetail
	Warnings []domain.ReviewWarning
}

func (r *Reviews) List(ctx context.Context, state string, after *ports.AssociationCursor, limit int) ([]ports.ReviewApplication, error) {
	return r.repo.List(ctx, state, after, limit)
}

func (r *Reviews) Get(ctx context.Context, id uuid.UUID) (ReviewView, error) {
	d, err := r.repo.Get(ctx, id)
	if err != nil {
		return ReviewView{}, err
	}
	return ReviewView{ReviewDetail: d, Warnings: warningsOf(d)}, nil
}

func (r *Reviews) Extract(ctx context.Context, applicationID, filingID uuid.UUID) (ports.ReviewExtract, error) {
	return r.repo.Extract(ctx, applicationID, filingID)
}

func (r *Reviews) RecordCheck(ctx context.Context, c ports.NewCheck) (ports.ReviewCheck, error) {
	return r.repo.RecordCheck(ctx, c)
}

// warningsOf checks the latest filing against its latest check; none blocks anything.
func warningsOf(d ports.ReviewDetail) []domain.ReviewWarning {
	if len(d.Filings) == 0 {
		return nil
	}
	f := d.Filings[0]
	var latest *ports.ReviewCheck
	for i := range d.Checks {
		if d.Checks[i].FilingID == f.ID {
			latest = &d.Checks[i]
		}
	}
	if latest == nil {
		return nil
	}
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		loc = time.UTC
	}
	y, m, day := f.CreatedAt.In(loc).Date()
	return domain.ReviewWarnings(f.ApplicantFullName, domain.ExtractCheck{
		SignerName: latest.SignerFullName, HeadFullName: latest.HeadFullName, ExtractDated: latest.ExtractDatedOn,
	}, time.Date(y, m, day, 0, 0, 0, 0, time.UTC))
}

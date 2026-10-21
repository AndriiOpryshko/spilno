package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
)

// fakeRepo serves List from a slice already in listing order and records the
// calls it gets.
type fakeRepo struct {
	ports.ContractRepository
	rows      []domain.Contract
	gotAfter  *ports.ContractKey
	gotLimit  int
	createdOK bool
}

func (f *fakeRepo) Create(_ context.Context, t domain.ContractTerms) (domain.Contract, error) {
	f.createdOK = true
	return domain.Contract{ID: uuid.New(), ContractTerms: t}, nil
}

func (f *fakeRepo) List(_ context.Context, after *ports.ContractKey, limit int) ([]domain.Contract, error) {
	f.gotAfter, f.gotLimit = after, limit
	start := 0
	if after != nil {
		for i, c := range f.rows {
			if c.ID == after.ID {
				start = i + 1
			}
		}
	}
	return f.rows[start:min(start+limit, len(f.rows))], nil
}

func rows(n int) []domain.Contract {
	base := time.Date(2026, 10, 21, 12, 0, 0, 123456000, time.UTC)
	out := make([]domain.Contract, n)
	for i := range out {
		out[i] = domain.Contract{ID: uuid.New(), CreatedAt: base.Add(-time.Duration(i) * time.Minute)}
	}
	return out
}

func TestCreateRejectsInvalidTermsBeforeStoring(t *testing.T) {
	repo := &fakeRepo{}
	_, err := app.NewContractService(repo).Create(context.Background(), domain.ContractTerms{})
	var v *domain.ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
	if repo.createdOK {
		t.Error("invalid terms reached the repository")
	}
}

func TestListWalksAllPages(t *testing.T) {
	repo := &fakeRepo{rows: rows(5)}
	svc := app.NewContractService(repo)

	var seen []uuid.UUID
	cursor := ""
	for range 10 {
		page, err := svc.List(context.Background(), 2, cursor)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range page.Items {
			seen = append(seen, c.ID)
		}
		if page.NextCursor == "" {
			break
		}
		if repo.gotLimit != 3 {
			t.Errorf("repo limit = %d, want page size + 1", repo.gotLimit)
		}
		cursor = page.NextCursor
	}
	if len(seen) != 5 {
		t.Fatalf("saw %d contracts, want 5", len(seen))
	}
	for i, c := range repo.rows {
		if seen[i] != c.ID {
			t.Errorf("item %d = %v, want %v", i, seen[i], c.ID)
		}
	}
}

func TestListCursorKeepsMicroseconds(t *testing.T) {
	repo := &fakeRepo{rows: rows(2)}
	svc := app.NewContractService(repo)
	page, err := svc.List(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.List(context.Background(), 1, page.NextCursor); err != nil {
		t.Fatal(err)
	}
	want := repo.rows[0]
	if repo.gotAfter == nil || repo.gotAfter.ID != want.ID || !repo.gotAfter.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("after = %+v, want {%v %v}", repo.gotAfter, want.CreatedAt, want.ID)
	}
}

func TestListLimitBounds(t *testing.T) {
	tests := []struct{ limit, wantRepo int }{
		{0, app.DefaultPageSize + 1},
		{-3, app.DefaultPageSize + 1},
		{1, 2},
		{app.MaxPageSize, app.MaxPageSize + 1},
		{app.MaxPageSize + 1, app.MaxPageSize + 1},
	}
	for _, tt := range tests {
		repo := &fakeRepo{}
		if _, err := app.NewContractService(repo).List(context.Background(), tt.limit, ""); err != nil {
			t.Fatal(err)
		}
		if repo.gotLimit != tt.wantRepo {
			t.Errorf("List(limit=%d): repo limit = %d, want %d", tt.limit, repo.gotLimit, tt.wantRepo)
		}
	}
}

func TestListRejectsForeignCursor(t *testing.T) {
	svc := app.NewContractService(&fakeRepo{})
	for _, c := range []string{"not base64!", "bm9jb2xvbg", "eDpub3QtYS11dWlk"} {
		if _, err := svc.List(context.Background(), 10, c); !errors.Is(err, app.ErrInvalidCursor) {
			t.Errorf("List(cursor=%q) err = %v, want ErrInvalidCursor", c, err)
		}
	}
}

package app

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Rejections records the оператор's rejection of an application.
type Rejections struct {
	repo ports.RejectionRepository
}

func NewRejections(repo ports.RejectionRepository) *Rejections { return &Rejections{repo: repo} }

// Reject refuses an empty reason, a decided application and the оператор's own one with
// *domain.CodedError.
func (r *Rejections) Reject(ctx context.Context, in ports.NewRejection) (ports.ReviewDecision, error) {
	return r.repo.Reject(ctx, in, func(a domain.Application) (domain.Application, error) {
		return a.Reject(in.OperatorAccountID.String(), in.Reason)
	})
}

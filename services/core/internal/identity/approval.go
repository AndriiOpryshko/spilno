package identity

import (
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// ApprovalConfig is what approving a заявка needs from outside identity: the roster step, which
// runs inside the approval's transaction (cmd/api passes governance.Module.SeatHead).
type ApprovalConfig struct {
	SeatHead ports.SeatHeadFunc
}

// EnableApprovals turns approveApplication on.
func (m *Module) EnableApprovals(cfg ApprovalConfig) {
	m.handler.EnableApprovals(app.NewApprovals(postgres.NewApprovals(m.pool), cfg.SeatHead))
}

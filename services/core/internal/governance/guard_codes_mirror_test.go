package governance_test

import (
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
)

// Governance may not import identity, so it mirrors the two guard codes a download refuses with
// (review of 2026-11-17, N05). The mirrors must stay equal to the codes the identity module exports.
func TestGuardCodesMirrorTheIdentityModule(t *testing.T) {
	for name, pair := range map[string][2]string{
		"association_disabled": {domain.CodeAssociationDisabled, identity.CodeAssociationDisabled},
		"not_current_member":   {domain.CodeNotCurrentMember, identity.CodeNotCurrentMember},
	} {
		if pair[0] != pair[1] {
			t.Errorf("%s: governance mirrors %q, identity exports %q", name, pair[0], pair[1])
		}
	}
}

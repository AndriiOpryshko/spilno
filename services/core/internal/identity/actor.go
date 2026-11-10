package identity

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
)

// Actor is the signed-in person as the guard resolved them.
type Actor = domain.Actor

type actorKey struct{}

func withActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, a)
}

// ActorFrom returns the actor the guard put in ctx.
func ActorFrom(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(actorKey{}).(Actor)
	return a, ok
}

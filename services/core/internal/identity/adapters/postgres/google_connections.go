// Package postgres implements the identity context's ports on Postgres.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

type GoogleConnections struct {
	pool *pgxpool.Pool
}

var _ ports.GoogleConnectionRepository = (*GoogleConnections)(nil)

func NewGoogleConnections(pool *pgxpool.Pool) *GoogleConnections {
	return &GoogleConnections{pool: pool}
}

func (s *GoogleConnections) Replace(ctx context.Context, c ports.GoogleConnection) error {
	return db.InAssociation(ctx, s.pool, c.AssociationID, "", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		q := sqlcgen.New(tx)
		association := pgtype.UUID{Bytes: c.AssociationID, Valid: true}
		if err := q.DeleteGoogleConnection(ctx, association); err != nil {
			return err
		}
		return q.InsertGoogleConnection(ctx, sqlcgen.InsertGoogleConnectionParams{
			ID:                     pgtype.UUID{Bytes: c.ID, Valid: true},
			AssociationID:          association,
			GoogleAccountEmail:     c.AccountEmail,
			RefreshTokenCiphertext: c.RefreshTokenCiphertext,
			RefreshTokenNonce:      c.RefreshTokenNonce,
			EncryptionKeyID:        c.EncryptionKeyID,
			Scopes:                 c.Scopes,
			DriveRootFolderID:      pgtype.Text{String: c.DriveRootFolderID, Valid: c.DriveRootFolderID != ""},
			AuthorizedByAccountID:  pgtype.UUID{Bytes: c.AuthorizedBy, Valid: true},
		})
	})
}

func (s *GoogleConnections) Get(ctx context.Context, associationID uuid.UUID) (ports.GoogleConnection, error) {
	var row sqlcgen.GoogleConnection
	err := db.InAssociation(ctx, s.pool, associationID, "", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		row, err = sqlcgen.New(tx).GetGoogleConnection(ctx, pgtype.UUID{Bytes: associationID, Valid: true})
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.GoogleConnection{}, ports.ErrGoogleConnectionNotFound
	}
	if err != nil {
		return ports.GoogleConnection{}, err
	}
	return ports.GoogleConnection{
		ID:                     row.ID.Bytes,
		AssociationID:          row.AssociationID.Bytes,
		AccountEmail:           row.GoogleAccountEmail,
		RefreshTokenCiphertext: row.RefreshTokenCiphertext,
		RefreshTokenNonce:      row.RefreshTokenNonce,
		EncryptionKeyID:        row.EncryptionKeyID,
		Scopes:                 row.Scopes,
		DriveRootFolderID:      row.DriveRootFolderID.String,
		AuthorizedBy:           row.AuthorizedByAccountID.Bytes,
	}, nil
}

package txmanager

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type accessModeKey struct{}

func (t *Transactor) WithReadOnly(ctx context.Context) context.Context {
	return injectAccessMode(ctx, AccessModeReadOnly)
}

func injectAccessMode(ctx context.Context, mode pgx.TxAccessMode) context.Context {
	return context.WithValue(ctx, accessModeKey{}, mode)
}

func extractAccessMode(ctx context.Context) (pgx.TxAccessMode, bool) {
	mode, ok := ctx.Value(accessModeKey{}).(pgx.TxAccessMode)
	return mode, ok
}
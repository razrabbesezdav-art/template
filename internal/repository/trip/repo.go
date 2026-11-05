package trip

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/razrabbesezdav-art/template/internal/infra/postgres/transaction"
)

type Querier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Repo struct {
	db           Querier
	builder      sq.StatementBuilderType
	queryTimeout time.Duration
}

func NewRepo(db Querier, queryTimeout time.Duration) *Repo {
	return &Repo{
		db:           db,
		builder:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		queryTimeout: queryTimeout,
	}
}

func (r *Repo) queryer(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}
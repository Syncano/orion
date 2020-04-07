package query

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// Lock performs query with locking of rows for update.
func Lock(q *orm.Query) error {
	return q.For("UPDATE OF ?TableAlias").Select()
}

// RequireOne panics on unexpected errors.
func RequireOne(err error) error {
	if err == nil || err == pg.ErrNoRows {
		return err
	}

	panic(err)
}

// CountEstimate returns count estimate results for given query.
func CountEstimate(ctx context.Context, db orm.DB, q orm.QueryAppender, threshold int) (int, error) {
	type res struct {
		ObjectsCount int
	}

	r := res{}

	qq, err := q.AppendQuery(db.Formatter(), nil)
	if err != nil {
		return 0, err
	}

	return r.ObjectsCount, db.ModelContext(ctx).
		ColumnExpr(`?schema.count_estimate(?, ?) AS "objects_count"`, string(qq), threshold).
		Select(&r)
}

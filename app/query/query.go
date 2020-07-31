package query

import (
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/rediscache"
)

type Factory struct {
	db *database.DB
	c  *rediscache.Cache
}

func NewFactory(db *database.DB, c *rediscache.Cache) *Factory {
	return &Factory{
		db: db,
		c:  c,
	}
}

func (q *Factory) Database() *database.DB {
	return q.db
}

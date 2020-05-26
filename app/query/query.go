package query

import (
	"github.com/Syncano/pkg-go/rediscache"
	"github.com/Syncano/pkg-go/storage"
)

type Factory struct {
	dbase *storage.Database
	c     *rediscache.Cache
}

func NewFactory(dbase *storage.Database, c *rediscache.Cache) *Factory {
	return &Factory{
		dbase: dbase,
		c:     c,
	}
}

func (q *Factory) Database() *storage.Database {
	return q.dbase
}

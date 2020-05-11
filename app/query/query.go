package query

import (
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

type Factory struct {
	dbase *storage.Database
	c     *cache.Cache
}

func NewFactory(dbase *storage.Database, c *cache.Cache) *Factory {
	return &Factory{
		dbase: dbase,
		c:     c,
	}
}

func (q *Factory) Database() *storage.Database {
	return q.dbase
}

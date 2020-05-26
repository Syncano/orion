package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/storage"
)

// ClassManager represents Class manager.
type ClassManager struct {
	*LiveManager
}

// NewClassManager creates and returns new Class manager.
func (q *Factory) NewClassManager(c storage.DBContext) *ClassManager {
	return &ClassManager{LiveManager: q.NewLiveTenantManager(c)}
}

// OneByName outputs object filtered by name.
func (m *ClassManager) OneByName(o *models.Class) error {
	o.Name = strings.ToLower(o.Name)
	if o.Name == "user" {
		o.Name = models.UserClassName
	}

	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).
				Where("name = ?", o.Name).
				Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *ClassManager) WithAccessQ(o interface{}) *orm.Query {
	q := m.Query(o).
		Where("visible IS TRUE").
		Column("class.*").
		ColumnExpr(`?schema.count_estimate('SELECT id FROM ?schema.data_dataobject WHERE _klass_id=' || "class"."id", ?) AS "objects_count"`,
			settings.API.DataObjectEstimateThreshold)

	return q
}

// WithAccessByNameQ returns one object that entity has access to filtered by name.
func (m *ClassManager) WithAccessByNameQ(o *models.Class) *orm.Query {
	return m.WithAccessQ(o).
		Where("?TableAlias.name = ?", o.Name)
}

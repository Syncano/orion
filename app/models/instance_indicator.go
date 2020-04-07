package models

import (
	"fmt"
)

// InstanceIndicatorType enum.
const (
	InstanceIndicatorTypeSchedulesCount int = iota
	InstanceIndicatorTypeStorageSize
)

// InstanceIndicatorType to string map.
var InstanceIndicatorType = map[int]string{
	InstanceIndicatorTypeSchedulesCount: "schedules_count",
	InstanceIndicatorTypeStorageSize:    "storage_size",
}

// InstanceIndicator represents InstanceIndicator model.
type InstanceIndicator struct {
	tableName struct{} `pg:"instances_instanceindicator"` // nolint

	ID         int
	Type       int
	Value      int
	InstanceID int
	Instance   *Instance
}

func (m *InstanceIndicator) String() string {
	return fmt.Sprintf("InstanceIndicator<ID=%d Type=%d Value=%d Instance=%d>", m.ID, m.Type, m.Value, m.InstanceID)
}

// VerboseName returns verbose name for model.
func (m *InstanceIndicator) VerboseName() string {
	return "Instance Indicator"
}

// TypeString returns verbose value for type.
func (m *InstanceIndicator) TypeString() string {
	return InstanceIndicatorType[m.Type]
}

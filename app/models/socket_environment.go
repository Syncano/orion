package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// SocketEnvironmentStatus enum.
const (
	SocketEnvironmentStatusError int = iota - 1 // start from -1
	SocketEnvironmentStatusProcessing
	SocketEnvironmentStatusOK
)

// SocketEnvironmentStatus to string map.
var SocketEnvironmentStatus = map[int]string{
	SocketEnvironmentStatusError:      "error",
	SocketEnvironmentStatusProcessing: "processing",
	SocketEnvironmentStatusOK:         "ok",
}

// SocketEnvironment represents socket environment model.
type SocketEnvironment struct {
	tableName struct{} `sql:"?schema.sockets_socketenvironment" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID          int
	Name        string
	Metadata    JSON
	Description string
	Status      int
	StatusInfo  string
	CreatedAt   Time
	UpdatedAt   Time
	Checksum    string

	ZipFile string
	FsFile  string
}

func (m *SocketEnvironment) String() string {
	return fmt.Sprintf("SocketEnvironment<ID=%d, Name=%q>", m.ID, m.Name)
}

// Hash ...
func (m *SocketEnvironment) Hash() string {
	return fmt.Sprintf("E:%s", m.Checksum)
}

// URL ...
func (m *SocketEnvironment) URL() string {
	return buildAbsoluteURL(m.FsFile)
}

// VerboseName returns verbose name for model.
func (m *SocketEnvironment) VerboseName() string {
	return "SocketEnvironment"
}

// StatusString returns verbose value for status.
func (m *SocketEnvironment) StatusString() string {
	return SocketEnvironmentStatus[m.Status]
}

// AfterUpdate hook.
func (m *SocketEnvironment) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *SocketEnvironment) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// BeforeUpdate hook.
func (m *SocketEnvironment) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}

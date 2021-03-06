package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database/fields"
	"github.com/Syncano/pkg-go/v2/storage"
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
	tableName struct{} `pg:"?schema.sockets_socketenvironment,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID          int
	Name        string
	Metadata    fields.JSON
	Description string
	Status      int
	StatusInfo  string
	CreatedAt   fields.Time
	UpdatedAt   fields.Time
	Checksum    string

	ZipFile string
	FsFile  string
}

func (m *SocketEnvironment) String() string {
	return fmt.Sprintf("SocketEnvironment<ID=%d, Name=%q>", m.ID, m.Name)
}

func (m *SocketEnvironment) Hash() string {
	return fmt.Sprintf("E:%s", m.Checksum)
}

func (m *SocketEnvironment) URL(s storage.DataStorage) string {
	return s.URL(settings.BucketData, m.FsFile)
}

// VerboseName returns verbose name for model.
func (m *SocketEnvironment) VerboseName() string {
	return "SocketEnvironment"
}

// StatusString returns verbose value for status.
func (m *SocketEnvironment) StatusString() string {
	return SocketEnvironmentStatus[m.Status]
}

// BeforeUpdate hook.
func (m *SocketEnvironment) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}

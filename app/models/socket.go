package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/settings"
)

// SocketStatus enum.
const (
	SocketStatusProcessing int = iota - 2 // start from -2
	SocketStatusError
	SocketStatusChecking
	SocketStatusOK
	SocketStatusPrompt
)

// SocketStatus to string map.
var SocketStatus = map[int]string{
	SocketStatusProcessing: "processing",
	SocketStatusError:      "error",
	SocketStatusChecking:   "checking",
	SocketStatusOK:         "ok",
	SocketStatusPrompt:     "prompt",
}

// Socket represents socket model.
type Socket struct {
	tableName struct{} `sql:"?schema.sockets_socket" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID          int
	Name        string
	Metadata    JSON
	Description string
	Status      int
	StatusInfo  string
	CreatedAt   Time
	UpdatedAt   Time
	Key         string
	Checksum    string

	Config        JSON
	InstallConfig JSON
	ZipFile       string
	ZipFileList   JSON
	Version       string
	Size          int
	Installed     JSON
	FileList      JSON

	EnvironmentID int
	Environment   *SocketEnvironment
}

func (m *Socket) String() string {
	return fmt.Sprintf("Socket<ID=%d, Name=%q>", m.ID, m.Name)
}

func getLocalPath(path string) string {
	if path[0] == '<' {
		path = path[1 : len(path)-1]
		path = regexp.MustCompile(`[^\w\s-]`).ReplaceAllString(path, "")
		path = regexp.MustCompile(`[-\s]+`).ReplaceAllString(path, "-")
	}
	return path
}

func buildAbsoluteURL(path string) string {
	url := settings.API.MediaPrefix + path
	if url[0] == '/' {
		return fmt.Sprintf("http://%s%s", settings.API.Host, url)
	}
	return url
}

// Files ...
func (m *Socket) Files() map[string]string {
	f := make(map[string]string)
	for path, data := range m.FileList.Get().(map[string]interface{}) {
		if path == settings.Socket.YAML {
			continue
		}
		url := data.(map[string]interface{})["file"].(string)
		f[buildAbsoluteURL(url)] = getLocalPath(path)
	}
	return f
}

// Hash ...
func (m *Socket) Hash() string {
	return fmt.Sprintf("S:%s", m.Checksum)
}

// VerboseName returns verbose name for model.
func (m *Socket) VerboseName() string {
	return "Socket"
}

// StatusString returns verbose value for status.
func (m *Socket) StatusString() string {
	return SocketStatus[m.Status]
}

// AfterUpdate hook.
func (m *Socket) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Socket) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// BeforeUpdate hook.
func (m *Socket) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}

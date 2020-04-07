package models

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

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
	tableName struct{} `pg:"?schema.sockets_socket,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

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
		path = strings.ToLower(path)
	}

	return path
}

func buildAbsoluteURL(path string) string {
	url := settings.API.StorageURL + path
	if url[0] == '/' {
		return fmt.Sprintf("http://%s%s", settings.API.Host, url)
	}

	return url
}

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

// BeforeUpdate hook.
func (m *Socket) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}

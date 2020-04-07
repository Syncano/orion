package models

import (
	"fmt"
)

// SocketEndpoint represents socket endpoint model.
type SocketEndpoint struct {
	tableName struct{} `pg:"?schema.sockets_socketendpoint,discard_unknown_columns"` // nolint

	ID       int
	Name     string
	Metadata JSON
	SocketID int
	Socket   *Socket
	Calls    JSON
}

func (m *SocketEndpoint) String() string {
	return fmt.Sprintf("SocketEndpoint<ID=%d, Name=%q>", m.ID, m.Name)
}

// MatchCall returns matched call.
func (m *SocketEndpoint) MatchCall(req string) map[string]interface{} {
	for _, call := range m.Calls.Get().([]interface{}) {
		methods := call.(map[string]interface{})["methods"].([]interface{})
		for _, meth := range methods {
			if meth == "*" || meth == req {
				return call.(map[string]interface{})
			}
		}
	}

	return nil
}

// Entrypoint returns call entrypoint.
func (m *SocketEndpoint) Entrypoint(call map[string]interface{}) string {
	path := call["path"].(string)
	return getLocalPath(path)
}

// VerboseName returns verbose name for model.
func (m *SocketEndpoint) VerboseName() string {
	return "SocketEndpoint"
}

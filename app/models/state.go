package models

import (
	"reflect"

	"github.com/go-pg/pg/orm"
	"github.com/mitchellh/hashstructure"
)

type snapshot struct {
	hash  map[string]uint64
	value map[string]interface{}
}

func newSnapshot() *snapshot {
	return &snapshot{
		hash:  make(map[string]uint64),
		value: make(map[string]interface{}),
	}
}

// State ...
type State struct {
	before *snapshot
	after  *snapshot

	virtualStore string
	virtual      map[string]struct{}
	sqlNames     map[string]string
}

// Snapshot ...
func (s *State) Snapshot(m interface{}, virt map[string]StateField) {
	var (
		name string
		tag  string
		hash uint64
		snap *snapshot
		val  interface{}
		err  error
	)

	// Rolling snapshot.
	if s.before == nil {
		s.before = newSnapshot()
		s.virtual = make(map[string]struct{})
		s.sqlNames = make(map[string]string)
		snap = s.before
	} else if s.after == nil {
		s.after = newSnapshot()
		snap = s.after
	} else {
		s.before = s.after
		s.after = newSnapshot()
		snap = s.after
	}

	// Process fields.
	strct := reflect.ValueOf(m).Elem()
	typ := strct.Type()
	table := orm.GetTable(typ)
	for _, field := range table.Fields {
		if tag = field.Field.Tag.Get("state"); tag == "-" {
			continue
		}

		val = field.Value(strct).Interface()
		name = field.GoName
		s.sqlNames[name] = field.SQLName

		if hash, err = hashstructure.Hash(val, nil); err != nil {
			continue
		}
		if tag == "virtual" {
			s.virtualStore = name
		}
		snap.hash[name] = hash
		snap.value[name] = val
	}

	// Process virtual fields.
	for name, v := range virt {
		val = v.Get(m)
		hash, err = hashstructure.Hash(val, nil)
		if err != nil {
			continue
		}

		s.virtual[name] = struct{}{}
		snap.hash[name] = hash
		snap.value[name] = val
	}
}

func (s *State) changed(virtual, sqlnames bool) []string {
	if s.after == nil {
		return nil
	}

	var dirty []string
	for k, v := range s.before.hash {
		if !virtual {
			// Skip virtual fields.
			if _, ok := s.virtual[k]; ok {
				continue
			}
		} else if k == s.virtualStore {
			// Skip field storing virtual fields (virtualStore) if including virtual fields themselves.
			continue
		}

		if sqlnames {
			k = s.sqlNames[k]
		}
		if v != s.after.hash[k] {
			dirty = append(dirty, k)
		}
	}
	return dirty
}

// Changes ...
func (s *State) Changes() []string {
	return s.changed(false, false)
}

// SQLChanges ...
func (s *State) SQLChanges() []string {
	return s.changed(false, true)
}

// ChangesVirtual ...
func (s *State) ChangesVirtual() []string {
	return s.changed(true, false)
}

// SQLChangesVirtual ...
func (s *State) SQLChangesVirtual() []string {
	return s.changed(true, true)
}

// HasChanged ...
func (s *State) HasChanged(f string) bool {
	if s.after == nil {
		return false
	}
	return s.before.hash[f] != s.after.hash[f]
}

// OldValue ...
func (s *State) OldValue(f string) interface{} {
	if s.before == nil {
		return nil
	}
	return s.before.value[f]
}

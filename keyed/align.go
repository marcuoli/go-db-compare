// Package keyed aligns complete or partial named snapshots deterministically.
package keyed

import (
	"errors"
	"github.com/marcuoli/go-db-compare/value"
	"sort"
)

// Entry is one uniquely keyed scalar. Keys retain caller-owned case and scope.
type Entry struct {
	Key   string
	Value value.Cell
}

// Snapshot distinguishes a successful empty inventory from a failed enumeration.
// Absent keys become Absent only when Complete is true; otherwise Unavailable.
type Snapshot struct {
	Complete bool
	Entries  []Entry
}

// Row aligns the same key on both sides and preserves exact comparison values.
type Row struct {
	Key        string
	Comparison value.Result
}

// Align returns a fresh sorted union. Duplicate keys and malformed values fail
// before any rows are returned; a partial prefix cannot masquerade as complete.
func Align(source, target Snapshot) ([]Row, error) {
	index := func(snapshot Snapshot) (map[string]value.Cell, error) {
		values := map[string]value.Cell{}
		if len(snapshot.Entries) > 100000 {
			return nil, errors.New("snapshot exceeds entry capacity")
		}
		for _, entry := range snapshot.Entries {
			if entry.Key == "" || len(entry.Key) > 4096 {
				return nil, errors.New("invalid comparison key")
			}
			if _, exists := values[entry.Key]; exists {
				return nil, errors.New("duplicate comparison key")
			}
			if err := value.Validate(entry.Value); err != nil {
				return nil, err
			}
			values[entry.Key] = entry.Value
		}
		return values, nil
	}
	first, err := index(source)
	if err != nil {
		return nil, err
	}
	second, err := index(target)
	if err != nil {
		return nil, err
	}
	keys := map[string]bool{}
	for key := range first {
		keys[key] = true
	}
	for key := range second {
		keys[key] = true
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	rows := make([]Row, 0, len(keys))
	for _, key := range ordered {
		a, exists := first[key]
		if !exists {
			a = value.Cell{State: value.Unavailable}
			if source.Complete {
				a.State = value.Absent
			}
		}
		b, exists := second[key]
		if !exists {
			b = value.Cell{State: value.Unavailable}
			if target.Complete {
				b.State = value.Absent
			}
		}
		result, err := value.Compare(a, b)
		if err != nil {
			return nil, err
		}
		rows = append(rows, Row{Key: key, Comparison: result})
	}
	return rows, nil
}

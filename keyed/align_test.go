package keyed_test

import (
	"github.com/marcuoli/go-db-compare/keyed"
	"github.com/marcuoli/go-db-compare/value"
	"reflect"
	"testing"
)

func known(text string) value.Cell {
	return value.Cell{State: value.Known, Kind: value.Text, Exact: text}
}
func TestAlignmentKeepsCompletenessAndSortedKeys(t *testing.T) {
	source := keyed.Snapshot{Complete: true, Entries: []keyed.Entry{{Key: "z", Value: known("")}, {Key: "a", Value: known("one")}}}
	target := keyed.Snapshot{Entries: []keyed.Entry{{Key: "b", Value: known("two")}}}
	rows, err := keyed.Align(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].Key != "a" || rows[1].Key != "b" || rows[2].Key != "z" || rows[0].Comparison.Status != value.Unknown || rows[1].Comparison.Status != value.TargetOnly {
		t.Fatal(rows)
	}
	target.Complete = true
	again, err := keyed.Align(source, target)
	if err != nil || again[0].Comparison.Status != value.SourceOnly {
		t.Fatal(again, err)
	}
	source.Entries[0].Value = known("mutated")
	if rows[2].Comparison.Source.Exact != "" {
		t.Fatal("result aliases source")
	}
	if reflect.DeepEqual(rows, again) {
		t.Fatal("completeness lost")
	}
}
func TestDuplicateNeverReturnsPrefix(t *testing.T) {
	rows, err := keyed.Align(keyed.Snapshot{Entries: []keyed.Entry{{Key: "a", Value: known("1")}, {Key: "a", Value: known("2")}}}, keyed.Snapshot{})
	if err == nil || len(rows) != 0 {
		t.Fatal("duplicate silently overwritten")
	}
}

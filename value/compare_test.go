package value_test

import (
	"github.com/marcuoli/go-db-compare/value"
	"testing"
)

func TestExactPresenceAndDelta(t *testing.T) {
	text := func(s string) value.Cell { return value.Cell{State: value.Known, Kind: value.Text, Exact: s} }
	integer := func(s string) value.Cell { return value.Cell{State: value.Known, Kind: value.Integer, Exact: s} }
	tests := []struct {
		name   string
		a, b   value.Cell
		status value.Status
		delta  string
	}{
		{"large", integer("9007199254740993"), integer("9007199254740994"), value.Different, "1"}, {"negative delta", integer("100"), integer("0"), value.Different, "-100"}, {"zero known", integer("0"), integer("0"), value.Equal, "0"}, {"null empty", value.Cell{State: value.Null}, text(""), value.Different, ""}, {"empty marker", text(""), text("<empty>"), value.Different, ""}, {"case", text("FALSE"), text("false"), value.CaseOnly, ""}, {"whitespace", text("x "), text("x"), value.Different, ""}, {"source only", integer("0"), value.Cell{State: value.Absent}, value.SourceOnly, ""}, {"unknown", integer("0"), value.Cell{State: value.Unavailable}, value.Unknown, ""}, {"both absent", value.Cell{State: value.Absent}, value.Cell{State: value.Absent}, value.Unknown, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := value.Compare(test.a, test.b)
			if err != nil || got.Status != test.status || got.Delta != test.delta || got.Source != test.a || got.Target != test.b {
				t.Fatalf("comparison changed: %+v %v", got, err)
			}
		})
	}
	for _, cell := range []value.Cell{{State: value.Known, Kind: value.Integer, Exact: "1.0"}, {State: value.Null, Exact: "0"}, {State: "bad"}, {State: value.Known, Kind: value.Text, Exact: string([]byte{0xff})}} {
		if _, err := value.Compare(cell, text("")); err == nil {
			t.Fatal("malformed cell admitted")
		}
	}
}
func TestCallerSelectedTextNormalization(t *testing.T) {
	if got := value.Analyze([]string{" FALSE ", "false"}, true); !got.Different || !got.CaseOnly {
		t.Fatal(got)
	}
	if got := value.Analyze([]string{" same ", "same"}, true); got.Different || got.CaseOnly {
		t.Fatal(got)
	}
	if value.Analyze(nil, true).Different {
		t.Fatal("empty input differed")
	}
}

func TestInvalidUTF8CannotBeCaseOnly(t *testing.T) {
	for _, pair := range [][2]string{{string([]byte{0xff}), string([]byte{0xfe})}, {string([]byte{0xff}), "\ufffd"}} {
		got := value.Analyze(pair[:], false)
		if !got.Different || got.CaseOnly {
			t.Fatalf("malformed bytes classified as case: %+v", got)
		}
	}
}

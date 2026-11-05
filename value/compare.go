// Package value compares exact scalars without collection or acceptance policy.
package value

import (
	"errors"
	"math/big"
	"strings"
	"unicode/utf8"
)

// State distinguishes authoritative absence and SQL NULL from unavailable data.
type State string

// Known retains even empty text; Absent is valid only after complete enumeration.
const (
	Known       State = "known"
	Null        State = "null"
	Absent      State = "absent"
	Unavailable State = "unavailable"
)

// Kind selects literal text or arbitrary-precision integer semantics.
type Kind string

// Text compares bytes exactly; Integer compares numeric values and retains text.
const (
	Text    Kind = "text"
	Integer Kind = "integer"
)

// Cell preserves original text. Non-Known states must have no Exact payload.
// Reason is caller-owned diagnostic metadata and does not change equality.
type Cell struct {
	State  State  `json:"state"`
	Kind   Kind   `json:"kind,omitempty"`
	Exact  string `json:"exact,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Status describes observed comparison, never approval or readiness.
type Status string

// Equal and the other statuses retain one-sided data and missing proof explicitly.
const (
	Equal      Status = "equal"
	Different  Status = "different"
	CaseOnly   Status = "case_only"
	SourceOnly Status = "source_only"
	TargetOnly Status = "target_only"
	Unknown    Status = "unavailable"
)

// Result preserves both cells and an exact target-minus-source integer delta.
// Delta is empty when either side lacks a known integer value.
type Result struct {
	Source Cell   `json:"source"`
	Target Cell   `json:"target"`
	Status Status `json:"status"`
	Delta  string `json:"delta,omitempty"`
}

// Validate rejects malformed states, invalid Unicode and excessive integer work.
func Validate(cell Cell) error {
	if !utf8.ValidString(cell.Exact) || !utf8.ValidString(cell.Reason) || len(cell.Exact) > 1<<20 || len(cell.Reason) > 4096 {
		return errors.New("comparison cell exceeds text contract")
	}
	switch cell.State {
	case Known:
		if cell.Kind != Text && cell.Kind != Integer {
			return errors.New("known cell requires a kind")
		}
		if cell.Kind == Integer {
			if len(cell.Exact) > 4096 {
				return errors.New("integer exceeds comparison capacity")
			}
			if _, ok := new(big.Int).SetString(cell.Exact, 10); !ok {
				return errors.New("invalid exact integer")
			}
		}
	case Null, Absent, Unavailable:
		if cell.Exact != "" {
			return errors.New("non-value state contains text")
		}
		if cell.Kind != "" && cell.Kind != Text && cell.Kind != Integer {
			return errors.New("invalid cell kind")
		}
	default:
		return errors.New("invalid comparison state")
	}
	return nil
}

// Compare owns no caller memory and applies no trimming, numeric float conversion
// or assumed zeros. Two absences are unknown, not a measured equality.
func Compare(source, target Cell) (Result, error) {
	result := Result{Source: source, Target: target, Status: Unknown}
	if err := Validate(source); err != nil {
		return result, err
	}
	if err := Validate(target); err != nil {
		return result, err
	}
	if source.State == Unavailable || target.State == Unavailable {
		return result, nil
	}
	if source.State == Absent && target.State == Absent {
		return result, nil
	}
	if source.State == Absent {
		result.Status = TargetOnly
		return result, nil
	}
	if target.State == Absent {
		result.Status = SourceOnly
		return result, nil
	}
	if source.State == Null || target.State == Null {
		result.Status = Different
		if source.State == Null && target.State == Null {
			result.Status = Equal
		}
		return result, nil
	}
	if source.Kind != target.Kind {
		result.Status = Different
		return result, nil
	}
	result.Status = Different
	if source.Kind == Integer {
		first, _ := new(big.Int).SetString(source.Exact, 10)
		second, _ := new(big.Int).SetString(target.Exact, 10)
		result.Delta = new(big.Int).Sub(second, first).String()
		if first.Cmp(second) == 0 {
			result.Status = Equal
		}
		return result, nil
	}
	text := Analyze([]string{source.Exact, target.Exact}, false)
	if !text.Different {
		result.Status = Equal
	} else if text.CaseOnly {
		result.Status = CaseOnly
	}
	return result, nil
}

// TextResult separates exact equality from a difference containing only case.
type TextResult struct{ Different, CaseOnly bool }

// Analyze compares all strings with optional caller-selected whitespace trimming.
// Empty text is a real value; literal marker strings never act as sentinels.
// Zero/one-element input has no difference and is not classified case-only.
func Analyze(values []string, trim bool) TextResult {
	if len(values) < 2 {
		return TextResult{}
	}
	normalize := func(value string) string {
		if trim {
			return strings.TrimSpace(value)
		}
		return value
	}
	first := normalize(values[0])
	different, foldEqual := false, utf8.ValidString(first)
	for _, value := range values[1:] {
		value = normalize(value)
		different = different || value != first
		foldEqual = foldEqual && utf8.ValidString(value) && strings.EqualFold(first, value)
	}
	return TextResult{Different: different, CaseOnly: different && foldEqual}
}

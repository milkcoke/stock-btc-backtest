package signal

import "strings"

// AllOf is met only when every member is. It is itself a Condition, so a
// composite rule can be nested or passed anywhere a single rule is expected.
type AllOf []Condition

func (a AllOf) Met(f Frame, i int) bool {
	for _, c := range a {
		if !c.Met(f, i) {
			return false
		}
	}
	return true
}

func (a AllOf) Describe() string {
	parts := make([]string, 0, len(a))
	for _, c := range a {
		parts = append(parts, c.Describe())
	}
	return strings.Join(parts, " + ")
}

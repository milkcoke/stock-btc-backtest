package signal

// Condition answers one question about one day. Implementations stay small on
// purpose: each owns a single rule, so adding, removing or reordering rules
// never touches the detector.
type Condition interface {
	Met(f Frame, i int) bool
	// Describe renders the rule for the report header, so the reader can always
	// see the exact rule that produced the table underneath it.
	Describe() string
}

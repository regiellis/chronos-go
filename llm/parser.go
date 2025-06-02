package llm

import "github.com/regiellis/chronos-go/chronos"

// Parser defines the interface for parsing natural language time entries.
type Parser interface {
	ParseEntry(input string) (*chronos.Entry, error)
}

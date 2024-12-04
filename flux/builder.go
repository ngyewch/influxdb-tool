package flux

import (
	"fmt"
	"strings"
)

type Builder struct {
	parts []string
}

func NewBuilder(bucket string) *Builder {
	return &Builder{
		parts: []string{fmt.Sprintf(`from(bucket: "%s")`, bucket)},
	}
}

func (b *Builder) String() string {
	return strings.Join(b.parts, "\n  |> ")
}

func (b *Builder) Range(start string, stop string) *Builder {
	if stop != "" {
		b.parts = append(b.parts, fmt.Sprintf(`range(start: %s, stop: %s)`, start, stop))
	} else {
		b.parts = append(b.parts, fmt.Sprintf(`range(start: %s)`, start))
	}
	return b
}

func (b *Builder) Filter(predicate string) *Builder {
	b.parts = append(b.parts, fmt.Sprintf(`filter(fn: (r) => %s)`, predicate))
	return b
}

func (b *Builder) AggregateWindow(every string, fn string, createEmpty bool) *Builder {
	b.parts = append(b.parts, fmt.Sprintf(`aggregateWindow(every: %s, fn: %s, createEmpty: %v)`, every, fn, createEmpty))
	return b
}

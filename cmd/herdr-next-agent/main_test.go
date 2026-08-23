package main

import (
	"errors"
	"testing"

	"github.com/choplin/herdr-next-agent/internal/navigation"
)

func TestParseDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want navigation.Direction
		ok   bool
	}{
		{name: "default", want: navigation.NextDirection, ok: true},
		{name: "next", args: []string{"next"}, want: navigation.NextDirection, ok: true},
		{name: "previous", args: []string{"previous"}, want: navigation.PreviousDirection, ok: true},
		{name: "unsupported", args: []string{"back"}},
		{name: "too many", args: []string{"next", "previous"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDirection(test.args)
			var argumentErr *ArgumentError
			if got != test.want || (err == nil) != test.ok {
				t.Fatalf("parseDirection(%q) = (%v, %v), want (%v, ok=%t)", test.args, got, err, test.want, test.ok)
			}
			if !test.ok && !errors.As(err, &argumentErr) {
				t.Fatalf("parseDirection(%q) error = %v, want ArgumentError", test.args, err)
			}
		})
	}
}

package handlers

import (
	"strings"
	"testing"
)

func TestNormalizeChatMessage(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		wantOK bool
	}{
		{name: "trims whitespace", input: "  안녕하세요  ", want: "안녕하세요", wantOK: true},
		{name: "rejects blank", input: " \n\t ", want: "", wantOK: false},
		{name: "accepts limit", input: strings.Repeat("가", 2000), want: strings.Repeat("가", 2000), wantOK: true},
		{name: "rejects over limit", input: strings.Repeat("가", 2001), want: strings.Repeat("가", 2001), wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeChatMessage(tt.input)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("normalizeChatMessage() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

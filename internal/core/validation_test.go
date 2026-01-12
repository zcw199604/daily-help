package core

import (
	"strings"
	"testing"
)

func TestValidateContainerName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "empty", in: "", wantErr: true},
		{name: "spaces", in: "   ", wantErr: true},
		{name: "invalid-char", in: "bad/name", wantErr: true},
		{name: "leading-dash", in: "-app", wantErr: true},
		{name: "unicode", in: "应用", wantErr: true},
		{name: "max-128", in: "a" + strings.Repeat("b", 127), want: "a" + strings.Repeat("b", 127)},
		{name: "too-long", in: "a" + strings.Repeat("b", 128), wantErr: true},
		{name: "ok-simple", in: "app", want: "app"},
		{name: "ok-trim", in: "  app  ", want: "app"},
		{name: "ok-dots-underscore-dash", in: "app_1.2-3", want: "app_1.2-3"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateContainerName(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ValidateContainerName() error = nil, want not nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateContainerName() error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("ValidateContainerName() = %q, want %q", got, tc.want)
			}
		})
	}
}

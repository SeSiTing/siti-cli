package shell

import "testing"

func TestExport(t *testing.T) {
	got := Export("FOO", "bar")
	want := `export FOO="bar";`
	if got != want {
		t.Errorf("Export() = %q, want %q", got, want)
	}
}

func TestExportRef(t *testing.T) {
	got := ExportRef("ANTHROPIC_BASE_URL", "MINIMAX_BASE_URL")
	want := `export ANTHROPIC_BASE_URL="$MINIMAX_BASE_URL";`
	if got != want {
		t.Errorf("ExportRef() = %q, want %q", got, want)
	}
}

func TestUnset(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{"empty", nil, ""},
		{"single", []string{"FOO"}, "unset FOO;"},
		{"multiple", []string{"A", "B", "C"}, "unset A B C;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unset(tt.keys...); got != tt.want {
				t.Errorf("Unset(%v) = %q, want %q", tt.keys, got, tt.want)
			}
		})
	}
}

func TestSourceIf(t *testing.T) {
	got := SourceIf("/tmp/x.sh")
	want := `if [ -f "/tmp/x.sh" ]; then source "/tmp/x.sh"; fi`
	if got != want {
		t.Errorf("SourceIf() = %q, want %q", got, want)
	}
}

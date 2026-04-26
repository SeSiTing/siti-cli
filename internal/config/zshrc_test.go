package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// withFakeHome temporarily sets HOME to a tmpdir and returns it.
// Restored automatically via t.Cleanup.
func withFakeHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", old) })
	return dir
}

func TestReadProviders_BasicDiscovery(t *testing.T) {
	home := withFakeHome(t)
	writeFile(t, filepath.Join(home, ".zshrc"), `
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export MINIMAX_API_KEY="sk-test"
export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/anthropic"
export ZHIPU_MODEL="glm-4.6"
export ANTHROPIC_BASE_URL="..."  # should be skipped (target itself)
`)

	got, err := ReadProviders()
	if err != nil {
		t.Fatalf("ReadProviders: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 providers, got %d: %+v", len(got), got)
	}

	mm, ok := got.Find("MINIMAX")
	if !ok {
		t.Fatal("MINIMAX not found")
	}
	if mm.AuthTokenVar != "MINIMAX_API_KEY" {
		t.Errorf("MINIMAX AuthTokenVar = %q, want MINIMAX_API_KEY", mm.AuthTokenVar)
	}
	if mm.ModelVar != "" {
		t.Errorf("MINIMAX ModelVar = %q, want empty", mm.ModelVar)
	}

	zp, _ := got.Find("ZHIPU")
	if zp.AuthTokenVar != "DEFAULT_AUTH_TOKEN" {
		t.Errorf("ZHIPU AuthTokenVar = %q, want DEFAULT_AUTH_TOKEN (no API_KEY defined)", zp.AuthTokenVar)
	}
	if zp.ModelVar != "ZHIPU_MODEL" {
		t.Errorf("ZHIPU ModelVar = %q, want ZHIPU_MODEL", zp.ModelVar)
	}
}

func TestReadProviders_ZshenvWinsOnDuplicate(t *testing.T) {
	home := withFakeHome(t)
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ALI_BASE_URL="https://from-zshenv.example"
export ALI_API_KEY="from-zshenv-key"
`)
	writeFile(t, filepath.Join(home, ".zshrc"), `
export ALI_BASE_URL="https://from-zshrc.example"
`)

	got, err := ReadProviders()
	if err != nil {
		t.Fatalf("ReadProviders: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 provider, got %d", len(got))
	}
	if got[0].AuthTokenVar != "ALI_API_KEY" {
		t.Errorf("zshenv should have won; AuthTokenVar = %q", got[0].AuthTokenVar)
	}
}

func TestReadSkipList(t *testing.T) {
	t.Setenv("SITI_AI_SKIP", "OPENAI,BAILIAN, AZURE ")
	got := ReadSkipList()
	want := []string{"OPENAI", "BAILIAN", "AZURE"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d]=%q, want %q", i, got[i], want[i])
		}
	}
}

func TestProvider_IsSkipped(t *testing.T) {
	p := Provider{Name: "MINIMAX"}
	if p.IsSkipped([]string{"OPENAI", "BAILIAN"}) {
		t.Error("MINIMAX should not be skipped")
	}
	if !p.IsSkipped([]string{"openai", "minimax"}) {
		t.Error("MINIMAX should be skipped (case-insensitive)")
	}
}

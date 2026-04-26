// Package config provides read-only parsing of shell config files to discover
// AI provider definitions (e.g. MINIMAX_BASE_URL, MINIMAX_API_KEY).
//
// siti never writes to ~/.zshrc or ~/.zshenv; those files are owned by the user.
package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Provider represents an AI service provider discovered from shell config files.
type Provider struct {
	// Name is the uppercase prefix, e.g. "MINIMAX", "KIMI", "ALI".
	Name string
	// BaseURLVar is the env var name holding the base URL, e.g. "MINIMAX_BASE_URL".
	BaseURLVar string
	// AuthTokenVar is the env var used for auth. Defaults to "DEFAULT_AUTH_TOKEN".
	AuthTokenVar string
	// ModelVar is the env var for model overrides, e.g. "MINIMAX_MODEL". Empty if unset.
	ModelVar string
}

// DisplayName returns the lowercase display name, e.g. "minimax".
func (p Provider) DisplayName() string {
	return strings.ToLower(p.Name)
}

// IsSkipped reports whether this provider is in the skip list.
func (p Provider) IsSkipped(skipList []string) bool {
	upper := strings.ToUpper(p.Name)
	for _, s := range skipList {
		if strings.ToUpper(strings.TrimSpace(s)) == upper {
			return true
		}
	}
	return false
}

// ProviderList is a slice of Provider with lookup helpers.
type ProviderList []Provider

// Find returns the provider matching name (case-insensitive) and whether it was found.
func (pl ProviderList) Find(name string) (Provider, bool) {
	upper := strings.ToUpper(name)
	for _, p := range pl {
		if p.Name == upper {
			return p, true
		}
	}
	return Provider{}, false
}

// reExportBaseURL matches lines like: export MINIMAX_BASE_URL="..."
var reExportBaseURL = regexp.MustCompile(`^export\s+([A-Z0-9_]+)_BASE_URL=`)

// ReadProviders parses ~/.zshenv and ~/.zshrc to discover AI provider definitions.
// It reads both files and deduplicates by provider name (zshenv takes precedence).
//
// A provider is recognized when a line matching `export <NAME>_BASE_URL=` is found.
// The following optional variables are also probed:
//   - <NAME>_API_KEY  → AuthTokenVar (falls back to DEFAULT_AUTH_TOKEN)
//   - <NAME>_MODEL    → ModelVar
func ReadProviders() (ProviderList, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Collect lines from both files; zshenv first so it wins on duplicate names.
	var lines []string
	for _, name := range []string{".zshenv", ".zshrc"} {
		ls, _ := readLines(filepath.Join(home, name)) // ignore missing files
		lines = append(lines, ls...)
	}

	// Build a set of all defined variable names for fast probing.
	defined := buildDefinedSet(lines)

	// Collect provider names in order (first occurrence wins).
	seen := map[string]bool{}
	var providers ProviderList

	for _, line := range lines {
		m := reExportBaseURL.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		if name == "ANTHROPIC" || seen[name] {
			continue // skip the target variable itself and duplicates
		}
		seen[name] = true

		p := Provider{
			Name:       name,
			BaseURLVar: name + "_BASE_URL",
		}
		if defined[name+"_API_KEY"] {
			p.AuthTokenVar = name + "_API_KEY"
		} else {
			p.AuthTokenVar = "DEFAULT_AUTH_TOKEN"
		}
		if defined[name+"_MODEL"] {
			p.ModelVar = name + "_MODEL"
		}
		providers = append(providers, p)
	}

	return providers, nil
}

// ReadSkipList returns the SITI_AI_SKIP provider names from the environment
// or from shell config files (comma-separated).
func ReadSkipList() []string {
	if v := os.Getenv("SITI_AI_SKIP"); v != "" {
		return splitComma(v)
	}

	home, _ := os.UserHomeDir()
	reSITI := regexp.MustCompile(`^export\s+SITI_AI_SKIP=["']?([^"'\n]+)["']?`)

	for _, name := range []string{".zshenv", ".zshrc"} {
		ls, _ := readLines(filepath.Join(home, name))
		for _, line := range ls {
			if m := reSITI.FindStringSubmatch(line); m != nil {
				return splitComma(m[1])
			}
		}
	}
	return nil
}

// CurrentProvider returns the provider name whose BASE_URL matches
// the current ANTHROPIC_BASE_URL environment variable.
func CurrentProvider(providers ProviderList) string {
	current := os.Getenv("ANTHROPIC_BASE_URL")
	if current == "" {
		return ""
	}
	for _, p := range providers {
		if v := os.Getenv(p.BaseURLVar); v != "" && v == current {
			return p.DisplayName()
		}
	}
	return ""
}

// readLines returns all lines from a file, ignoring errors (e.g. file not found).
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// buildDefinedSet returns a set of variable names that appear as `export VAR=` in lines.
var reExportAny = regexp.MustCompile(`^export\s+([A-Z0-9_]+)=`)

func buildDefinedSet(lines []string) map[string]bool {
	set := map[string]bool{}
	for _, line := range lines {
		if m := reExportAny.FindStringSubmatch(line); m != nil {
			set[m[1]] = true
		}
	}
	return set
}

func splitComma(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

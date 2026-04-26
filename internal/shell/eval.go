// Package shell provides helpers for generating shell statements that the
// parent shell wrapper will eval (exit-code-10 protocol).
//
// These helpers are pure string builders. They are queued by `cmd.Eval(c, ...)`
// (see cmd/root.go), which stores them in the per-invocation buffer attached
// to the cobra context. cmd.Execute() prints the buffer to stdout and exits
// with code 10 after the command returns nil.
//
// Example (from cmd/proxy.go):
//
//	cmd.Eval(c,
//	    shell.Export("http_proxy",  "http://127.0.0.1:7890"),
//	    shell.Export("HTTPS_PROXY", "http://127.0.0.1:7890"),
//	)
//	return nil
package shell

import "fmt"

// Export generates:  export KEY="value";
func Export(key, value string) string {
	return fmt.Sprintf(`export %s="%s";`, key, value)
}

// ExportRef generates:  export KEY="$REF_VAR";
func ExportRef(key, refVar string) string {
	return fmt.Sprintf(`export %s="$%s";`, key, refVar)
}

// Unset generates:  unset KEY1 KEY2;
func Unset(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	s := "unset"
	for _, k := range keys {
		s += " " + k
	}
	return s + ";"
}

// SourceIf generates:  if [ -f "file" ]; then source "file"; fi
func SourceIf(file string) string {
	return fmt.Sprintf(`if [ -f "%s" ]; then source "%s"; fi`, file, file)
}

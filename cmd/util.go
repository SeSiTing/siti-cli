package cmd

import "os"

// lookupEnv returns the value of the named environment variable, or "" if unset.
func lookupEnv(key string) string {
	v, _ := os.LookupEnv(key)
	return v
}

// firstNonEmpty returns the first non-empty string from the provided values.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

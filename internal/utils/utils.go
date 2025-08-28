package utils

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
)

// MergeVars reads key=value pairs from one or more files
// and returns a new lookupEnv func that prefers these vars over fallback.
func MergeVars(files []string, fallback func(string) (string, bool)) (func(string) (string, bool), error) {
	all := make(map[string]string)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		vars, err := readVars(f)
		_ = f.Close()
		if err != nil {
			return nil, fmt.Errorf("parsing vars in %q: %w", file, err)
		}
		// merge: later files override earlier ones
		maps.Copy(all, vars)
	}

	// Create new lookup func that prefers vars over fallback.
	return func(key string) (string, bool) {
		if v, ok := all[key]; ok {
			return v, true
		}
		return fallback(key)
	}, nil
}

// readVars reads KEY=VAL lines into a map.
func readVars(r io.Reader) (map[string]string, error) {
	vars := make(map[string]string)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("invalid var line %q (expected KEY=VALUE)", line)
		}
		vars[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return vars, nil
}

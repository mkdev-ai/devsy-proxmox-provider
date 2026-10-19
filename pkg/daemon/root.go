package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultStateRoot is the absolute default state root used when the provider
// pins Devsy's daemon data directory. It never contains a literal ~.
const DefaultStateRoot = "/root/.devsy/data"

// resolveStateRoot expands a leading ~ to the real home directory and
// requires the result to be an absolute path.
//
// The provider's agent dataPath (provider.yaml manifest) controls where Devsy
// installs its daemon. Devsy's own daemon installer rejects a relative or
// non-expanded state root with
// "daemon state root must be absolute: "~/.devsy/data"". This helper is the
// contract for resolving a valid absolute root; callers must pass an absolute
// path (or a ~ that resolves to one).
func resolveStateRoot(raw string) (string, error) {
	root := raw
	if len(root) > 0 && root[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving daemon state root: %w", err)
		}
		root = home + root[1:]
	}
	if !filepath.IsAbs(root) {
		return "", &relativeRootError{raw: raw}
	}
	return root, nil
}

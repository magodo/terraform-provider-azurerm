package tfxsdk

import (
	"fmt"
	"slices"
)

// SchemaVersionIndex checks if the schema version is supported by the registered schema versions.
// The regVersions is supposed to be ordered from the least version.
// If supported, the index of the version in the regVersions is returned. Otherwise, -1 is returned.
func SchemaVersioIndex(regVersions []int, version int) (int, error) {
	if len(regVersions) == 0 {
		return -1, nil
	}

	idx := slices.Index(regVersions, version)
	if idx == -1 {
		return -1, nil
	}

	// Ensure the schema versions are contiguous from the supported one.
	for _, ver := range regVersions[idx:] {
		if ver != version {
			return -1, fmt.Errorf("the version after %d expects to be %d, got=%d", version-1, version, ver)
		}
		version += 1
	}
	return idx, nil
}

package service

import (
	"errors"
	"path/filepath"
	"strings"
)

var errInvalidPath = errors.New("invalid path")

// ResolveWalkPath joins path onto root and rejects values that leave root.
// filepath.Join treats an absolute element as a new root, so "/etc" would
// otherwise walk the host filesystem instead of the FTP tree.
func ResolveWalkPath(root, path string) (string, error) {
	if path == "" {
		return root, nil
	}
	if filepath.IsAbs(path) || strings.Contains(path, "..") {
		return "", errInvalidPath
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	joined, err := filepath.Abs(filepath.Join(absRoot, path))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, joined)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errInvalidPath
	}
	return joined, nil
}

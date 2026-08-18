package service

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveWalkPath(t *testing.T) {
	root := t.TempDir()

	got, err := ResolveWalkPath(root, "")
	require.NoError(t, err)
	require.Equal(t, root, got)

	got, err = ResolveWalkPath(root, "outbound")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "outbound"), got)

	_, err = ResolveWalkPath(root, "..")
	require.EqualError(t, err, "invalid path")

	_, err = ResolveWalkPath(root, "../outbound")
	require.EqualError(t, err, "invalid path")

	_, err = ResolveWalkPath(root, "/etc")
	require.EqualError(t, err, "invalid path")
}

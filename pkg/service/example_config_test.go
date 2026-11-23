package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

// TestExampleConfigs loads each example config that the repository ships or
// shows in the README, so that a broken example fails CI.
func TestExampleConfigs(t *testing.T) {
	root := filepath.Join("..", "..")

	paths := []string{
		filepath.Join(root, "examples", "config.yml"),
		filepath.Join(root, "configs", "config.default.yml"),
		filepath.Join(root, "configs", "config.docker.yml"),
	}

	// Write each full config in the README (a yaml block that starts with
	// "ACHTestHarness:") to a file and load it too.
	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	require.NoError(t, err)

	var found int
	for _, block := range strings.Split(string(readme), "```yaml\n")[1:] {
		block, _, _ = strings.Cut(block, "```")
		if !strings.HasPrefix(block, "ACHTestHarness:") {
			continue
		}
		found++

		path := filepath.Join(t.TempDir(), "readme.yml")
		require.NoError(t, os.WriteFile(path, []byte(block), 0600))
		paths = append(paths, path)
	}
	require.Positive(t, found, "no config found in README.md")

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Setenv("APP_CONFIG", path)

			cfg, err := service.LoadConfig(log.NewTestLogger())
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

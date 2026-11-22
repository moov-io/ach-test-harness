// generated-from:0fc1026d5a99d1d9297b54c4711c5f5aa2501f3e9f921bf02f02c16dcf7ce62b DO NOT REMOVE, DO UPDATE

package achtestharness

import (
	"embed"
)

// ConfigDefaults holds the default config, which is built into the binary.
// The binary then does not need the configs/ directory at run time.
//
//go:embed configs/config.default.yml
var ConfigDefaults embed.FS

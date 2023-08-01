package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig__Match(t *testing.T) {
	m := Match{}
	require.True(t, m.Empty())

	m.IndividualName = "John Doe"
	require.False(t, m.Empty())
}

func TestConfig__Amount(t *testing.T) {
	a := &Amount{}
	require.True(t, a.Empty())

	a.Value = 1234
	require.False(t, a.Empty())
}

func TestConfig__Return(t *testing.T) {
	var r Return

	r.Code = "R01"
	require.NoError(t, r.Validate())

	r.Code = "R99"
	require.Error(t, r.Validate())
}

// TODO JB: tests

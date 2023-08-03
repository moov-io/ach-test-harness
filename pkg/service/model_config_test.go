package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig__Config(t *testing.T) {
	cfg := &Config{
		Responses: []Response{
			{
				Action: Action{
					Copy:   &Copy{Path: "/reconciliation/"},
					Return: &Return{Code: "R01"},
				},
			},
		},
	}
	require.Error(t, cfg.Validate())

}

func TestConfig__Match(t *testing.T) {
	m := Match{}
	require.True(t, m.Empty())

	m.IndividualName = "John Doe"
	require.False(t, m.Empty())
}

func TestConfig__Response(t *testing.T) {
	r := &Response{
		Match: Match{
			IndividualName: "John Doe",
		},
		Action: Action{ // invalid
			Copy:   &Copy{Path: "/reconciliation/"},
			Return: &Return{Code: "R01"},
		},
	}
	require.Error(t, r.Validate())

	r = &Response{
		Match: Match{
			IndividualName: "John Doe",
		},
		Action: Action{
			Copy: &Copy{Path: "/reconciliation/"},
		},
	}
	require.NoError(t, r.Validate())

	r = &Response{
		// invalid - no Match
		Action: Action{
			Return: &Return{Code: "R01"},
		},
	}
	require.Error(t, r.Validate())
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

func TestConfig__Action(t *testing.T) {
	var delay, err = time.ParseDuration("12h")
	require.NoError(t, err)

	var actionCopy = &Copy{Path: "/reconciliation/"}
	var actionReturn = &Return{Code: "R01"}
	var actionCorrection = &Correction{Code: "C01"}

	t.Run("Delay only", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		require.Error(t, a.Validate())
	})

	t.Run("Delay + Copy", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		a.Copy = actionCopy
		require.Error(t, a.Validate())
	})

	t.Run("Delay + Return", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		a.Return = actionReturn
		require.NoError(t, a.Validate())
	})

	t.Run("Delay + Correction", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		a.Correction = actionCorrection
		require.NoError(t, a.Validate())
	})

	t.Run("Delay + Copy + Return", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		a.Copy = actionCopy
		a.Return = actionReturn
		require.Error(t, a.Validate())
	})

	t.Run("Delay + Copy + Correction", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		a.Copy = actionCopy
		a.Correction = actionCorrection
		require.Error(t, a.Validate())
	})

	t.Run("Delay + Copy + Return + Correction", func(t *testing.T) {
		var a Action
		a.Delay = &delay
		a.Copy = actionCopy
		a.Return = actionReturn
		a.Correction = actionCorrection
		require.Error(t, a.Validate())
	})

	t.Run("Copy only", func(t *testing.T) {
		var a Action
		a.Copy = actionCopy
		require.NoError(t, a.Validate())
	})

	t.Run("Copy + Return", func(t *testing.T) {
		var a Action
		a.Copy = actionCopy
		a.Return = actionReturn
		require.Error(t, a.Validate())
	})

	t.Run("Copy + Correction", func(t *testing.T) {
		var a Action
		a.Copy = actionCopy
		a.Correction = actionCorrection
		require.Error(t, a.Validate())
	})

	t.Run("Copy + Return + Correction", func(t *testing.T) {
		var a Action
		a.Copy = actionCopy
		a.Return = actionReturn
		a.Correction = actionCorrection
		require.Error(t, a.Validate())
	})

	t.Run("Return only", func(t *testing.T) {
		var a Action
		a.Return = actionReturn
		require.NoError(t, a.Validate())
	})

	t.Run("Return + Correction", func(t *testing.T) {
		var a Action
		a.Return = actionReturn
		a.Correction = actionCorrection
		require.Error(t, a.Validate())
	})

	t.Run("Correction only", func(t *testing.T) {
		var a Action
		a.Correction = actionCorrection
		require.NoError(t, a.Validate())
	})
}

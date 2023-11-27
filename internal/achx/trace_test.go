// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package achx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrace__ABA(t *testing.T) {
	// 9 digits
	require.Equal(t, "23138010", ABA8("231380104"))
	require.Equal(t, "4", ABACheckDigit("231380104"))

	// 10 digit from ACH server
	require.Equal(t, "12345678", ABA8("0123456789"))
	require.Equal(t, "9", ABACheckDigit("0123456789"))

	// 8 digits
	require.Equal(t, "12345678", ABA8("12345678"))
	require.Equal(t, "0", ABACheckDigit("12345678"))

	// short
	require.Equal(t, "", ABA8("1234"))
	require.Equal(t, "", ABACheckDigit("1234"))
	require.Equal(t, "", ABA8(""))
	require.Equal(t, "", ABACheckDigit(""))
}

func TestTraceNumber(t *testing.T) {
	for i := 0; i < 10000; i++ {
		trace, err := TraceNumber("121042882")
		require.NoError(t, err)
		require.NotEmpty(t, trace)
	}
}

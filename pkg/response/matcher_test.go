package response

import (
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestMatchAccountNumber(t *testing.T) {
	m := service.Match{
		AccountNumber: "777-33-11",
	}
	ed := ach.NewEntryDetail()
	ed.DFIAccountNumber = "777-33-11"

	// positive match
	require.True(t, matchesAccountNumber(m, ed))

	// negative match
	ed.DFIAccountNumber = "8171241"
	require.False(t, matchesAccountNumber(m, ed))
}

func TestMatchDebit(t *testing.T) {
	m := service.Match{
		Debit: &service.Debit{},
	}
	ed := ach.NewEntryDetail()

	tests := []int{ach.CheckingDebit, ach.SavingsDebit}
	for i := range tests {
		ed.TransactionCode = tests[i]

		require.True(t, matchedDebit(m, ed))
	}

	// negative matches
	tests = []int{ach.CheckingCredit, ach.SavingsCredit}
	for i := range tests {
		ed.TransactionCode = tests[i]

		require.False(t, matchedDebit(m, ed))
	}
}

func TestMatchIndividualName(t *testing.T) {
	m := service.Match{
		IndividualName: "John Doe",
	}
	ed := ach.NewEntryDetail()
	ed.IndividualName = "John Doe"

	// positive match
	require.True(t, matchedIndividualName(m, ed))

	// negative match
	ed.IndividualName = "Jane Doe"
	require.False(t, matchedIndividualName(m, ed))
}

func TestMatchTraceNumber(t *testing.T) {
	m := service.Match{
		TraceNumber: "12345678901234",
	}
	ed := ach.NewEntryDetail()
	ed.TraceNumber = "12345678901234"

	// positive match
	require.True(t, matchesTraceNumber(m, ed))

	// negative match
	ed.TraceNumber = "9876543201234"
	require.False(t, matchesTraceNumber(m, ed))
}

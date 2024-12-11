package match

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

func TestMatchAccountNumber(t *testing.T) {
	m := service.Match{
		AccountNumber: "777-33-11",
	}
	ed := ach.NewEntryDetail()
	ed.DFIAccountNumber = "777-33-11"

	// positive match
	require.True(t, AccountNumber(m, ed))

	// negative match
	ed.DFIAccountNumber = "8171241"
	require.False(t, AccountNumber(m, ed))
}

func TestMatchRoutingNumber(t *testing.T) {
	m := service.Match{
		RoutingNumber: "231380104",
	}
	ed := ach.NewEntryDetail()
	ed.RDFIIdentification = "23138010"
	ed.CheckDigit = "4"

	// positive match
	require.True(t, RoutingNumber(m, ed))

	// negative match - only CheckDigit matches
	ed.CheckDigit = "1"
	require.False(t, RoutingNumber(m, ed))

	// negative match - only RDFIIdentification matches
	ed.RDFIIdentification = "11111111"
	ed.CheckDigit = "4"
	require.False(t, RoutingNumber(m, ed))

	// negative match
	ed.CheckDigit = "1"
	require.False(t, RoutingNumber(m, ed))
}

func TestMatchAmount(t *testing.T) {
	m1 := service.Match{
		Amount: &service.Amount{
			Value: 12345,
		},
	}
	ed1 := ach.NewEntryDetail()
	ed1.Amount = 12345

	// positive match
	require.True(t, Amount(m1, ed1))

	// negative match
	ed1.Amount = 54321
	require.False(t, Amount(m1, ed1))

	m2 := service.Match{
		Amount: &service.Amount{
			Min: 10000,
			Max: 20000,
		},
	}
	ed2 := ach.NewEntryDetail()
	ed2.Amount = 12345

	// positive match
	require.True(t, Amount(m2, ed2))
	ed2.Amount = 10000
	require.True(t, Amount(m2, ed2))

	// negative match
	ed2.Amount = 100
	require.False(t, Amount(m2, ed2))
}

func TestMatchDebit(t *testing.T) {
	type test struct {
		input int
		want  bool
	}

	tests := []test{
		{input: ach.CheckingDebit, want: true},
		{input: ach.SavingsDebit, want: true},
		{input: ach.GLDebit, want: true},
		{input: ach.LoanDebit, want: true},
		{input: ach.CheckingCredit, want: false},
	}

	m := service.Match{}
	for _, tc := range tests {
		ed := ach.NewEntryDetail()
		ed.TransactionCode = tc.input
		require.Equal(t, tc.want, matchedDebit(m, ed))
	}
}

func TestMatchEntryType(t *testing.T) {
	type test struct {
		input     int
		entryType service.EntryType
		want      bool
	}

	tests := []test{
		{input: ach.CheckingDebit, entryType: service.EntryTypeDebit, want: true},
		{input: ach.SavingsDebit, entryType: service.EntryTypeDebit, want: true},
		{input: ach.CheckingCredit, entryType: service.EntryTypeCredit, want: true},
		{input: ach.CheckingDebit, entryType: service.EntryTypeCredit, want: false},
		{input: ach.SavingsDebit, entryType: service.EntryTypeCredit, want: false},
		{input: ach.CheckingCredit, entryType: service.EntryTypeDebit, want: false},
		// Prenotes
		{input: ach.CheckingDebit, entryType: service.EntryTypePrenote, want: false},
		{input: ach.CheckingPrenoteDebit, entryType: service.EntryTypePrenote, want: true},
		{input: ach.SavingsPrenoteDebit, entryType: service.EntryTypePrenote, want: true},
	}

	for _, tc := range tests {
		m := service.Match{
			EntryType: tc.entryType,
		}
		ed := ach.NewEntryDetail()
		ed.TransactionCode = tc.input

		require.Equal(t, tc.want, matchedEntryType(m, ed))
	}

	t.Run("exact value", func(t *testing.T) {
		m := service.Match{
			EntryType: "",
		}
		ed := ach.NewEntryDetail()
		ed.TransactionCode = ach.CheckingCredit

		require.False(t, matchedEntryType(m, ed))

		// Make them match
		m.EntryType = "22"
		require.True(t, matchedEntryType(m, ed))
	})
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
	require.True(t, TraceNumber(m, ed))

	// negative match
	ed.TraceNumber = "9876543201234"
	require.False(t, TraceNumber(m, ed))
}

// Following data is used for TestMultiMatch
// TransactionCode  RDFIIdentification  AccountNumber      Amount  Name                    TraceNumber      Category
// 37               22147578            221475786          100     John Doe                273976367520468

// TransactionCode  RDFIIdentification  AccountNumber      Amount  Name                    TraceNumber      Category
// 22               27397636            273976369          100     Incorrect Name          273976367520469
func TestMultiMatch(t *testing.T) {
	var delay, err = time.ParseDuration("12h")
	require.NoError(t, err)

	var matchNone = service.Match{
		Amount: &service.Amount{
			Min: 500000,  // $5,000.00
			Max: 1000000, // $10,000.00
		},
		EntryType: service.EntryTypeDebit,
	}
	var matchEntry1 = service.Match{
		IndividualName: "Incorrect Name",
	}
	var actionCopy = service.Action{
		Copy: &service.Copy{
			Path: "/reconciliation",
		},
	}
	var actionReturn = service.Action{
		Return: &service.Return{
			Code: "R01",
		},
	}
	var actionCorrection = service.Action{
		Correction: &service.Correction{
			Code: "C04",
			Data: "Correct Name",
		},
	}
	var actionDelayReturn = actionReturn
	actionDelayReturn.Delay = &delay
	var actionDelayCorrection = actionCorrection
	actionDelayCorrection.Delay = &delay

	bh := ach.NewBatchHeader()
	bh.CompanyIdentification = "Classbook"
	bh.CompanyEntryDescription = "Payment"

	t.Run("No Match", func(t *testing.T) {
		var matcher Matcher
		matcher.Logger = log.NewTestLogger()
		matcher.Responses = []service.Response{}

		// Read our test file
		file, err := ach.ReadFile(filepath.Join("..", "..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, file)
		require.True(t, len(file.Batches) > 0)
		entries := file.Batches[0].GetEntries()

		// Find our Action
		copyAction, processAction := matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.Nil(t, processAction)

		// Find our Action
		copyAction, processAction = matcher.FindAction(bh, entries[1])
		require.Nil(t, copyAction)
		require.Nil(t, processAction)
	})

	t.Run("Match Copy only", func(t *testing.T) {
		var matcher Matcher
		matcher.Logger = log.NewTestLogger()
		matcher.Responses = []service.Response{
			{
				Match:  matchEntry1,
				Action: actionCopy,
			},
		}

		// Read our test file
		file, err := ach.ReadFile(filepath.Join("..", "..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, file)
		require.True(t, len(file.Batches) > 0)
		entries := file.Batches[0].GetEntries()

		// Find our Action
		copyAction, processAction := matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.Nil(t, processAction)

		// Find our Action
		copyAction, processAction = matcher.FindAction(bh, entries[1])
		require.NotNil(t, copyAction)
		require.Equal(t, actionCopy, *copyAction)
		require.Nil(t, processAction)
	})

	t.Run("Match Process only", func(t *testing.T) {
		var matcher Matcher
		matcher.Logger = log.NewTestLogger()
		matcher.Responses = []service.Response{
			{
				Match:  matchEntry1,
				Action: actionReturn,
			},
		}

		// Read our test file
		file, err := ach.ReadFile(filepath.Join("..", "..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, file)
		require.True(t, len(file.Batches) > 0)
		entries := file.Batches[0].GetEntries()

		// Find our Action
		copyAction, processAction := matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.Nil(t, processAction)

		// Find our Action
		copyAction, processAction = matcher.FindAction(bh, entries[1])
		require.Nil(t, copyAction)
		require.NotNil(t, processAction)
		require.Equal(t, actionReturn, *processAction)
	})

	t.Run("Match Copy + Process", func(t *testing.T) {
		var matcher Matcher
		matcher.Logger = log.NewTestLogger()
		matcher.Responses = []service.Response{
			{
				Match:  matchEntry1,
				Action: actionDelayCorrection,
			},
			{
				Match:  matchNone,
				Action: actionReturn,
			},
			{
				Match:  matchEntry1,
				Action: actionCopy,
			},
		}

		// Read our test file
		file, err := ach.ReadFile(filepath.Join("..", "..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, file)
		require.True(t, len(file.Batches) > 0)
		entries := file.Batches[0].GetEntries()

		// Find our Action
		copyAction, processAction := matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.Nil(t, processAction)

		// Find our Action
		copyAction, processAction = matcher.FindAction(bh, entries[1])
		require.NotNil(t, copyAction)
		require.Equal(t, actionCopy, *copyAction)
		require.NotNil(t, processAction)
		require.Equal(t, actionDelayCorrection, *processAction)
	})

	t.Run("Match BatchHeader Fields", func(t *testing.T) {
		var matcher Matcher
		matcher.Logger = log.NewTestLogger()
		matcher.Responses = []service.Response{}

		// Read our test file
		file, err := ach.ReadFile(filepath.Join("..", "..", "..", "testdata", "20230809-144155-102000021.ach"))
		require.NoError(t, err)
		require.NotNil(t, file)
		require.True(t, len(file.Batches) > 0)

		bh := file.Batches[0].GetHeader()
		entries := file.Batches[0].GetEntries()

		// Match no entries
		copyAction, processAction := matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.Nil(t, processAction)

		// Match based on CompanyID
		matcher.Responses = append(matcher.Responses, service.Response{
			Match: service.Match{
				CompanyIdentification: "Classbook",
			},
			Action: actionReturn,
		})
		copyAction, processAction = matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.NotNil(t, processAction)
		require.Equal(t, actionReturn, *processAction)

		// Match based on CompanyEntryDescription
		matcher.Responses = nil
		matcher.Responses = append(matcher.Responses, service.Response{
			Match: service.Match{
				CompanyEntryDescription: "Payment",
			},
			Action: actionReturn,
		})
		copyAction, processAction = matcher.FindAction(bh, entries[0])
		require.Nil(t, copyAction)
		require.NotNil(t, processAction)
		require.Equal(t, actionReturn, *processAction)
	})
}

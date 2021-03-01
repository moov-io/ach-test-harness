package response

import (
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type Matcher struct {
	Responses []service.Response
}

func (m Matcher) FindAction(ed *ach.EntryDetail) *service.Action {
	for i := range m.Responses {
		var matched bool // Matcher are AND'd together
		matcher := m.Responses[i].Match

		// Trace Number
		if matchesTraceNumber(matcher, ed) {
			matched = true
		}

		// Account Number
		if matchesAccountNumber(matcher, ed) {
			matched = true
		}

		// Check if the Amount matches
		if matchedAmount(matcher, ed) {
			matched = true
		}

		// Check if this Entry is a debit
		if matchedDebit(matcher, ed) {
			matched = true
		}

		// Return the Action if we've still matched
		if matched {
			return &m.Responses[i].Action
		}
	}
	return nil
}

func matchesTraceNumber(m service.Match, ed *ach.EntryDetail) bool {
	return (m.TraceNumber != "") && (ed.TraceNumber == m.TraceNumber)
}

func matchesAccountNumber(m service.Match, ed *ach.EntryDetail) bool {
	return (m.AccountNumber != "") && (strings.TrimSpace(ed.DFIAccountNumber) == m.AccountNumber)
}

func matchedAmount(m service.Match, ed *ach.EntryDetail) bool {
	if m.Amount != nil {
		var inner bool
		if m.Amount.Value != 0 {
			inner = (ed.Amount == m.Amount.Value)
		}
		if m.Amount.Min > 0 && m.Amount.Max > 0 {
			inner = inner && (m.Amount.Min <= ed.Amount && m.Amount.Max >= ed.Amount)
		}
		return inner
	}
	return false
}

func matchedDebit(m service.Match, ed *ach.EntryDetail) bool {
	if m.Debit != nil {
		switch ed.TransactionCode {
		case ach.CheckingDebit, ach.SavingsDebit, ach.GLDebit, ach.LoanDebit:
			return true
		}
	}
	return false
}

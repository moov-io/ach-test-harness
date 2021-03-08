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
		positive, negative := 0, 0 // Matchers are AND'd together
		matcher := m.Responses[i].Match

		// Trace Number
		if matcher.TraceNumber != "" {
			if matchesTraceNumber(matcher, ed) {
				positive++
			} else {
				negative++
			}
		}

		// Account Number
		if matcher.AccountNumber != "" {
			if matchesAccountNumber(matcher, ed) {
				positive++
			} else {
				negative++
			}
		}

		// Check if the Amount matches
		if matcher.Amount != nil {
			if matchedAmount(matcher, ed) {
				positive++
			} else {
				negative++
			}
		}

		// Check if this Entry is a debit
		if matcher.Debit != nil {
			if matchedDebit(matcher, ed) {
				positive++
			} else {
				negative++
			}
		}

		if matcher.IndividualName != "" {
			if matchedIndividualName(matcher, ed) {
				positive++
			} else {
				negative++
			}
		}

		// Return the Action if we've still matched
		if negative == 0 && positive > 0 {
			return &m.Responses[i].Action
		}
	}
	return nil
}

func matchesTraceNumber(m service.Match, ed *ach.EntryDetail) bool {
	return ed.TraceNumber == m.TraceNumber
}

func matchesAccountNumber(m service.Match, ed *ach.EntryDetail) bool {
	return strings.TrimSpace(ed.DFIAccountNumber) == m.AccountNumber
}

func matchedAmount(m service.Match, ed *ach.EntryDetail) bool {
	var inner bool
	if m.Amount.Value != 0 {
		inner = (ed.Amount == m.Amount.Value)
	}
	if m.Amount.Min > 0 && m.Amount.Max > 0 {
		inner = inner && (m.Amount.Min <= ed.Amount && m.Amount.Max >= ed.Amount)
	}
	return inner
}

func matchedDebit(m service.Match, ed *ach.EntryDetail) bool {
	switch ed.TransactionCode {
	case ach.CheckingDebit, ach.SavingsDebit, ach.GLDebit, ach.LoanDebit:
		return true
	}
	return false
}

func matchedIndividualName(m service.Match, ed *ach.EntryDetail) bool {
	return strings.TrimSpace(ed.IndividualName) == m.IndividualName
}

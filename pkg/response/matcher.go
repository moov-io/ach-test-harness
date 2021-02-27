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
		if matcher.TraceNumber != "" {
			matched = (ed.TraceNumber == matcher.TraceNumber)
		}

		// Account Number
		if matcher.AccountNumber != "" {
			matched = (strings.TrimSpace(ed.DFIAccountNumber) == matcher.AccountNumber)
		}

		// Check if the Amount matches
		if amt := m.Responses[i].Match.Amount; amt != nil {
			var inner bool
			if amt.Value != 0 {
				inner = (ed.Amount == amt.Value)
			}
			if amt.Min > 0 && amt.Max > 0 {
				inner = inner && (amt.Min <= ed.Amount && amt.Max >= ed.Amount)
			}
			matched = inner
		}

		// Check if this Entry is a debit
		if matcher.Debit != nil {
			switch ed.TransactionCode {
			case ach.CheckingDebit, ach.SavingsDebit, ach.GLDebit, ach.LoanDebit:
				matched = matched && true
			default:
				matched = false
			}
		}

		// Return the Action if we've still matched
		if matched {
			return &m.Responses[i].Action
		}
	}
	return nil
}

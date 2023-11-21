package match

import (
	"fmt"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/internal/achx"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
)

type Matcher struct {
	Debug  bool
	Logger log.Logger

	Responses []service.Response
}

func New(logger log.Logger, cfg service.Matching, responses []service.Response) Matcher {
	if cfg.Debug {
		logger.Info().Log("matcher: enable debug logging")
	}
	return Matcher{
		Debug:     cfg.Debug,
		Logger:    logger,
		Responses: responses,
	}
}

func (m Matcher) FindAction(ed *ach.EntryDetail) (copyAction *service.Action, processAction *service.Action) {
	/*
	 * See https://github.com/moov-io/ach-test-harness#config-schema for more details on how to configure.
	 */
	for i := range m.Responses {
		logger := m.Logger.With(log.Fields{
			"entry_trace_number": log.String(ed.TraceNumber),
		})
		logger.Log("starting EntryDetail matching")

		positive, negative := 0, 0 // Matchers are AND'd together
		matcher := m.Responses[i].Match
		action := m.Responses[i].Action

		if copyAction != nil && action.Copy != nil {
			continue // skip, we already have a copy action
		}
		if processAction != nil && action.Return != nil {
			continue // skip, we already have a process action
		}

		logger = logger.With(action)
		logger = logger.With(matcher)

		if m.Debug {
			logger = logger.With(log.Fields{
				"response_idx":         log.Int(i),
				"account_number":       log.String(matcher.AccountNumber),
				"entry_type":           log.String(string(matcher.EntryType)),
				"individual_name":      log.String(matcher.IndividualName),
				"routing_number":       log.String(matcher.RoutingNumber),
				"matcher_trace_number": log.String(matcher.TraceNumber),
			})
		}

		// Trace Number
		if matcher.TraceNumber != "" {
			if TraceNumber(matcher, ed) {
				if m.Debug {
					logger.Log("TraceNumber positive match")
				}
				positive++
			} else {
				if m.Debug {
					logger.Log("TraceNumber negative match")
				}
				negative++
			}
		}

		// Account Number
		if matcher.AccountNumber != "" {
			if AccountNumber(matcher, ed) {
				if m.Debug {
					logger.Log("DFIAccountNumber positive match")
				}
				positive++
			} else {
				if m.Debug {
					logger.Log("DFIAccountNumber negative match")
				}
				negative++
			}
		}

		// Routing Number
		if matcher.RoutingNumber != "" {
			if RoutingNumber(matcher, ed) {
				if m.Debug {
					logger.Log("RDFIIdentification positive match")
				}
				positive++
			} else {
				if m.Debug {
					logger.Log("RDFIIdentification negative match")
				}
				negative++
			}
		}

		// Check if the Amount matches
		if matcher.Amount != nil {
			if Amount(matcher, ed) {
				if m.Debug {
					logger.Log("Amount positive match")
				}
				positive++
			} else {
				if m.Debug {
					logger.Log("Amount negative match")
				}
				negative++
			}
		}

		// Check if this Entry is a debit
		if matcher.EntryType != service.EntryTypeEmpty {
			if matchedEntryType(matcher, ed) {
				if m.Debug {
					logger.Log("TransactionCode type positive match")
				}
				positive++
			} else {
				if m.Debug {
					logger.Log("TransactionCode type negative match")
				}
				negative++
			}
		}

		if matcher.IndividualName != "" {
			if matchedIndividualName(matcher, ed) {
				if m.Debug {
					logger.Log("IndividualName positive match")
				}
				positive++
			} else {
				if m.Debug {
					logger.Log("IndividualName negative match")
				}
				negative++
			}
		}

		// Return the Action if we've still matched
		logger.Logf("FINAL matching score negative=%d positive=%d", negative, positive)

		if negative == 0 && positive > 0 {
			// Action is valid, figure out where it belongs
			if m.Responses[i].Action.Copy != nil {
				copyAction = &m.Responses[i].Action
			} else {
				processAction = &m.Responses[i].Action
				// A non-Copy (process) Action with no Delay supersedes everything else
				if processAction.Delay == nil {
					return nil, processAction
				}
			}
		}
	}
	return
}

func TraceNumber(m service.Match, ed *ach.EntryDetail) bool {
	return ed.TraceNumber == m.TraceNumber
}

func AccountNumber(m service.Match, ed *ach.EntryDetail) bool {
	return strings.TrimSpace(ed.DFIAccountNumber) == m.AccountNumber
}

func RoutingNumber(m service.Match, ed *ach.EntryDetail) bool {
	aba8 := achx.ABA8(m.RoutingNumber) == ed.RDFIIdentification
	check := achx.ABACheckDigit(m.RoutingNumber) == ed.CheckDigit
	return aba8 && check
}

func Amount(m service.Match, ed *ach.EntryDetail) bool {
	var inner bool
	if m.Amount.Value != 0 {
		inner = (ed.Amount == m.Amount.Value)
	}
	if m.Amount.Min > 0 && m.Amount.Max > 0 {
		inner = (m.Amount.Min <= ed.Amount && m.Amount.Max >= ed.Amount)
	}
	return inner
}

func matchedEntryType(m service.Match, ed *ach.EntryDetail) bool {
	switch {
	case m.EntryType == service.EntryTypeDebit && matchedDebit(m, ed):
		return true
	case m.EntryType == service.EntryTypeCredit && matchedCredit(m, ed):
		return true
	case m.EntryType == service.EntryTypePrenote && matchedPrenote(m, ed):
		return true
	default:
		exists := m.EntryType != ""
		matches := string(m.EntryType) == fmt.Sprintf("%d", ed.TransactionCode)

		return exists && matches
	}
}

func matchedDebit(m service.Match, ed *ach.EntryDetail) bool {
	switch ed.TransactionCode {
	case ach.CheckingDebit, ach.SavingsDebit, ach.GLDebit, ach.LoanDebit:
		return true
	}
	return false
}

func matchedCredit(m service.Match, ed *ach.EntryDetail) bool {
	switch ed.TransactionCode {
	case ach.CheckingCredit, ach.SavingsCredit, ach.GLCredit, ach.LoanCredit:
		return true
	}
	return false
}

func matchedPrenote(m service.Match, ed *ach.EntryDetail) bool {
	switch ed.TransactionCode {
	case ach.CheckingPrenoteCredit, ach.SavingsPrenoteCredit, ach.GLPrenoteCredit, ach.LoanPrenoteCredit:
		return true
	case ach.CheckingPrenoteDebit, ach.SavingsPrenoteDebit, ach.GLPrenoteDebit:
		return true
	}
	return false
}

func matchedIndividualName(m service.Match, ed *ach.EntryDetail) bool {
	return strings.TrimSpace(ed.IndividualName) == m.IndividualName
}

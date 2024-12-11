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

func (m Matcher) FindAction(bh *ach.BatchHeader, ed *ach.EntryDetail) (copyAction *service.Action, processAction *service.Action) {
	/*
	 * See https://github.com/moov-io/ach-test-harness#config-schema for more details on how to configure.
	 */
	for idx, resp := range m.Responses {
		logger := m.Logger.With(log.Fields{
			"matcher.response_idx": log.Int(idx),
			"entry_trace_number":   log.String(ed.TraceNumber),
		})
		if m.Debug {
			logger.Info().Log("starting EntryDetail matching")
		}

		if copyAction != nil && resp.Action.Copy != nil {
			continue // skip, we already have a copy action
		}
		if processAction != nil && resp.Action.Return != nil {
			continue // skip, we already have a process action
		}

		// Run .Match and .Not matchers
		positiveMatches, negativeMatches := m.runMatchers(logger, resp.Match, bh, ed)

		// The Not matchers need to be inverted from the affirmative
		notPositiveMatches, notNegativeMatches := m.runMatchers(logger, resp.Not, bh, ed)

		// Add the affirmative positive matches with the populated Not matches
		positive := len(positiveMatches) + len(notNegativeMatches)

		// Add the affirmative negative matches with the populated Not negative matches
		negative := len(negativeMatches) + len(notPositiveMatches)

		// Return the Action if we've still matched
		if negative == 0 && positive > 0 {
			// Action is valid, figure out where it belongs
			if resp.Action.Copy != nil {
				copyAction = &resp.Action
			} else {
				processAction = &resp.Action
				// A non-Copy (process) Action with no Delay supersedes everything else
				if processAction.Delay == nil {
					return nil, processAction
				}
			}
		}
	}

	return
}

func (m Matcher) runMatchers(logger log.Logger, matcher service.Match, bh *ach.BatchHeader, ed *ach.EntryDetail) (positiveMatchers, negativeMatchers []string) {
	logger = logger.With(matcher)

	if m.Debug {
		logger = logger.With(log.Fields{
			"matcher.account_number":  log.String(matcher.AccountNumber),
			"matcher.entry_type":      log.String(string(matcher.EntryType)),
			"matcher.individual_name": log.String(matcher.IndividualName),
			"matcher.routing_number":  log.String(matcher.RoutingNumber),
			"matcher.trace_number":    log.String(matcher.TraceNumber),
			"ed.account_number":       log.String(ed.DFIAccountNumber),
			"ed.entry_type":           log.String(fmt.Sprintf("%d", ed.TransactionCode)),
			"ed.individual_name":      log.String(ed.IndividualName),
			"ed.routing_number":       log.String(ed.RDFIIdentification + ed.CheckDigit),
			"ed.trace_number":         log.String(ed.TraceNumber),
			"ed.amount":               log.String(fmt.Sprintf("%d", ed.Amount)),
		})
	}

	// Trace Number
	if matcher.TraceNumber != "" {
		if TraceNumber(matcher, ed) {
			positiveMatchers = append(positiveMatchers, "TraceNumber")
		} else {
			negativeMatchers = append(negativeMatchers, "TraceNumber")
		}
	}

	// Account Number
	if matcher.AccountNumber != "" {
		if AccountNumber(matcher, ed) {
			positiveMatchers = append(positiveMatchers, "DFIAccountNumber")
		} else {
			negativeMatchers = append(negativeMatchers, "DFIAccountNumber")
		}
	}

	// Routing Number
	if matcher.RoutingNumber != "" {
		if RoutingNumber(matcher, ed) {
			positiveMatchers = append(positiveMatchers, "RDFIIdentification")
		} else {
			negativeMatchers = append(negativeMatchers, "RDFIIdentification")
		}
	}

	// Check if the Amount matches
	if matcher.Amount != nil {
		if Amount(matcher, ed) {
			positiveMatchers = append(positiveMatchers, "Amount")
		} else {
			negativeMatchers = append(negativeMatchers, "Amount")
		}
	}

	// Check if this Entry is a debit
	if matcher.EntryType != service.EntryTypeEmpty {
		if matchedEntryType(matcher, ed) {
			positiveMatchers = append(positiveMatchers, "TransactionCode")
		} else {
			negativeMatchers = append(negativeMatchers, "TransactionCode")
		}
	}

	if matcher.IndividualName != "" {
		if matchedIndividualName(matcher, ed) {
			positiveMatchers = append(positiveMatchers, "IndividualName")
		} else {
			negativeMatchers = append(negativeMatchers, "IndividualName")
		}
	}

	// BatchHeader fields
	if matcher.CompanyIdentification != "" {
		if matchedCompanyIdentification(matcher, bh) {
			positiveMatchers = append(positiveMatchers, "CompanyIdentification")
		} else {
			negativeMatchers = append(negativeMatchers, "CompanyIdentification")
		}
	}
	if matcher.CompanyEntryDescription != "" {
		if matchedCompanyEntryDescription(matcher, bh) {
			positiveMatchers = append(positiveMatchers, "CompanyEntryDescription")
		} else {
			negativeMatchers = append(negativeMatchers, "CompanyEntryDescription")
		}
	}

	// format the list of negative and positive matchers for logging
	var b strings.Builder

	b.WriteString(fmt.Sprintf("FINAL matching score negative=%d", len(negativeMatchers)))
	if len(negativeMatchers) > 0 {
		b.WriteString(fmt.Sprintf(" (%s)", strings.Join(negativeMatchers, ", ")))
	}

	b.WriteString(fmt.Sprintf(" positive=%d", len(positiveMatchers)))
	if len(positiveMatchers) > 0 {
		b.WriteString(fmt.Sprintf(" (%s)", strings.Join(positiveMatchers, ", ")))
	}

	if m.Debug {
		logger.Log(b.String())
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

func matchedCompanyIdentification(m service.Match, bh *ach.BatchHeader) bool {
	return strings.EqualFold(strings.TrimSpace(m.CompanyIdentification), strings.TrimSpace(bh.CompanyIdentification))
}

func matchedCompanyEntryDescription(m service.Match, bh *ach.BatchHeader) bool {
	return strings.EqualFold(strings.TrimSpace(m.CompanyEntryDescription), strings.TrimSpace(bh.CompanyEntryDescription))
}

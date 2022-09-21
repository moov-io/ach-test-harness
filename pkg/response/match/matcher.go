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

func (m Matcher) FindAction(ed *ach.EntryDetail) *service.Action {
	m.debugLog(fmt.Sprintf("matching EntryDetail TraceNumber=%s", ed.TraceNumber))

	for i := range m.Responses {
		positive, negative := 0, 0 // Matchers are AND'd together
		matcher := m.Responses[i].Match
		action := m.Responses[i].Action

		var copyPath string
		var correctionCode string
		var correctionData string
		var returnCode string
		var amount int

		// Safely retrieve several values that are needed for the debug log below
		if action.Copy != nil {
			copyPath = action.Copy.Path
		}

		if action.Correction != nil {
			correctionCode = action.Correction.Code
			correctionData = action.Correction.Data
		}

		if action.Return != nil {
			returnCode = action.Return.Code
		}

		if matcher.Amount != nil {
			amount = matcher.Amount.Value
		}

		m.debugLog(
			fmt.Sprintf(
				"attempting matcher resp[%d]= AccountNumber: %s, Amount: %d, EntryType: %s, IndividualName: %s, RoutingNumber: %s, TraceNumber: %s, CopyPath: %s, CorrectionCode: %s, CorrectionData: %s, ReturnCode: %s",
				i,
				matcher.AccountNumber,
				amount,
				string(matcher.EntryType),
				matcher.IndividualName,
				matcher.RoutingNumber,
				matcher.TraceNumber,
				copyPath,
				correctionCode,
				correctionData,
				returnCode))

		// Trace Number
		if matcher.TraceNumber != "" {
			if TraceNumber(matcher, ed) {
				m.debugLog(fmt.Sprintf("EntryDetail.TraceNumber=%s positive match", ed.TraceNumber))
				positive++
			} else {
				m.debugLog(fmt.Sprintf("EntryDetail.TraceNumber=%s negative match", ed.TraceNumber))
				negative++
			}
		}

		// Account Number
		if matcher.AccountNumber != "" {
			if AccountNumber(matcher, ed) {
				m.debugLog("EntryDetail.DFIAccountNumber positive match")
				positive++
			} else {
				m.debugLog("EntryDetail.DFIAccountNumber negative match")
				negative++
			}
		}

		// Routing Number
		if matcher.RoutingNumber != "" {
			if RoutingNumber(matcher, ed) {
				m.debugLog(fmt.Sprintf("EntryDetail.RDFIIdentification=%s positive match", ed.RDFIIdentification))
				positive++
			} else {
				m.debugLog(fmt.Sprintf("EntryDetail=%s negative match", ed.RDFIIdentification))
				negative++
			}
		}

		// Check if the Amount matches
		if matcher.Amount != nil {
			if Amount(matcher, ed) {
				m.debugLog(fmt.Sprintf("EntryDetail.Amount=%d positive match", ed.Amount))
				positive++
			} else {
				m.debugLog(fmt.Sprintf("EntryDetail.Amount=%d negative match", ed.Amount))
				negative++
			}
		}

		// Check if this Entry is a debit
		if matcher.EntryType != service.EntryTypeEmpty {
			if matchedEntryType(matcher, ed) {
				m.debugLog(fmt.Sprintf("EntryDetail.TransactionCode=%d type positive match", ed.TransactionCode))
				positive++
			} else {
				m.debugLog(fmt.Sprintf("EntryDetail.TransactionCode=%d type negative match", ed.TransactionCode))
				negative++
			}
		}

		if matcher.IndividualName != "" {
			if matchedIndividualName(matcher, ed) {
				m.debugLog(fmt.Sprintf("EntryDetail.IndividualName='%s' positive match", ed.IndividualName))
				positive++
			} else {
				m.debugLog(fmt.Sprintf("EntryDetail.IndividualName='%s' negative match", ed.IndividualName))
				negative++
			}
		}

		// Return the Action if we've still matched
		m.debugLog(fmt.Sprintf("FINAL EntryDetail.TraceNumber=%s score negative=%d positive=%d\n", ed.TraceNumber, negative, positive))
		if negative == 0 && positive > 0 {
			return &m.Responses[i].Action
		}
	}
	return nil
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
	case m.EntryType == service.EntryTypeCredit && !matchedDebit(m, ed):
		return true
	default:
		return false
	}
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

func (m Matcher) debugLog(msg string) {
	if m.Debug {
		m.Logger.Info().Log(msg)
	}
}

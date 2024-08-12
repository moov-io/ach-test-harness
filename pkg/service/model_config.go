// generated-from:d5d5aa0731228b23a10b49fa2c69819a212e4565205c0258432d2b35dba9f169 DO NOT REMOVE, DO UPDATE

package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"
)

type GlobalConfig struct {
	ACHTestHarness Config
}

func (gc *GlobalConfig) Validate() error {
	return gc.ACHTestHarness.Validate()
}

// Config defines all the configuration for the app
type Config struct {
	Servers   ServerConfig
	Telemetry telemetry.Config

	ValidateOpts *ach.ValidateOpts
	Matching     Matching
	Responses    []Response
}

func (cfg *Config) Validate() error {
	for i := range cfg.Responses {
		if err := cfg.Responses[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *Config) responsePaths() []string {
	var out []string
	for i := range cfg.Responses {
		if cfg.Responses[i].Action.Copy != nil {
			out = append(out, cfg.Responses[i].Action.Copy.Path)
		}
	}
	return out
}

// ServerConfig - Groups all the http configs for the servers and ports that get opened.
type ServerConfig struct {
	FTP   *FTPConfig
	Admin HTTPConfig
}

// FTPConfig configuration for running an FTP server
type FTPConfig struct {
	RootPath     string
	Hostname     string
	Auth         FTPAuth
	Port         int
	PassivePorts string
	Paths        Paths
}

type FTPAuth struct {
	Username string
	Password string
}

type Paths struct {
	// Incoming Files
	Files string

	// Outgoing Files
	Return string
}

// HTTPConfig configuration for running an http server
type HTTPConfig struct {
	Bind BindAddress
}

// BindAddress specifies where the http server should bind to.
type BindAddress struct {
	Address string
}

type Matching struct {
	Debug bool
}

type Response struct {
	Match  Match
	Action Action
}

func (r *Response) Validate() error {
	if r.Match.Empty() {
		return errors.New("no Match configured")
	}
	return r.Action.Validate()
}

type Match struct {
	AccountNumber  string
	Amount         *Amount
	EntryType      EntryType
	IndividualName string
	RoutingNumber  string
	TraceNumber    string
}

func (m Match) Context() map[string]log.Valuer {
	logFields := log.Fields{}

	if m.Amount != nil {
		var amount = m.Amount.Value
		logFields["matcher.amount"] = log.Int(amount)
	}

	return logFields
}

func (m Match) Empty() bool {
	return m.AccountNumber == "" && m.Amount.Empty() &&
		string(m.EntryType) == "" && m.IndividualName == "" &&
		m.RoutingNumber == "" && m.TraceNumber == ""
}

type Amount struct {
	Value int
	Min   int
	Max   int
}

func (a *Amount) Empty() bool {
	if a == nil {
		return true
	}
	return a.Value == 0 && a.Min == 0 && a.Max == 0
}

type EntryType string

const (
	EntryTypeEmpty   EntryType = ""
	EntryTypeDebit   EntryType = "debit"
	EntryTypeCredit  EntryType = "credit"
	EntryTypePrenote EntryType = "prenote"
)

type Action struct {
	Delay      *time.Duration // e.g. "12h" or "10s"
	Copy       *Copy
	Correction *Correction
	Return     *Return
}

func (a Action) Context() map[string]log.Valuer {
	logFields := log.Fields{}

	// Safely retrieve several values that are needed for the debug log below
	if a.Delay != nil {
		var delayTime = a.Delay.String()
		logFields["delay"] = log.String(delayTime)
	}

	if a.Copy != nil {
		var copyPath = a.Copy.Path
		logFields["copy_path"] = log.String(copyPath)
	}

	if a.Correction != nil {
		var correctionCode = a.Correction.Code
		var correctionData = a.Correction.Data
		logFields["correction_code"] = log.String(correctionCode)
		logFields["correction_data"] = log.String(correctionData)
	}

	if a.Return != nil {
		var returnCode = a.Return.Code
		logFields["return_code"] = log.String(returnCode)
	}

	return logFields
}

func (a *Action) Validate() error {
	// Delay is only valid for Return and Correction
	if a.Delay != nil && a.Copy != nil {
		return errors.New("Delay and Copy are not valid together in an Action")
	}

	// only allowed 1 of Copy, Return, Correction to be configured
	var count = 0
	if a.Copy != nil {
		count++
	}
	if a.Return != nil {
		count++
	}
	if a.Correction != nil {
		count++
	}
	if count > 1 {
		return errors.New("only 1 of Copy, Return, Correction can be configured in an Action")
	}
	if a.Delay != nil && count == 0 {
		return errors.New("either Return or Correction is required if Delay is set")
	}

	return nil
}

type Copy struct {
	Path string
}

type Correction struct {
	Code string
	Data string
}

type Return struct {
	Code string
}

func (r Return) Validate() error {
	if r.Code == "" {
		return nil
	}
	if code := ach.LookupReturnCode(r.Code); code != nil {
		return nil
	}
	return fmt.Errorf("unexpected return code %s", r.Code)
}

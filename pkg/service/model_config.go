// generated-from:d5d5aa0731228b23a10b49fa2c69819a212e4565205c0258432d2b35dba9f169 DO NOT REMOVE, DO UPDATE

package service

import (
	"fmt"

	"github.com/moov-io/ach"
)

type GlobalConfig struct {
	ACHTestHarness Config
}

// Config defines all the configuration for the app
type Config struct {
	Servers      ServerConfig
	ValidateOpts *ach.ValidateOpts
	Matching     Matching
	Responses    []Response
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
	Action *Action
	Future *Future
}

type Match struct {
	AccountNumber  string
	Amount         *Amount
	EntryType      EntryType
	IndividualName string
	RoutingNumber  string
	TraceNumber    string
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
	Copy       *Copy
	Correction *Correction
	Return     *Return
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

type Future struct {
	Delay      string // ex: "12h"
	Correction *Correction
	Return     *Return
}

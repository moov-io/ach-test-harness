// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package achx

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"unicode/utf8"

	"github.com/moov-io/ach"
)

// TraceNumber returns a trace number from a given routing number
// and uses a hidden random generator. These values are not expected
// to be cryptographically secure.
//
// We decided not to increment trace numbers from a shared counter because
// it will create a lot of "NNNNN00001" trace numbers that are duplicated
// and hard to uniquely identify a Transfer.
func TraceNumber(routingNumber string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1e15))
	if err != nil {
		return "", fmt.Errorf("ERROR creating trace number: %w", err)
	}
	v := fmt.Sprintf("%s%s", ABA8(routingNumber), n.String())
	if utf8.RuneCountInString(v) > 15 {
		return v[:15], nil
	}
	return v, nil
}

// ABA8 returns the first 8 digits of an ABA routing number.
// If the input is invalid then an empty string is returned.
func ABA8(rtn string) string {
	n := utf8.RuneCountInString(rtn)
	if n == 10 {
		return rtn[1:9] // ACH server will prefix with space, 0, or 1
	}
	if n != 8 && n != 9 {
		return ""
	}
	return rtn[:8]
}

// ABACheckDigit returns the last digit of an ABA routing number.
// If the input is invalid then an empty string is returned.
func ABACheckDigit(rtn string) string {
	n := utf8.RuneCountInString(rtn)
	if n == 10 {
		return rtn[9:] // ACH server will prefix with space, 0, or 1
	}
	if n == 8 {
		return strconv.Itoa(ach.CalculateCheckDigit(rtn))
	}
	if n != 8 && n != 9 {
		return ""
	}
	return rtn[8:9]
}

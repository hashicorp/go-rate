// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"time"
)

// LimitPer identifies how a limit is allocated.
type LimitPer string

func (p LimitPer) String() string {
	return string(p)
}

// IsValid checks if the given LimitPer is valid.
func (p LimitPer) IsValid() bool {
	switch p {
	case LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken:
		return true
	}
	return false
}

const (
	// LimitPerIPAddress indicates that the limit applies per IP address.
	LimitPerIPAddress LimitPer = "ip-address"
	// LimitPerAuthToken indicates that the limit applies per auth token.
	LimitPerAuthToken LimitPer = "auth-token"
	// LimitPerTotal indicates that the limit applies for all IP address and all Auth Tokens.
	LimitPerTotal LimitPer = "total"
)

// Limit defines the number of requests that can be made to perform an action
// against a resource in a time period, allocated per IP address, auth token,
// or in total.
type Limit struct {
	Resource string
	Action   string
	Per      LimitPer

	Unlimited bool

	MaxRequests uint64
	Period      time.Duration
}

// IsValid checks if the given Limit is valid. A Limit can either be
// "unlimited" or have a max requests and period defined. Therefore, it is
// considered invalid if Unlimited is true and has a non-zero MaxRequests
// and/or Period. Likewise, it is invalid if it has a zero MaxRequests and/or
// Period and Unlimited is false. Finally, the Limit must have a valid
// LimitPer.
func (l *Limit) IsValid() bool {
	if !(l.Per.IsValid()) {
		return false
	}

	switch {
	case l.Unlimited && (l.MaxRequests != 0 || l.Period != 0):
		return false
	case !l.Unlimited && (l.MaxRequests == 0 || l.Period <= 0):
		return false
	}

	return true
}

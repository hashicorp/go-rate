// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidLimitPer(t *testing.T) {
	cases := []struct {
		name string
		in   LimitPer
		want bool
	}{
		{
			LimitPerTotal.String(),
			LimitPerTotal,
			true,
		},
		{
			LimitPerIPAddress.String(),
			LimitPerIPAddress,
			true,
		},
		{
			LimitPerAuthToken.String(),
			LimitPerAuthToken,
			true,
		},
		{
			"Invalid",
			LimitPer("invalid"),
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.IsValid()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestValidLimit(t *testing.T) {
	cases := []struct {
		name string
		in   Limit
		err  error
	}{
		{
			"Valid_TotalMaxRequests",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			nil,
		},
		{
			"Valid_IPAddressMaxRequests",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerIPAddress,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			nil,
		},
		{
			"Valid_AuthTokenMaxRequests",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerAuthToken,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			nil,
		},
		{
			"Valid_TotalUnlimited",
			&Unlimited{
				Resource: "resource",
				Action:   "action",
				Per:      LimitPerTotal,
			},
			nil,
		},
		{
			"Valid_IPAddressUnlimited",
			&Unlimited{
				Resource: "resource",
				Action:   "action",
				Per:      LimitPerIPAddress,
			},
			nil,
		},
		{
			"Valid_AuthTokenUnlimited",
			&Unlimited{
				Resource: "resource",
				Action:   "action",
				Per:      LimitPerAuthToken,
			},
			nil,
		},
		{
			"Invalid_LimitPerMaxRequests",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPer("invalid"),
				MaxRequests: 10,
				Period:      time.Minute,
			},
			ErrInvalidLimitPer,
		},
		{
			"Invalid_LimitPerUnlimited",
			&Unlimited{
				Resource: "resource",
				Action:   "action",
				Per:      LimitPer("invalid"),
			},
			ErrInvalidLimitPer,
		},
		{
			"Invalid_MaxRequestsZeroPeriod",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 50,
				Period:      0,
			},
			ErrInvalidLimit,
		},
		{
			"Invalid_MaxRequestsNegativePeriod",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 50,
				Period:      time.Second * -1,
			},
			ErrInvalidLimit,
		},
		{
			"Invalid_ZeroMaxRequestsPeriod",
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 0,
				Period:      time.Minute,
			},
			ErrInvalidLimit,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.validate()
			assert.ErrorIs(t, got, tc.err)
		})
	}
}

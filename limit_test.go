// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate_test

import (
	"testing"
	"time"

	"github.com/hashicorp/go-rate"
	"github.com/stretchr/testify/assert"
)

func TestValidLimitPer(t *testing.T) {
	cases := []struct {
		name string
		in   rate.LimitPer
		want bool
	}{
		{
			rate.LimitPerTotal.String(),
			rate.LimitPerTotal,
			true,
		},
		{
			rate.LimitPerIPAddress.String(),
			rate.LimitPerIPAddress,
			true,
		},
		{
			rate.LimitPerAuthToken.String(),
			rate.LimitPerAuthToken,
			true,
		},
		{
			"Invalid",
			rate.LimitPer("invalid"),
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
		in   *rate.Limit
		want bool
	}{
		{
			"Valid_TotalMaxRequests",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   false,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			true,
		},
		{
			"Valid_IPAddressMaxRequests",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerIPAddress,
				Unlimited:   false,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			true,
		},
		{
			"Valid_AuthTokenMaxRequests",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerAuthToken,
				Unlimited:   false,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			true,
		},
		{
			"Valid_TotalUnlimited",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   true,
				MaxRequests: 0,
				Period:      0,
			},
			true,
		},
		{
			"Valid_IPAddressUnlimited",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerIPAddress,
				Unlimited:   true,
				MaxRequests: 0,
				Period:      0,
			},
			true,
		},
		{
			"Valid_AuthTokenUnlimited",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerAuthToken,
				Unlimited:   true,
				MaxRequests: 0,
				Period:      0,
			},
			true,
		},
		{
			"Invalid_LimitPerMaxRequests",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPer("invalid"),
				Unlimited:   false,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			false,
		},
		{
			"Invalid_LimitPerUnlimited",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPer("invalid"),
				Unlimited:   true,
				MaxRequests: 0,
				Period:      0,
			},
			false,
		},
		{
			"Invalid_MaxRequestsZeroPeriod",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   false,
				MaxRequests: 50,
				Period:      0,
			},
			false,
		},
		{
			"Invalid_MaxRequestsNegativePeriod",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   false,
				MaxRequests: 50,
				Period:      time.Second * -1,
			},
			false,
		},
		{
			"Invalid_ZeroMaxRequestsPeriod",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   false,
				MaxRequests: 0,
				Period:      time.Minute,
			},
			false,
		},
		{
			"Invalid_UnlimitedMaxRequests",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   true,
				MaxRequests: 50,
				Period:      0,
			},
			false,
		},
		{
			"Invalid_UnlimitedPeriod",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   true,
				MaxRequests: 0,
				Period:      time.Minute,
			},
			false,
		},
		{
			"Invalid_UnlimitedMaxRequestsPeriod",
			&rate.Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         rate.LimitPerTotal,
				Unlimited:   true,
				MaxRequests: 50,
				Period:      time.Minute,
			},
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

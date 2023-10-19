// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLimiter(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		maxSize        int
		limits         []*Limit
		options        []Option
		expectErr      error
		expectPolicies map[string]*limitPolicy
	}{
		{
			"ValidPolicy",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			nil,
			map[string]*limitPolicy{
				"resource:action": {
					resource: "resource",
					action:   "action",
					m: map[LimitPer]*Limit{
						LimitPerTotal: {
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerTotal,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
						LimitPerIPAddress: {
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerIPAddress,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
						LimitPerAuthToken: {
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
					},
					policy: `100;w=60;comment="total", 100;w=60;comment="ip-address", 100;w=60;comment="auth-token"`,
				},
			},
		},
		{
			"MultiplePolicies",
			10,
			[]*Limit{
				{
					Resource:    "resource1",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource1",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource1",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			nil,
			map[string]*limitPolicy{
				"resource1:action": {
					resource: "resource1",
					action:   "action",
					m: map[LimitPer]*Limit{
						LimitPerTotal: {
							Resource:    "resource1",
							Action:      "action",
							Per:         LimitPerTotal,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
						LimitPerIPAddress: {
							Resource:    "resource1",
							Action:      "action",
							Per:         LimitPerIPAddress,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
						LimitPerAuthToken: {
							Resource:    "resource1",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
					},
					policy: `100;w=60;comment="total", 100;w=60;comment="ip-address", 100;w=60;comment="auth-token"`,
				},
				"resource2:action": {
					resource: "resource2",
					action:   "action",
					m: map[LimitPer]*Limit{
						LimitPerTotal: {
							Resource:    "resource2",
							Action:      "action",
							Per:         LimitPerTotal,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
						LimitPerIPAddress: {
							Resource:    "resource2",
							Action:      "action",
							Per:         LimitPerIPAddress,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
						LimitPerAuthToken: {
							Resource:    "resource2",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 100,
							Period:      time.Minute,
						},
					},
					policy: `100;w=60;comment="total", 100;w=60;comment="ip-address", 100;w=60;comment="auth-token"`,
				},
			},
		},
		{
			"OneLimit",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			ErrInvalidLimitPolicy,
			nil,
		},
		{
			"MultipleLimits",
			10,
			[]*Limit{
				{
					Resource:    "res2",
					Action:      "action2",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Second,
				},
				{
					Resource:    "res1",
					Action:      "action1",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			ErrInvalidLimitPolicy,
			nil,
		},
		{
			"DuplicateLimits",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Second,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 10,
					Period:      time.Minute,
				},
			},
			[]Option{},
			ErrDuplicateLimit,
			nil,
		},
		{
			"InvalidLimit",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   true,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			ErrInvalidLimit,
			nil,
		},
		{
			"NoLimits",
			10,
			[]*Limit{},
			[]Option{},
			ErrEmptyLimits,
			nil,
		},
		{
			"InvalidMaxSize",
			0,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			ErrInvalidMaxSize,
			nil,
		},
		{
			"InvalidNumberBuckets",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{WithNumberBuckets(0)},
			ErrInvalidNumberBuckets,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := NewLimiter(tc.limits, tc.maxSize, tc.options...)
			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, l)
			assert.Equal(t, l.policies, tc.expectPolicies)
		})
	}
}

type allowTestRequest struct {
	resource  string
	action    string
	ip        string
	authToken string

	expectAllowed bool
	expectErr     error
	expectQuota   *Quota
}

func TestLimiterAllow(t *testing.T) {
	cases := []struct {
		name    string
		maxSize int
		limits  []*Limit
		options []Option
		reqs    []allowTestRequest
	}{
		{
			"OneRequest",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 50,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 25,
					Period:      time.Minute,
				},
			},
			[]Option{},
			[]allowTestRequest{
				{
					resource:      "resource",
					action:        "action",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 25,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
			},
		},
		{
			"MissingLimitPolicy",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 50,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 25,
					Period:      time.Minute,
				},
			},
			[]Option{},
			[]allowTestRequest{
				{
					resource:      "missing",
					action:        "missing",
					expectAllowed: false,
					expectErr:     ErrLimitPolicyNotFound,
					expectQuota:   nil,
				},
			},
		},
		{
			"ConsumeQuota",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 50,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 2,
					Period:      time.Minute,
				},
			},
			[]Option{},
			[]allowTestRequest{
				{
					resource:      "resource",
					action:        "action",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 2,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
				{
					resource:      "resource",
					action:        "action",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 2,
							Period:      time.Minute,
						},
						used: 2,
					},
				},
				// Quota should be consumed, so this should not be allowed
				{
					resource:      "resource",
					action:        "action",
					expectAllowed: false,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 2,
							Period:      time.Minute,
						},
						used: 2,
					},
				},
			},
		},
		{
			"ReachedCapacity",
			6,
			[]*Limit{
				{
					Resource:    "resource1",
					Action:      "action1",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action2",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource3",
					Action:      "action3",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource1",
					Action:      "action1",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 50,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action2",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 50,
					Period:      time.Minute,
				},
				{
					Resource:    "resource3",
					Action:      "action3",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 50,
					Period:      time.Minute,
				},
				{
					Resource:    "resource1",
					Action:      "action1",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 25,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action2",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 1,
					Period:      time.Minute,
				},
				{
					Resource:    "resource3",
					Action:      "action3",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 2,
					Period:      time.Minute,
				},
			},
			[]Option{},
			[]allowTestRequest{
				{
					resource:      "resource1",
					action:        "action1",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource1",
							Action:      "action1",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 25,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
				{
					resource:      "resource2",
					action:        "action2",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource2",
							Action:      "action2",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 1,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
				// Out of space to store quotas so request for new quotas should
				// not be allowed
				{
					resource:      "resource3",
					action:        "action3",
					expectAllowed: false,
					expectErr:     &ErrLimiterFull{RetryIn: (time.Minute / time.Duration(DefaultNumberBuckets-1))},
					expectQuota:   nil,
				},
				// However, requests for quotas already in the store should
				// still be allowed if the quota is not consumed
				{
					resource:      "resource1",
					action:        "action1",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource1",
							Action:      "action1",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 25,
							Period:      time.Minute,
						},
						used: 2,
					},
				},
				{
					resource:      "resource2",
					action:        "action2",
					expectAllowed: false,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource2",
							Action:      "action2",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 1,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
			},
		},
		{
			"MultipleIPAddress",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 3,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 2,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 1,
					Period:      time.Minute,
				},
			},
			[]Option{},
			[]allowTestRequest{
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token1",
					ip:            "ip1",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 1,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token2",
					ip:            "ip2",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 1,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token3",
					ip:            "ip3",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerTotal,
							Unlimited:   false,
							MaxRequests: 3,
							Period:      time.Minute,
						},
						used: 3,
					},
				},
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token4",
					ip:            "ip4",
					expectAllowed: false,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerTotal,
							Unlimited:   false,
							MaxRequests: 3,
							Period:      time.Minute,
						},
						used: 3,
					},
				},
			},
		},
		{
			"MultipleAuthTokens",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 2,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 1,
					Period:      time.Minute,
				},
			},
			[]Option{},
			[]allowTestRequest{
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token1",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerAuthToken,
							Unlimited:   false,
							MaxRequests: 1,
							Period:      time.Minute,
						},
						used: 1,
					},
				},
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token2",
					expectAllowed: true,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerIPAddress,
							Unlimited:   false,
							MaxRequests: 2,
							Period:      time.Minute,
						},
						used: 2,
					},
				},
				{
					resource:      "resource",
					action:        "action",
					authToken:     "token3",
					expectAllowed: false,
					expectErr:     nil,
					expectQuota: &Quota{
						limit: &Limit{
							Resource:    "resource",
							Action:      "action",
							Per:         LimitPerIPAddress,
							Unlimited:   false,
							MaxRequests: 2,
							Period:      time.Minute,
						},
						used: 2,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := NewLimiter(tc.limits, tc.maxSize, tc.options...)
			require.NoError(t, err)
			require.NotNil(t, l)

			for _, r := range tc.reqs {
				allowed, q, err := l.Allow(r.resource, r.action, r.ip, r.authToken)
				if r.expectErr != nil {
					require.EqualError(t, err, r.expectErr.Error())
					assert.Equal(t, r.expectAllowed, allowed)
					if want, ok := r.expectErr.(*ErrLimiterFull); ok {
						got, ok := err.(*ErrLimiterFull)
						assert.True(t, ok, "did not get an ErrLimiterFull error")
						assert.Equal(t, want.RetryIn, got.RetryIn)
					}
					continue
				}

				require.NoError(t, err)
				assert.Equal(t, r.expectAllowed, allowed)
				assert.Equal(t, r.expectQuota.limit, q.limit)
				assert.Equal(t, r.expectQuota.used, q.used)
				assert.Equal(t, r.expectQuota.Remaining(), q.Remaining())
			}
		})
	}
}

func TestSetPolicyHeader(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		maxSize           int
		limits            []*Limit
		options           []Option
		resource          string
		action            string
		expectErr         error
		expectHeader      string
		expectHeaderValue string
	}{
		{
			"ValidPolicy",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			"resource",
			"action",
			nil,
			DefaultPolicyHeader,
			`100;w=60;comment="total", 100;w=60;comment="ip-address", 100;w=60;comment="auth-token"`,
		},
		{
			"ValidPolicyAlternateHeader",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{WithPolicyHeader("Policy-Header")},
			"resource",
			"action",
			nil,
			"Policy-Header",
			`100;w=60;comment="total", 100;w=60;comment="ip-address", 100;w=60;comment="auth-token"`,
		},
		{
			"PolicyNotFound",
			10,
			[]*Limit{
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			"resource-missing",
			"action",
			ErrLimitPolicyNotFound,
			DefaultPolicyHeader,
			``,
		},
		{
			"Unlimited",
			10,
			[]*Limit{
				{
					Resource:  "resource",
					Action:    "action",
					Per:       LimitPerTotal,
					Unlimited: true,
				},
				{
					Resource:  "resource",
					Action:    "action",
					Per:       LimitPerIPAddress,
					Unlimited: true,
				},
				{
					Resource:  "resource",
					Action:    "action",
					Per:       LimitPerAuthToken,
					Unlimited: true,
				},
				{
					Resource:    "resource2",
					Action:      "action",
					Per:         LimitPerTotal,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action",
					Per:         LimitPerIPAddress,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
				{
					Resource:    "resource2",
					Action:      "action",
					Per:         LimitPerAuthToken,
					Unlimited:   false,
					MaxRequests: 100,
					Period:      time.Minute,
				},
			},
			[]Option{},
			"resource",
			"action",
			nil,
			DefaultPolicyHeader,
			``,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := NewLimiter(tc.limits, tc.maxSize, tc.options...)
			require.NoError(t, err)
			require.NotNil(t, l)

			h := make(http.Header)
			err = l.SetPolicyHeader(tc.resource, tc.action, h)
			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
				return
			}
			got := h.Get(tc.expectHeader)
			assert.Equal(t, tc.expectHeaderValue, got)
		})
	}
}

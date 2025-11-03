// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimitPolicy_add(t *testing.T) {
	cases := []struct {
		name        string
		limitPolicy *limitPolicy
		add         Limit
		expectErr   error
	}{
		{
			"NoError",
			newLimitPolicy("resource", "action"),
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
			"InvalidLimit",
			newLimitPolicy("resource", "action"),
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 0,
				Period:      time.Minute,
			},
			ErrInvalidLimit,
		},
		{
			"IncorrectResource",
			newLimitPolicy("resource", "action"),
			&Limited{
				Resource:    "resource1",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			ErrInvalidLimit,
		},
		{
			"IncorrectAction",
			newLimitPolicy("resource", "action"),
			&Limited{
				Resource:    "resource",
				Action:      "action1",
				Per:         LimitPerTotal,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			ErrInvalidLimit,
		},
		{
			"DuplicateLimit",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				if err := lp.add(&Limited{
					Resource:    "resource",
					Action:      "action",
					Per:         LimitPerTotal,
					MaxRequests: 20,
					Period:      time.Minute,
				}); err != nil {
					t.Fatal(err)
				}
				return lp
			}(),
			&Limited{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				MaxRequests: 10,
				Period:      time.Minute,
			},
			ErrDuplicateLimit,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.limitPolicy.add(tc.add)
			if tc.expectErr != nil {
				require.ErrorIs(t, got, tc.expectErr)
				return
			}
			require.NoError(t, got)
		})
	}
}

func TestLimitPolicyLimit(t *testing.T) {
	cases := []struct {
		name        string
		limitPolicy *limitPolicy
		per         LimitPer
		expect      Limit
		expectErr   error
	}{
		{
			"Total",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			LimitPerTotal,
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
			"LimitPerAuthToken",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			LimitPerAuthToken,
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
			"IPAddress",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			LimitPerIPAddress,
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
			"Missing",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			LimitPerTotal,
			nil,
			ErrLimitNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.limitPolicy.limit(tc.per)
			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestLimitPolicy_httpHeaderValue(t *testing.T) {
	cases := []struct {
		name        string
		limitPolicy *limitPolicy
		expect      string
	}{
		{
			"Policy",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="total", 10;w=60;comment="ip-address", 10;w=60;comment="auth-token"`,
		},
		{
			"UnlimitedTotal",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					var err error
					switch per {
					case LimitPerTotal:
						err = lp.add(&Unlimited{
							Resource: "resource",
							Action:   "action",
							Per:      per,
						})
					default:
						err = lp.add(&Limited{
							Resource:    "resource",
							Action:      "action",
							Per:         per,
							MaxRequests: 10,
							Period:      time.Minute,
						})
					}
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="ip-address", 10;w=60;comment="auth-token"`,
		},
		{
			"UnlimitedIpAddress",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					var err error
					switch per {
					case LimitPerIPAddress:
						err = lp.add(&Unlimited{
							Resource: "resource",
							Action:   "action",
							Per:      per,
						})
					default:
						err = lp.add(&Limited{
							Resource:    "resource",
							Action:      "action",
							Per:         per,
							MaxRequests: 10,
							Period:      time.Minute,
						})
					}
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="total", 10;w=60;comment="auth-token"`,
		},
		{
			"UnlimitedAuthToken",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					var err error
					switch per {
					case LimitPerAuthToken:
						err = lp.add(&Unlimited{
							Resource: "resource",
							Action:   "action",
							Per:      per,
						})
					default:
						err = lp.add(&Limited{
							Resource:    "resource",
							Action:      "action",
							Per:         per,
							MaxRequests: 10,
							Period:      time.Minute,
						})
					}
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="total", 10;w=60;comment="ip-address"`,
		},
		{
			"UnlimitedIpAddressAuthToken",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					var err error
					switch per {
					case LimitPerIPAddress, LimitPerAuthToken:
						err = lp.add(&Unlimited{
							Resource: "resource",
							Action:   "action",
							Per:      per,
						})
					default:
						err = lp.add(&Limited{
							Resource:    "resource",
							Action:      "action",
							Per:         per,
							MaxRequests: 10,
							Period:      time.Minute,
						})
					}
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="total"`,
		},
		{
			"UnlimitedIpAddressTotal",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					var err error
					switch per {
					case LimitPerIPAddress, LimitPerTotal:
						err = lp.add(&Unlimited{
							Resource: "resource",
							Action:   "action",
							Per:      per,
						})
					default:
						err = lp.add(&Limited{
							Resource:    "resource",
							Action:      "action",
							Per:         per,
							MaxRequests: 10,
							Period:      time.Minute,
						})
					}
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="auth-token"`,
		},
		{
			"UnlimitedAuthTokenTotal",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					var err error
					switch per {
					case LimitPerAuthToken, LimitPerTotal:
						err = lp.add(&Unlimited{
							Resource: "resource",
							Action:   "action",
							Per:      per,
						})
					default:
						err = lp.add(&Limited{
							Resource:    "resource",
							Action:      "action",
							Per:         per,
							MaxRequests: 10,
							Period:      time.Minute,
						})
					}
					require.NoError(t, err)
				}
				return lp
			}(),
			`10;w=60;comment="ip-address"`,
		},
		{
			"AllUnlimited",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Unlimited{
						Resource: "resource",
						Action:   "action",
						Per:      per,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			``,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.limitPolicy.httpHeaderValue()
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestLimitPolicy_validate(t *testing.T) {
	cases := []struct {
		name        string
		limitPolicy *limitPolicy
		expectErr   error
	}{
		{
			"NoError",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			nil,
		},
		{
			"MissingResource",
			&limitPolicy{
				resource: "",
				action:   "action",
			},
			ErrInvalidLimitPolicy,
		},
		{
			"MissingAction",
			&limitPolicy{
				resource: "resource",
				action:   "",
			},
			ErrInvalidLimitPolicy,
		},
		{
			"MissingPerTotal",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerIPAddress, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			ErrInvalidLimitPolicy,
		},
		{
			"MissingPerIPAddress",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerAuthToken} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			ErrInvalidLimitPolicy,
		},
		{
			"MissingPerAuthToken",
			func() *limitPolicy {
				lp := newLimitPolicy("resource", "action")
				for _, per := range []LimitPer{LimitPerTotal, LimitPerIPAddress} {
					err := lp.add(&Limited{
						Resource:    "resource",
						Action:      "action",
						Per:         per,
						MaxRequests: 10,
						Period:      time.Minute,
					})
					require.NoError(t, err)
				}
				return lp
			}(),
			ErrInvalidLimitPolicy,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.limitPolicy.validate()
			if tc.expectErr != nil {
				require.ErrorIs(t, got, tc.expectErr)
				return
			}
			require.NoError(t, got)
		})
	}
}

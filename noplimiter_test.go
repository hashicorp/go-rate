// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate_test

import (
	"net/http"
	"testing"

	"github.com/hashicorp/go-rate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnlimiterAllow(t *testing.T) {
	cases := []struct {
		name      string
		res       string
		action    string
		ip        string
		authtoken string
	}{
		{
			"empty",
			"",
			"",
			"",
			"",
		},
		{
			"notEmpty",
			"res",
			"action",
			"127.0.0.1",
			"auth-token",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, q, err := rate.NopLimiter.Allow(tc.res, tc.action, tc.ip, tc.authtoken)
			require.NoError(t, err)
			assert.Nil(t, q)
			assert.True(t, a)
		})
	}
}

func TestUnlimitedSetPolicyHeader(t *testing.T) {
	cases := []struct {
		name   string
		res    string
		action string
	}{
		{
			"empty",
			"",
			"",
		},
		{
			"notEmpty",
			"res",
			"action",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := make(http.Header)
			err := rate.NopLimiter.SetPolicyHeader(tc.res, tc.action, h)
			require.NoError(t, err)
			assert.Empty(t, h)
		})
	}
}

func TestUnlimitedSetUsageHeader(t *testing.T) {
	cases := []struct {
		name string
		q    *rate.Quota
	}{
		{
			"nil",
			nil,
		},
		{
			"notNil",
			&rate.Quota{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := make(http.Header)
			rate.NopLimiter.SetUsageHeader(tc.q, h)
			assert.Empty(t, h)
		})
	}
}

func TestUnlimitedShutdown(t *testing.T) {
	err := rate.NopLimiter.Shutdown()
	assert.NoError(t, err)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuota_reset(t *testing.T) {
	q := &Quota{}
	require.Nil(t, q.limit)
	require.Equal(t, uint64(0), q.used)
	require.True(t, q.Expiration().IsZero())

	l := &Limit{
		Resource:    "resource",
		Action:      "action",
		Per:         LimitPerTotal,
		Unlimited:   false,
		MaxRequests: 10,
		Period:      time.Minute,
	}
	q.reset(l)
	assert.Equal(t, l, q.limit)
	assert.Equal(t, uint64(0), q.used)
	assert.Equal(t, uint64(10), q.MaxRequests())
	q.used = 5

	l2 := &Limit{
		Resource:    "resource",
		Action:      "action",
		Per:         LimitPerTotal,
		Unlimited:   false,
		MaxRequests: 50,
		Period:      time.Minute * 10,
	}
	q.reset(l2)
	assert.Equal(t, l2, q.limit)
	assert.Equal(t, uint64(0), q.used)
	assert.Equal(t, uint64(50), q.MaxRequests())
}

func TestQuotaConsume(t *testing.T) {
	l := &Limit{
		Resource:    "resource",
		Action:      "action",
		Per:         LimitPerTotal,
		Unlimited:   false,
		MaxRequests: 10,
		Period:      time.Minute,
	}
	q := &Quota{}
	q.reset(l)
	require.Equal(t, uint64(0), q.used)

	q.Consume()
	assert.Equal(t, uint64(1), q.used)
}

func TestQuotaExpired(t *testing.T) {
	cases := []struct {
		name   string
		period time.Duration
		sleep  time.Duration
		want   bool
	}{
		{
			"notExpired",
			time.Minute,
			0,
			false,
		},
		{
			"expired",
			time.Millisecond,
			time.Millisecond * 2,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := &Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				Unlimited:   false,
				MaxRequests: 10,
				Period:      tc.period,
			}
			q := &Quota{}
			q.reset(l)
			time.Sleep(tc.sleep)
			got := q.Expired()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestQuotaRemaining(t *testing.T) {
	cases := []struct {
		name        string
		maxRequests uint64
		used        uint64
		want        uint64
	}{
		{
			"remaining",
			20,
			10,
			10,
		},
		{
			"none",
			20,
			20,
			0,
		},
		{
			"negative",
			20,
			21,
			0,
		},
		{
			"maxused",
			20,
			math.MaxUint64,
			0,
		},
		{
			"maxMaxRequests",
			math.MaxUint64,
			math.MaxUint64 - 1,
			1,
		},
		{
			"maxMaxRequestsUsed",
			math.MaxUint64,
			math.MaxUint64,
			0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := &Limit{
				Resource:    "resource",
				Action:      "action",
				Per:         LimitPerTotal,
				Unlimited:   false,
				MaxRequests: tc.maxRequests,
				Period:      time.Minute,
			}
			q := &Quota{}
			q.reset(l)
			q.used = tc.used

			got := q.Remaining()
			require.Equal(t, tc.want, got)
		})
	}
}

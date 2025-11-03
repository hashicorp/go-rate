// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newExpirableStore(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		maxSize         int
		maxEntryTTL     time.Duration
		options         []Option
		expectBuckets   int
		expectBucketTTL time.Duration
		expectErr       error
	}{
		{
			"DefaultNumberBuckets",
			10,
			time.Second * DefaultNumberBuckets,
			[]Option{},
			DefaultNumberBuckets,
			time.Second * DefaultNumberBuckets / (DefaultNumberBuckets - 1),
			nil,
		},
		{
			"WithNumberBuckets",
			10,
			time.Minute,
			[]Option{WithNumberBuckets(60)},
			60,
			time.Second * 60 / (60 - 1),
			nil,
		},
		{
			"ZeroSize",
			0,
			time.Minute,
			[]Option{},
			60,
			time.Second,
			ErrInvalidMaxSize,
		},
		{
			"NegativeSize",
			-1,
			time.Minute,
			[]Option{},
			60,
			time.Second,
			ErrInvalidMaxSize,
		},
		{
			"ZeroBuckets",
			10,
			time.Minute,
			[]Option{WithNumberBuckets(0)},
			60,
			time.Second,
			ErrInvalidNumberBuckets,
		},
		{
			"NegativeBuckets",
			10,
			time.Minute,
			[]Option{WithNumberBuckets(-1)},
			60,
			time.Second,
			ErrInvalidNumberBuckets,
		},
		{
			"ZeroMaxTTL",
			10,
			0,
			[]Option{},
			60,
			time.Second,
			ErrInvalidParameter,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := newExpirableStore(tc.maxSize, tc.maxEntryTTL, tc.options...)
			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
				return
			}
			require.NotNil(t, s)
			assert.Equal(t, tc.expectBuckets, len(s.buckets))
			assert.Equal(t, tc.expectBucketTTL, s.bucketTTL)
		})
	}
}

func Test_storeCapacity(t *testing.T) {
	maxSize := 5
	s, err := newExpirableStore(maxSize, time.Minute)
	require.NoError(t, err)

	limit := &Limited{
		Resource:    "resource",
		Action:      "action",
		Per:         LimitPerTotal,
		MaxRequests: 10,
		Period:      time.Minute,
	}

	var i int
	for ; i < maxSize; i++ {
		_, err := s.fetch(fmt.Sprintf("id-%d", i), limit)
		require.NoError(t, err)
	}

	_, err = s.fetch(fmt.Sprintf("id-%d", maxSize), limit)
	require.EqualError(t, err, (&ErrLimiterFull{}).Error())
}

func Test_storeDeleteExpired(t *testing.T) {
	maxPeriod := 5 * time.Second
	numberBuckets := 10 * int(maxPeriod.Seconds())
	s, err := newExpirableStore(20, maxPeriod, WithNumberBuckets(numberBuckets))
	require.NoError(t, err)

	short := &Limited{
		Resource:    "resource",
		Action:      "short",
		Per:         LimitPerTotal,
		MaxRequests: 10,
		Period:      maxPeriod / time.Duration(numberBuckets),
	}

	long := &Limited{
		Resource:    "resource",
		Action:      "long",
		Per:         LimitPerTotal,
		MaxRequests: 10,
		Period:      maxPeriod,
	}

	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		ids = append(ids, fmt.Sprintf("id-%d", i))
	}

	for _, id := range ids {
		_, err := s.fetch(id, short)
		require.NoError(t, err)

		_, err = s.fetch(id, long)
		require.NoError(t, err)
	}

	s.mu.Lock()
	got := len(s.items)
	s.mu.Unlock()
	require.Equal(t, 10, got)

	// sleep to let cleanup run
	time.Sleep(short.Period * 2)

	s.mu.Lock()
	got = len(s.items)
	s.mu.Unlock()
	require.Equal(t, 5, got)
}

func Test_ResetBucketSize(t *testing.T) {
	maxPeriod := time.Millisecond * 500
	numberBuckets := 1

	s, err := newExpirableStore(20, maxPeriod, WithNumberBuckets(numberBuckets))
	require.NoError(t, err)

	limit := &Limited{
		Resource:    "resource",
		Action:      "action",
		Per:         LimitPerTotal,
		MaxRequests: 10,
		Period:      maxPeriod,
	}

	quotas := bucketSizeThreshold + 1
	ids := make([]string, 0, quotas)
	for i := 0; i < quotas; i++ {
		ids = append(ids, fmt.Sprintf("id-%d", i))
	}

	for _, id := range ids {
		_, err := s.fetch(id, limit)
		require.NoError(t, err)
	}

	s.mu.Lock()
	got := len(s.items)
	gotBucketSize := len(s.buckets[0].entries)
	require.Equal(t, 1, len(s.buckets))
	initialBucketPtr := reflect.ValueOf(s.buckets[0].entries).Pointer()
	s.mu.Unlock()
	require.Equal(t, quotas, got)
	require.Equal(t, quotas, gotBucketSize)

	// sleep to let cleanup run
	time.Sleep(maxPeriod * 2)

	s.mu.Lock()
	got = len(s.items)
	gotBucketSize = len(s.buckets[0].entries)
	newBucketPtr := reflect.ValueOf(s.buckets[0].entries).Pointer()
	s.mu.Unlock()
	require.Equal(t, 0, got)
	require.Equal(t, 0, gotBucketSize)
	// Check that we have a pointer to a new map, since it should have allocated
	// a new one to reduce the capacity.
	require.NotEqual(t, initialBucketPtr, newBucketPtr)
}

func Test_storeFetchExpired(t *testing.T) {
	maxPeriod := time.Minute
	// Use a small number of buckets so that each bucket is larger
	// and we can be sure that the quotas are not deleted during the test
	numberBuckets := 5
	s, err := newExpirableStore(20, maxPeriod, WithNumberBuckets(numberBuckets))
	require.NoError(t, err)

	limit := &Limited{
		Resource:    "resource",
		Action:      "short",
		Per:         LimitPerTotal,
		MaxRequests: 10,
		Period:      time.Millisecond,
	}
	id := "id"

	q, err := s.fetch(id, limit)
	require.NoError(t, err)
	assert.Equal(t, uint64(10), q.Remaining())
	// Consume a quota so that remaining is now 9
	q.Consume()
	assert.Equal(t, uint64(9), q.Remaining())

	// Wait for the quota to expire
	time.Sleep(q.ResetsIn())

	q, err = s.fetch(id, limit)
	require.NoError(t, err)
	// Ensure quota has reset.
	assert.Equal(t, uint64(10), q.Remaining())
}

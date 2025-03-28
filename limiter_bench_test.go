// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkAllow(b *testing.B) {
	numResources := 128
	resources := make([]string, 128)
	limits := make([]Limit, 0, 3*numResources)
	action := "action"

	for i := 0; i < numResources; i++ {
		res := fmt.Sprintf("res_%d", i)
		resources[i] = res
		limits = append(
			limits,
			&Limited{
				Resource:    res,
				Action:      action,
				Per:         LimitPerTotal,
				MaxRequests: 100,
				Period:      time.Minute,
			},
			&Limited{
				Resource:    res,
				Action:      action,
				Per:         LimitPerIPAddress,
				MaxRequests: 100,
				Period:      time.Minute,
			},
			&Limited{
				Resource:    res,
				Action:      action,
				Per:         LimitPerAuthToken,
				MaxRequests: 100,
				Period:      time.Minute,
			},
		)
	}

	l, err := NewLimiter(limits, numResources*3)
	if err != nil {
		b.Fatalf("unexpected error: %q", err)
	}
	var rIdx int

	ss, ok := l.quotaFetcher.(*expirableStore)
	if !ok {
		b.Fatalf("quotaFetcher is not an expirableStore")
	}

	// pre-allocate into the sync pool
	p := make([]any, ss.maxSize)
	for i := 0; i < ss.maxSize; i++ {
		p[i] = ss.pool.Get()
	}
	for _, e := range p {
		ss.pool.Put(e)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rIdx = i % numResources
		_, _, err := l.Allow(resources[rIdx], action, "127.0.0.1", "token")
		if err != nil {
			b.Fatal(err)
		}
	}
}

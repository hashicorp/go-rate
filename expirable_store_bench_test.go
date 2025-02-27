// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"strconv"
	"testing"
	"time"
)

// package variables to prevent allocations from getting optimized away
var (
	benchStore  *expirableStore
	benchBucket bucket
)

// Benchmark_expirableStore is used to report on initial memory allocations
// of the expirableStore when given different max sizes.
func Benchmark_expirableStore(b *testing.B) {
	cases := []int{
		1,
		2048,
		32768,
		262144,
		524288,
	}
	for _, bc := range cases {
		b.Run(strconv.Itoa(bc), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				benchStore, err = newExpirableStore(bc, time.Minute)
				if err != nil {
					b.Fatal(err)
				}
				if err := benchStore.shutdown(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark_bucket shows how much memory is allocated when creating a map
// of buckets at the bucketSizeThreshold. This is mainly to detect if
// something changes with how go allocates maps.
func Benchmark_bucket(b *testing.B) {
	cases := []int{
		bucketSizeThreshold,
		bucketSizeThreshold + 1,
	}

	for _, bc := range cases {
		b.Run(strconv.Itoa(bc), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchBucket = bucket{
					entries: make(map[string]*entry, bc),
				}
			}
		})
	}
}

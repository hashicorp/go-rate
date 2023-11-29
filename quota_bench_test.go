// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"testing"
)

// package variables to prevent allocations from getting optimized away
var benchQuota *Quota

func BenchmarkQuota(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchQuota = &Quota{}
	}
}

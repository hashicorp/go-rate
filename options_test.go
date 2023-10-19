// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOpts(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		opts := getOpts()
		testOpts := options{
			withNumberBuckets: DefaultNumberBuckets,
			withPolicyHeader:  DefaultPolicyHeader,
			withUsageHeader:   DefaultUsageHeader,
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithNumberBuckets", func(t *testing.T) {
		opts := getOpts(WithNumberBuckets(40))
		testOpts := options{
			withNumberBuckets: 40,
			withPolicyHeader:  DefaultPolicyHeader,
			withUsageHeader:   DefaultUsageHeader,
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithPolicyHeader", func(t *testing.T) {
		opts := getOpts(WithPolicyHeader("Limit-Policy"))
		testOpts := options{
			withNumberBuckets: DefaultNumberBuckets,
			withPolicyHeader:  "Limit-Policy",
			withUsageHeader:   DefaultUsageHeader,
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithUsageHeader", func(t *testing.T) {
		opts := getOpts(WithUsageHeader("Quota-Usage"))
		testOpts := options{
			withNumberBuckets: DefaultNumberBuckets,
			withPolicyHeader:  DefaultPolicyHeader,
			withUsageHeader:   "Quota-Usage",
		}
		assert.Equal(t, opts, testOpts)
	})
}

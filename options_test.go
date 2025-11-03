// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testGauge struct {
	v float64
}

func (t *testGauge) Set(f float64) {
	t.v = f
}

func TestGetOpts(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		opts := getOpts()
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithNumberBuckets", func(t *testing.T) {
		opts := getOpts(WithNumberBuckets(40))
		testOpts := options{
			withNumberBuckets:              40,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithPolicyHeader", func(t *testing.T) {
		opts := getOpts(WithPolicyHeader("Limit-Policy"))
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               "Limit-Policy",
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithUsageHeader", func(t *testing.T) {
		opts := getOpts(WithUsageHeader("Quota-Usage"))
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                "Quota-Usage",
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithQuotaStorageCapacityMetric", func(t *testing.T) {
		g := &testGauge{}
		g.Set(5.0)
		opts := getOpts(WithQuotaStorageCapacityMetric(g))
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: g,
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithQuotaStorageCapacityMetricNil", func(t *testing.T) {
		opts := getOpts(WithQuotaStorageCapacityMetric(nil))
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithQuotaStorageUsageMetric", func(t *testing.T) {
		g := &testGauge{}
		g.Set(5.0)
		opts := getOpts(WithQuotaStorageUsageMetric(g))
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    g,
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithQuotaStorageUsageMetricNil", func(t *testing.T) {
		opts := getOpts(WithQuotaStorageUsageMetric(nil))
		testOpts := options{
			withNumberBuckets:              DefaultNumberBuckets,
			withPolicyHeader:               DefaultPolicyHeader,
			withUsageHeader:                DefaultUsageHeader,
			withQuotaStorageCapacityMetric: &nilGauge{},
			withQuotaStorageUsageMetric:    &nilGauge{},
		}
		assert.Equal(t, opts, testOpts)
	})
}

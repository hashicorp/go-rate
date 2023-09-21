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
		}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithNumberBuckets", func(t *testing.T) {
		opts := getOpts(WithNumberBuckets(40))
		testOpts := options{
			withNumberBuckets: 40,
		}
		assert.Equal(t, opts, testOpts)
	})
}

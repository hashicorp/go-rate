// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getKey(t *testing.T) {
	cases := []struct {
		parts []string
		want  string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{"one"},
			"one",
		},
		{
			[]string{"one", "two"},
			"one:two",
		},
		{
			[]string{"one", "two", "three"},
			"one:two:three",
		},
		{
			[]string{"one", "two", "three", "four"},
			"one:two:three:four",
		},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			got := join(tc.parts...)
			assert.Equal(t, tc.want, got)
		})
	}
}

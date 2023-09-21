// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

// DefaultNumberBuckets is the default number of buckets created for the quota store.
const DefaultNumberBuckets = 4096

// Option provides a way to pass optional arguments.
type Option func(*options)

func getOpts(opt ...Option) options {
	opts := getDefaultOptions()
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

type options struct {
	withNumberBuckets int
}

func getDefaultOptions() options {
	return options{
		withNumberBuckets: DefaultNumberBuckets,
	}
}

// WithNumberBuckets is used to set the number of buckets created for the quota store.
func WithNumberBuckets(n int) Option {
	return func(o *options) {
		o.withNumberBuckets = n
	}
}

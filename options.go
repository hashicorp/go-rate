// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

const (
	// DefaultNumberBuckets is the default number of buckets created for the quota store.
	DefaultNumberBuckets = 4096

	// DefaultPolicyHeader is the default HTTP header for reporting the rate limit policy.
	DefaultPolicyHeader = "RateLimit-Policy"
)

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
	withPolicyHeader  string
}

func getDefaultOptions() options {
	return options{
		withNumberBuckets: DefaultNumberBuckets,
		withPolicyHeader:  DefaultPolicyHeader,
	}
}

// WithNumberBuckets is used to set the number of buckets created for the quota store.
func WithNumberBuckets(n int) Option {
	return func(o *options) {
		o.withNumberBuckets = n
	}
}

// WithPolicyHeader is used to set the header key used by the Limiter for
// reporting the limit policy.
func WithPolicyHeader(h string) Option {
	return func(o *options) {
		o.withPolicyHeader = h
	}
}

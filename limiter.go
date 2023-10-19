// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type quotaFetcher interface {
	// fetch will get a Quota for the provided key.
	// If no quota is found, a new one will be created using the provided Limit.
	fetch(key string, limit *Limit) (*Quota, error)
	// shutdown stops a quotaFetcher.
	shutdown() error
}

// Limiter is used to determine if a request for a given resource and action
// should be allowed.
// TODO: expand this doc
type Limiter struct {
	policies     map[string]*limitPolicy
	policyHeader string

	mu sync.RWMutex

	quotaFetcher quotaFetcher
}

// NewLimiter will create a Limiter with the provided limits and max size. The
// limits must each be unique, where uniqueness is determined by the
// combination of "resource", "action", and "per". The maxSize must be greater
// than zero. This size is the number of individual quotas that can be stored
// in memory at any given time. Once this size is reached, requests that would
// result in a new quota being inserted will not be allowed. Requests that
// correspond to existing quotas will still be processed as normal. Space will
// become available once quotas expire and are removed.
//
// Supported options are:
//   - WithNumberBuckets: Sets the number of buckets used for expiring quotas.
//     This must be greater than zero, and defaults to DefaultNumberBuckets. A
//     larger number of buckets can increase the efficiency at which expired
//     quotas are deleted to free up space. However, it does also marginally
//     increase the amount of memory needed, and can increase the frequency
//     in which the delete routine runs and must acquire a lock.
//   - WithPolicyHeader: Sets the HTTP Header key to use when setting the policy
//     header via SetPolicyHeader. This defaults to "RateLimit-Policy".
func NewLimiter(limits []*Limit, maxSize int, o ...Option) (*Limiter, error) {
	const op = "rate.NewLimiter"

	switch {
	case len(limits) <= 0:
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyLimits)
	}

	opts := getOpts(o...)

	policies := make(map[string]*limitPolicy, len(limits)/3)

	var policy *limitPolicy
	var ok bool
	var maxEntryTTL time.Duration
	for _, l := range limits {

		if !l.IsValid() {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidLimit)
		}
		polKey := getKey(l.Resource, l.Action)

		policy, ok = policies[polKey]
		if !ok {
			policy = newLimitPolicy(l.Resource, l.Action)
			policies[polKey] = policy
		}
		if err := policy.add(l); err != nil {
			return nil, err
		}

		if l.Period > maxEntryTTL {
			maxEntryTTL = l.Period
		}
	}

	for _, p := range policies {
		if err := p.validate(); err != nil {
			return nil, err
		}
	}

	// TODO: handle special case where all of the provided limits have Unlimited = true.
	// If this is the case, we can skip the creation of a quotaFetcher

	s, err := newExpirableStore(maxSize, maxEntryTTL, o...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	l := &Limiter{
		policies:     policies,
		policyHeader: opts.withPolicyHeader,
		quotaFetcher: s,
	}

	return l, nil
}

// SetPolicyHeader sets the rate limit policy HTTP header for the provided
// resource and action.
func (l *Limiter) SetPolicyHeader(resource, action string, header http.Header) error {
	polKey := getKey(resource, action)
	pol, ok := l.policies[polKey]
	if !ok {
		return ErrLimitPolicyNotFound
	}
	p := pol.String()
	if p == "" {
		return nil
	}

	header.Set(l.policyHeader, pol.String())
	return nil
}

// Allow checks if a request for the given resource and action should be allowed.
// A request is not allowed if:
//   - Any of the associated quotas have been exhausted.
//   - A new quota needs to be stored but there is no available space to store it.
//     The error returned in this case will be a ErrLimiterFull with a provided
//     RetryIn duration. Callers should use this time as an estimation of when
//     the limiter should no longer be full.
//   - There is no corresponding limit for the resource and action.
func (l *Limiter) Allow(resource, action, ip, authToken string) (allowed bool, quota *Quota, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allowOrder := []LimitPer{
		LimitPerTotal,
		LimitPerIPAddress,
		LimitPerAuthToken,
	}

	quotas := make(map[LimitPer]*Quota, len(allowOrder))
	keys := map[LimitPer]string{
		LimitPerTotal:     string(LimitPerTotal),
		LimitPerIPAddress: ip,
		LimitPerAuthToken: authToken,
	}

	var ok bool
	var limit *Limit
	var policy *limitPolicy
	var q *Quota
	var key string
	allowed = true
	for per, id := range keys {
		key = getKey(resource, action)
		policy, ok = l.policies[key]
		if !ok {
			allowed = false
			err = ErrLimitPolicyNotFound
			return
		}

		limit, err = policy.limit(per)
		if err != nil {
			allowed = false
			return
		}

		q, err = l.quotaFetcher.fetch(id, limit)
		if err != nil {
			allowed = false
			return
		}

		if q.Remaining() <= 0 {
			allowed = false
			quota = q
			return
		}

		quotas[per] = q
	}

	for _, per := range allowOrder {
		q, _ := quotas[per]
		q.Consume()
		if quota == nil || q.Remaining() < quota.Remaining() {
			quota = q
		}
	}

	return
}

// Shutdown stops a Limiter. After calling this, any future calls to Allow
// will result in ErrStopped being returned.
func (l *Limiter) Shutdown() error {
	return l.quotaFetcher.shutdown()
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"fmt"
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
	limits map[string]*Limit

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
func NewLimiter(limits []*Limit, maxSize int, o ...Option) (*Limiter, error) {
	const op = "rate.NewLimiter"

	switch {
	case len(limits) <= 0:
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyLimits)
	}

	byKey := make(map[string]*Limit, len(limits))

	var maxEntryTTL time.Duration
	for _, l := range limits {
		if !l.IsValid() {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidLimit)
		}
		key := getKey(l.Resource, l.Action, string(l.Per))
		if _, ok := byKey[key]; ok {
			return nil, fmt.Errorf("%s: %s %s %s: %w", op, l.Resource, l.Action, l.Per, ErrDuplicateLimit)
		}
		byKey[key] = l
		if l.Period > maxEntryTTL {
			maxEntryTTL = l.Period
		}
	}

	// TODO: handle special case where all of the provided limits have Unlimited = true.
	// If this is the case, we can skip the creation of a quotaFetcher

	s, err := newExpirableStore(maxSize, maxEntryTTL, o...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	l := &Limiter{
		limits:       byKey,
		quotaFetcher: s,
	}

	return l, nil
}

// Allow checks if a request for the given resource and action should be allowed.
// A request is not allowed if:
//   - Any of the associated quotas have been exhausted.
//   - A new quota needs to be stored but there is no available space to store it.
//     The error returned in this case will be a ErrLimiterFull with a provided
//     RetryIn duration. Callers should use this time as an estimation of when
//     the limiter should no longer be full.
//   - There is no corresponding limit for the resource and action.
func (l *Limiter) Allow(resource, action string) (allowed bool, quota *Quota, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allowOrder := []LimitPer{LimitPerTotal}

	quotas := make(map[LimitPer]*Quota, len(allowOrder))
	keys := map[LimitPer]string{
		LimitPerTotal: string(LimitPerTotal),
	}

	var ok bool
	var limit *Limit
	var q *Quota
	var key string
	allowed = true
	for per, id := range keys {
		key = getKey(resource, action, string(per))
		limit, ok = l.limits[key]
		if !ok {
			allowed = false
			err = ErrLimitNotFound
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

		quotas[LimitPerTotal] = q
	}

	for _, q := range quotas {
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

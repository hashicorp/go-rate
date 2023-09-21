// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type entry struct {
	key   string
	value *Quota

	bucket int
}

type bucket struct {
	entries map[string]*entry

	expiresAt time.Time
}

// TODO document this, in particular provide some details around:
// - the purpose and use of "buckets"
type expirableStore struct {
	maxSize int

	items map[string]*entry

	buckets            []bucket
	bucketTTL          time.Duration
	numberBuckets      int
	nextBucketToExpire int

	mu sync.Mutex

	pool sync.Pool

	cancelFunc context.CancelFunc
	ctx        context.Context
}

func newExpirableStore(maxSize int, maxEntryTTL time.Duration, o ...Option) (*expirableStore, error) {
	const op = "rate.newExpirableStore"

	opts := getOpts(o...)

	switch {
	case maxSize <= 0:
		return nil, fmt.Errorf("%s: max size must be greater than zero: %w", op, ErrInvalidMaxSize)
	case maxEntryTTL <= 0:
		return nil, fmt.Errorf("%s: max entry ttl must be greater than zero: %w", op, ErrInvalidParameter)
	case opts.withNumberBuckets <= 0:
		return nil, fmt.Errorf("%s: number of buckets must be greater than zero: %w", op, ErrInvalidNumberBuckets)
	}

	bucketTTL := (maxEntryTTL) / time.Duration(opts.withNumberBuckets-1)

	buckets := make([]bucket, opts.withNumberBuckets)
	for i := 0; i < opts.withNumberBuckets; i++ {
		buckets[i] = bucket{
			entries: make(map[string]*entry, maxSize),
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &expirableStore{
		maxSize:       maxSize,
		items:         make(map[string]*entry, maxSize),
		buckets:       buckets,
		bucketTTL:     bucketTTL,
		numberBuckets: opts.withNumberBuckets,
		pool: sync.Pool{
			New: func() any {
				return &entry{
					value: &Quota{},
				}
			},
		},
		cancelFunc: cancel,
		ctx:        ctx,
	}

	go s.deleteExpired()
	return s, nil
}

func (s *expirableStore) shutdown() error {
	s.cancelFunc()
	return nil
}

func (s *expirableStore) deleteExpired() {
	ticker := time.NewTicker(s.bucketTTL)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.emptyExpiredBucket()
		}
	}
}

// TODO: document this
func (s *expirableStore) fetch(id string, limit *Limit) (*Quota, error) {
	select {
	case <-s.ctx.Done():
		return nil, ErrStopped
	default:
		// continue
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := getKey(limit.Resource, limit.Action, string(limit.Per), id)

	e, ok := s.items[key]
	switch {
	case !ok:
		e = s.pool.Get().(*entry)
		e.key = key
		e.value.reset(limit)
		if err := s.add(e); err != nil {
			s.pool.Put(e)
			return nil, err
		}
	case e.value.Expired():
		s.removeFromBucket(e)
		e.value.reset(limit)
		s.addToBucket(e)
	}

	return e.value, nil
}

// add attempts to add an entry to the store. If the store has reached its
// max capacity, ErrLimiterFull is returned.
//
// add should always be called by a function that first acquires a lock
func (s *expirableStore) add(e *entry) error {
	const op = "rate.(expirableStore).add"
	if s.mu.TryLock() {
		panic(fmt.Sprintf("%s: called without lock", op))
	}
	if _, ok := s.items[e.key]; !ok && len(s.items) >= s.maxSize {
		return ErrLimiterFull
	}
	s.items[e.key] = e
	s.addToBucket(e)
	return nil
}

// addToBucket adds the entry to a bucket based on the entry's expiration time.
//
// addToBucket should always be called by a function that first acquires a lock
func (s *expirableStore) addToBucket(e *entry) {
	const op = "rate.(expirableStore).addToBucket"
	if s.mu.TryLock() {
		panic(fmt.Sprintf("%s: called without lock", op))
	}
	e.bucket = (int(e.value.limit.Period/s.bucketTTL) + s.nextBucketToExpire) % s.numberBuckets
	s.buckets[e.bucket].entries[e.key] = e
	if s.buckets[e.bucket].expiresAt.Before(e.value.expiresAt) {
		s.buckets[e.bucket].expiresAt = e.value.expiresAt
	}
}

// emptyExpiredBuckets is called via a go routine. It should run approximately
// once every s.bucketTTL to delete all of the items in the next expired bucket.
func (s *expirableStore) emptyExpiredBucket() {
	s.mu.Lock()

	toExpire := s.nextBucketToExpire
	s.nextBucketToExpire = (s.nextBucketToExpire + 1) % s.numberBuckets

	timeToExpire := time.Until(s.buckets[toExpire].expiresAt)
	// Just in case, check to see if this has run early and there is still some
	// time before the bucket expires. in which case wait until the bucket has
	// expired before deleting.
	if timeToExpire > 0 {
		s.mu.Unlock()
		time.Sleep(timeToExpire)
		s.mu.Lock()
	}
	defer s.mu.Unlock()
	for _, delEnt := range s.buckets[toExpire].entries {
		s.removeEntry(delEnt)
	}
}

// removeEntry removes the entry from the store and adds the entry back to
// the sync pool.
//
// removeEntry should always be called by a function that first acquires a lock
func (s *expirableStore) removeEntry(e *entry) {
	const op = "rate.(expirableStore).removeEntry"
	if s.mu.TryLock() {
		panic(fmt.Sprintf("%s: called without lock", op))
	}
	delete(s.items, e.key)
	s.removeFromBucket(e)
	s.pool.Put(e)
}

// removeFromBucket removes the entry from the corresponding bucket.
//
// removeFromBucket should always be called by a function that first acquires a lock
func (s *expirableStore) removeFromBucket(e *entry) {
	const op = "rate.(expirableStore).removeFromBucket"
	if s.mu.TryLock() {
		panic(fmt.Sprintf("%s: called without lock", op))
	}
	delete(s.buckets[e.bucket].entries, e.key)
}

// ensure expirableStore can be used as a quotaFetcher
var _ quotaFetcher = (*expirableStore)(nil)

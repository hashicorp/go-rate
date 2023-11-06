// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"errors"
	"time"
)

// ErrLimiterFull is returned by Limiter.Allow when the limiter cannot store
// any additional quotas.
type ErrLimiterFull struct {
	RetryIn time.Duration
}

func (l *ErrLimiterFull) Error() string {
	return "limiter full"
}

var (
	// ErrLimitNotFound is returned by Limiter.Allow when a limit could not be
	// found for a given resource+action.
	ErrLimitNotFound = errors.New("limit not found")
	// ErrInvalidParameter represents an invalid parameter error.
	ErrInvalidParameter = errors.New("invalid parameter")
	// ErrEmptyLimits is returned by NewLimiter when no limits are provided.
	ErrEmptyLimits = errors.New("limits must not be empty")
	// ErrInvalidLimit is returned by NewLimiter when a limit is not valid.
	ErrInvalidLimit = errors.New("invalid limit")
	// ErrDuplicateLimit is returned by NewLimiter when it is provided duplicate
	// limits.
	ErrDuplicateLimit = errors.New("duplicate limit")
	// ErrInvalidNumberBuckets is returned by NewLimiter when an invalid number
	// of buckets is provided.
	ErrInvalidNumberBuckets = errors.New("invalid number of buckets")
	// ErrInvalidMaxSize is returned by NewLimiter when provided an invalid max size.
	ErrInvalidMaxSize = errors.New("invalid max size")
	// ErrStopped is returned by Limiter.Allow if the limiter has been stopped
	// and cannot return a quota.
	ErrStopped = errors.New("limiter stopped")
)

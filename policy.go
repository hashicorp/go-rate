// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"fmt"
	"strings"
)

// limitPolicy is a collection of Limits for the same resource and action. A limitPolicy
// should contain one Limit for each valid LimitPer.
type limitPolicy struct {
	resource string
	action   string

	m map[LimitPer]Limit

	policy string
}

var requiredLimitPer = []LimitPer{LimitPerTotal, LimitPerIPAddress, LimitPerAuthToken}

func newLimitPolicy(resource, action string) *limitPolicy {
	return &limitPolicy{
		resource: resource,
		action:   action,
		m:        make(map[LimitPer]Limit, 3),
	}
}

// httpHeaderValue returns a string representation of the LimitPolicy. This is
// formatted for use as a rate limit policy HTTP header as outlined in:
// https://datatracker.ietf.org/doc/draft-ietf-httpapi-ratelimit-headers/
func (p *limitPolicy) httpHeaderValue() string {
	return p.policy
}

// limit returns the corresponding limit for the given LimitPer. If the policy
// does not have a corresponding limit, ErrLimitNotFound is returned.
func (p *limitPolicy) limit(per LimitPer) (Limit, error) {
	l, ok := p.m[per]
	if !ok {
		return nil, ErrLimitNotFound
	}
	return l, nil
}

func (p *limitPolicy) add(l Limit) error {
	if err := l.validate(); err != nil {
		return err
	}

	switch {
	case l.GetResource() != p.resource:
		return fmt.Errorf("limit's resource does not match limit policy's: %w", ErrInvalidLimit)
	case l.GetAction() != p.action:
		return fmt.Errorf("limit's action does not match limit policy's: %w", ErrInvalidLimit)
	}

	if _, ok := p.m[l.GetPer()]; ok {
		return ErrDuplicateLimit
	}

	p.m[l.GetPer()] = l
	p.buildStr()
	return nil
}

func (p *limitPolicy) buildStr() {
	s := make([]string, 0, 3)
	for _, per := range requiredLimitPer {
		l, ok := p.m[per]
		if !ok {
			continue
		}
		switch ll := l.(type) {
		case *Limited:
			s = append(s, fmt.Sprintf("%d;w=%d;comment=%q", ll.MaxRequests, uint64(ll.Period.Seconds()), ll.Per.String()))
		}

	}

	p.policy = strings.Join(s, ", ")
}

func (p *limitPolicy) validate() error {
	switch {
	case p.resource == "":
		return fmt.Errorf("missing resource: %w", ErrInvalidLimitPolicy)
	case p.action == "":
		return fmt.Errorf("missing action: %w", ErrInvalidLimitPolicy)
	case len(p.m) != 3:
		for _, per := range requiredLimitPer {
			if _, ok := p.m[per]; !ok {
				return fmt.Errorf("mising limit for %q: %w", per, ErrInvalidLimitPolicy)
			}
		}
	}
	return nil
}

// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import (
	"regexp"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/syncutil"
)

// RegexpCacheKey allows cache keys to take the form of different types,
// as long as they are comparable and can produce a pattern when needed
// for regexp compilation. The pattern method will not be called until
// after a cache lookup is performed and the result is a miss.
type RegexpCacheKey interface {
	Pattern() (string, error)
}

// A RegexpCache is a cache used to store compiled regular expressions.
// The cache is safe for concurrent use by multiple goroutines. It is also
// safe to use the cache through a nil reference, where it will act like a valid
// cache with no capacity.
type RegexpCache struct {
	mu    syncutil.Mutex
}



// GetRegexp consults the cache for the regular expressions stored for
// the given key, compiling the key's pattern if it is not already
// in the cache.
func (rc *RegexpCache) GetRegexp(key RegexpCacheKey) (*regexp.Regexp, error) {
	if rc != nil {
		re := rc.lookup(key)
		if re != nil {
			return re, nil
		}
	}

	pattern, err := key.Pattern()
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	if rc != nil {
		rc.update(key, re)
	}
	return re, nil
}

// lookup checks for the regular expression in the cache in a
// synchronized manner, returning it if it exists.
func (rc *RegexpCache) lookup(key RegexpCacheKey) *regexp.Regexp {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return nil
}

// update invalidates the regular expression for the given pattern.
// If a new regular expression is passed in, it is inserted into the cache.
func (rc *RegexpCache) update(key RegexpCacheKey, re *regexp.Regexp) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
}

// Len returns the number of compiled regular expressions in the cache.
func (rc *RegexpCache) Len() int {
	if rc == nil {
		return 0
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return 0
}

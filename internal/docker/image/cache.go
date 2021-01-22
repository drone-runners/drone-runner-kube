// Copyright 2020 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
)

// cachedImage stores the cached image.
type cachedImage struct {
	image *Image
	until time.Time
}

// LookupCache provides an in-memory cache for retrieving
// image metadata from a remote registry.
type LookupCache struct {
	get LookupFunc
	lru *lru.Cache
	ttl time.Duration
}

// NewLookupCache returns a new LookupCache.
func NewLookupCache(fn LookupFunc, size int, ttl time.Duration) *LookupCache {
	lru, _ := lru.New(size)
	return &LookupCache{
		get: fn,
		ttl: ttl,
		lru: lru,
	}
}

// Lookup returns the image metadata from an in-memory
// cache. If the metadata is not available in the cache, it
// will lookup the metadata in the remote docker registry,
// authenticating if needed.
func (c *LookupCache) Lookup(image, username, password string) (*Image, error) {
	// check to see if the image exists in the lru
	// cache and is not expired.
	cached, ok := c.lru.Get(image)
	if ok {
		wrapper := cached.(*cachedImage)
		// if the image is expired, it is revoked from
		// the cache so that it can be refreshed, otherwise
		// it is returned.
		if time.Now().After(wrapper.until) {
			c.lru.Remove(image)
		} else {
			return wrapper.image, nil
		}
	}

	// if the image does not exist in the remote cache,
	// looku the remote image from the registry.
	found, err := c.get(image, username, password)
	if err != nil {
		return nil, err
	}

	// add the image to the cache with the default TTL
	// set to evict the item from the cache.
	c.lru.Add(image, &cachedImage{
		image: found,
		until: time.Now().Add(c.ttl),
	})

	return found, nil
}

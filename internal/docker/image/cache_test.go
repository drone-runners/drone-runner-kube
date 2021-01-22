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
	"testing"
	"time"
)

func TestCache_Exists(t *testing.T) {
	want := &Image{}
	fn := func(image, username, password string) (*Image, error) {
		t.Errorf("Expect image returned from cache")
		return nil, nil
	}
	cache := NewLookupCache(fn, 25, time.Minute)
	cache.lru.Add("alpine", &cachedImage{want, time.Now().Add(time.Minute)})
	got, err := cache.Lookup("alpine", "", "")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Expected image fetched from remote repository")
	}
}

func TestCache_Expired(t *testing.T) {
	want := &Image{}
	fn := func(image, username, password string) (*Image, error) {
		return want, nil
	}
	cache := NewLookupCache(fn, 25, time.Minute)
	cache.lru.Add("alpine", &cachedImage{nil, time.Unix(0, 0)})
	got, err := cache.Lookup("alpine", "", "")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Expected image fetched from remote repository")
	}
}

func TestCache_NotExists(t *testing.T) {
	want := &Image{}
	fn := func(image, username, password string) (*Image, error) {
		return want, nil
	}
	cache := NewLookupCache(fn, 25, time.Minute)
	got, err := cache.Lookup("alpine", "", "")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Expected image fetched from remote repository")
	}
}

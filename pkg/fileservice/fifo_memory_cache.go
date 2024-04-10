// Copyright 2024 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fileservice

import (
	"context"

	"github.com/matrixorigin/matrixone/pkg/fileservice/fifocache"
	"github.com/matrixorigin/matrixone/pkg/perfcounter"
	v2 "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
)

type FIFOMemoryCache struct {
	cache       *fifocache.Cache[CacheKey, Bytes]
	counterSets []*perfcounter.CounterSet
}

func NewFIFOMemoryCache(
	capacity int,
	counterSets []*perfcounter.CounterSet,
) *FIFOMemoryCache {
	ret := &FIFOMemoryCache{
		cache: fifocache.New[CacheKey, Bytes](capacity, nil, func(key CacheKey) uint8 {
			return uint8(key.Offset ^ key.Sz)
		}),
		counterSets: counterSets,
	}
	return ret
}

var _ IOVectorCache = new(FIFOMemoryCache)

func (m *FIFOMemoryCache) Read(
	ctx context.Context,
	vector *IOVector,
) (
	err error,
) {

	if vector.Policy.Any(SkipMemoryCacheReads) {
		return nil
	}

	var numHit, numRead int64
	defer func() {
		v2.FSReadHitMemCounter.Add(float64(numHit))
		perfcounter.Update(ctx, func(c *perfcounter.CounterSet) {
			c.FileService.Cache.Read.Add(numRead)
			c.FileService.Cache.Hit.Add(numHit)
			c.FileService.Cache.Memory.Read.Add(numRead)
			c.FileService.Cache.Memory.Hit.Add(numHit)
		}, m.counterSets...)
	}()

	path, err := ParsePath(vector.FilePath)
	if err != nil {
		return err
	}

	for i, entry := range vector.Entries {
		if entry.done {
			continue
		}
		key := CacheKey{
			Path:   path.File,
			Offset: entry.Offset,
			Sz:     entry.Size,
		}
		bs, ok := m.cache.Get(key)
		numRead++
		if ok {
			vector.Entries[i].CachedData = bs
			vector.Entries[i].done = true
			vector.Entries[i].fromCache = m
			numHit++
		}
	}

	return
}

func (m *FIFOMemoryCache) Update(
	ctx context.Context,
	vector *IOVector,
	async bool,
) error {

	if vector.Policy.Any(SkipMemoryCacheWrites) {
		return nil
	}

	path, err := ParsePath(vector.FilePath)
	if err != nil {
		return err
	}

	for _, entry := range vector.Entries {
		if entry.CachedData == nil {
			continue
		}
		if entry.fromCache == m {
			continue
		}

		key := CacheKey{
			Path:   path.File,
			Offset: entry.Offset,
			Sz:     entry.Size,
		}

		data := Bytes(entry.CachedData.Bytes())
		m.cache.Set(key, data, len(data))
	}
	return nil
}

func (m *FIFOMemoryCache) Flush() {
}

func (m *FIFOMemoryCache) DeletePaths(
	ctx context.Context,
	paths []string,
) error {
	return nil
}

// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fileservice

import (
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"go.uber.org/zap"
	"sync/atomic"
)

type CachingFsStatsCollector struct {
	cachingFs *CachingFileService
}

func NewCachingFsStatsCollector(cachingFs *CachingFileService) metric.StatsCollector {
	return &CachingFsStatsCollector{
		cachingFs: cachingFs,
	}
}

// Collect returns the fields and its values.
// TODO: Replace []zap.Field with `map[string]int` or `Struct`
func (c *CachingFsStatsCollector) Collect() []zap.Field {
	var fields []zap.Field

	counter := (*c.cachingFs).CacheCounter()

	reads := atomic.LoadInt64(&counter.CacheRead)
	hits := atomic.LoadInt64(&counter.CacheHit)
	memReads := atomic.LoadInt64(&counter.MemCacheRead)
	memHits := atomic.LoadInt64(&counter.MemCacheHit)
	diskReads := atomic.LoadInt64(&counter.DiskCacheRead)
	diskHits := atomic.LoadInt64(&counter.DiskCacheHit)

	fields = append(fields, zap.Any("reads", reads))
	fields = append(fields, zap.Any("hits", hits))
	fields = append(fields, zap.Any("hit rate", float64(hits)/float64(reads)))
	fields = append(fields, zap.Any("mem reads", memReads))
	fields = append(fields, zap.Any("mem hits", memHits))
	fields = append(fields, zap.Any("mem hit rate", float64(memHits)/float64(memReads)))

	fields = append(fields, zap.Any("disk reads", diskReads))
	fields = append(fields, zap.Any("disk hits", diskHits))

	fields = append(fields, zap.Any("disk hit rate", float64(diskHits)/float64(diskReads)))

	return fields
}

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
	"sync"

	"github.com/matrixorigin/matrixone/pkg/common/malloc"
	metric "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
)

const (
	// 1/freezeFraction of allocations may be frozen to detect wrong mutation such as mutating the memory cache objects
	freezeFraction = 100
)

func decorateAllocator(allocator malloc.Allocator) malloc.Allocator {
	// freeze randomly to detect wrong mutation
	// this makes the allocator randomly freezable, whether to freeze is decided by the callers of Allocate method
	allocator = malloc.NewRandomAllocator(
		allocator,
		malloc.NewReadOnlyAllocator(allocator),
		freezeFraction,
	)
	return allocator
}

var memoryCacheAllocator = sync.OnceValue(func() malloc.Allocator {
	allocator := malloc.GetDefault(nil)
	// with metrics
	allocator = malloc.NewMetricsAllocator(
		allocator,
		metric.MallocCounter.WithLabelValues("memory-cache-allocate"),
		metric.MallocGauge.WithLabelValues("memory-cache-inuse"),
		metric.MallocCounter.WithLabelValues("memory-cache-allocate-objects"),
		metric.MallocGauge.WithLabelValues("memory-cache-inuse-objects"),
	)
	// peak in-use
	allocator = malloc.NewInuseTrackingAllocator(
		allocator,
		func(inuse uint64) {
			malloc.GlobalPeakInuseTracker.Update("/malloc/memory-cache", inuse)
		},
	)
	// decorate
	allocator = decorateAllocator(allocator)
	return allocator
})

var ioAllocator = sync.OnceValue(func() malloc.Allocator {
	allocator := malloc.GetDefault(nil)
	// with metrics
	allocator = malloc.NewMetricsAllocator(
		allocator,
		metric.MallocCounter.WithLabelValues("io-allocate"),
		metric.MallocGauge.WithLabelValues("io-inuse"),
		metric.MallocCounter.WithLabelValues("io-allocate-objects"),
		metric.MallocGauge.WithLabelValues("io-inuse-objects"),
	)
	// peak in-use
	allocator = malloc.NewInuseTrackingAllocator(
		allocator,
		func(inuse uint64) {
			malloc.GlobalPeakInuseTracker.Update("/malloc/io", inuse)
		},
	)
	// decorate
	allocator = decorateAllocator(allocator)
	return allocator
})

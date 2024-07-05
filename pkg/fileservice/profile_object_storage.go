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
	"io"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/google/pprof/profile"
	"github.com/matrixorigin/matrixone/pkg/common/malloc"
)

type ObjectStorageSampleValues struct {
	All    malloc.ShardedCounter[int64, atomic.Int64, *atomic.Int64]
	Flying malloc.ShardedCounter[int64, atomic.Int64, *atomic.Int64]
}

var _ malloc.SampleValues = new(ObjectStorageSampleValues)

func (o *ObjectStorageSampleValues) Init() {
	o.All = *malloc.NewShardedCounter[int64, atomic.Int64](runtime.GOMAXPROCS(0))
	o.Flying = *malloc.NewShardedCounter[int64, atomic.Int64](runtime.GOMAXPROCS(0))
}

func (o *ObjectStorageSampleValues) SampleTypes() []*profile.ValueType {
	return []*profile.ValueType{
		{
			Type: "all",
			Unit: "times",
		},
		{
			Type: "flying",
			Unit: "times",
		},
	}
}

func (o *ObjectStorageSampleValues) DefaultSampleType() string {
	return "all"
}

func (o *ObjectStorageSampleValues) Values() []int64 {
	return []int64{
		o.All.Load(),
		o.Flying.Load(),
	}
}

type ProfileObjectStorage struct {
	upstream ObjectStorage
	profiler *malloc.Profiler[ObjectStorageSampleValues, *ObjectStorageSampleValues]
}

func NewProfileObjectStorage(
	upstream ObjectStorage,
	profiler *malloc.Profiler[ObjectStorageSampleValues, *ObjectStorageSampleValues],
) *ProfileObjectStorage {
	return &ProfileObjectStorage{
		upstream: upstream,
		profiler: profiler,
	}
}

func (p *ProfileObjectStorage) sample() *ObjectStorageSampleValues {
	values := p.profiler.Sample(1, 1)
	values.All.Add(1)
	values.Flying.Add(1)
	return values
}

var _ ObjectStorage = new(ProfileObjectStorage)

func (p *ProfileObjectStorage) Delete(ctx context.Context, keys ...string) (err error) {
	values := p.sample()
	defer values.Flying.Add(-1)
	return p.upstream.Delete(ctx, keys...)
}

func (p *ProfileObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	values := p.sample()
	defer values.Flying.Add(-1)
	return p.upstream.Exists(ctx, key)
}

func (p *ProfileObjectStorage) List(ctx context.Context, prefix string, fn func(isPrefix bool, key string, size int64) (bool, error)) (err error) {
	values := p.sample()
	defer values.Flying.Add(-1)
	return p.upstream.List(ctx, prefix, fn)
}

func (p *ProfileObjectStorage) Read(ctx context.Context, key string, min *int64, max *int64) (r io.ReadCloser, err error) {
	values := p.sample()
	defer values.Flying.Add(-1)
	return p.upstream.Read(ctx, key, min, max)
}

func (p *ProfileObjectStorage) Stat(ctx context.Context, key string) (size int64, err error) {
	values := p.sample()
	defer values.Flying.Add(-1)
	return p.upstream.Stat(ctx, key)
}

func (p *ProfileObjectStorage) Write(ctx context.Context, key string, r io.Reader, size int64, expire *time.Time) (err error) {
	values := p.sample()
	defer values.Flying.Add(-1)
	return p.upstream.Write(ctx, key, r, size, expire)
}

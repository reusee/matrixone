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

package missingfreeguard

import (
	"net/http"
	"runtime"
	"sync/atomic"

	"github.com/google/pprof/profile"
	"github.com/matrixorigin/matrixone/pkg/common/malloc"
)

type Manager struct {
	profiler *malloc.Profiler[sampleValues, *sampleValues]
}

func NewManager(
	profileHandlePath string,
) *Manager {

	profiler := malloc.NewProfiler[sampleValues]()
	http.HandleFunc(profileHandlePath, func(w http.ResponseWriter, req *http.Request) {
		profiler.Write(w)
	})

	return &Manager{
		profiler: profiler,
	}
}

type Guard struct {
	free   bool
	values *sampleValues
	bytes  int64
	_      NoCopy
}

func (g *Manager) NewGuard(target any, bytes int64) *Guard {
	values := g.profiler.Sample(1, 1)
	values.Bytes.Add(bytes)
	ret := &Guard{
		values: values,
	}
	if target != nil {
		runtime.SetFinalizer(target, func(any) {
			if !ret.free {
				ret.values.Missing.Store(true)
			}
		})
	} else {
		runtime.SetFinalizer(ret, func(guard *Guard) {
			if !guard.free {
				guard.values.Missing.Store(true)
			}
		})
	}
	return ret
}

func (f *Guard) Free() {
	f.free = true
	f.values.Bytes.Add(-f.bytes)
}

type sampleValues struct {
	Missing atomic.Bool
	Bytes   atomic.Int64
}

var _ malloc.SampleValues = new(sampleValues)

func (m *sampleValues) DefaultSampleType() string {
	return "missing"
}

func (m *sampleValues) Init() {
}

func (m *sampleValues) SampleTypes() []*profile.ValueType {
	return []*profile.ValueType{
		{
			Type: "missing",
			Unit: "trace",
		},
		{
			Type: "bytes",
			Unit: "byte",
		},
	}
}

func (m *sampleValues) Values() []int64 {
	if m.Missing.Load() {
		return []int64{1, m.Bytes.Load()}
	}
	return []int64{0, 0}
}

type NoCopy struct{}

func (n *NoCopy) Lock()   {}
func (n *NoCopy) Unlock() {}

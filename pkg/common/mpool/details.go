// Copyright 2021 - 2024 Matrix Origin
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

package mpool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/matrixorigin/matrixone/pkg/util/stack"
)

type detailInfo struct {
	cnt, bytes int64
}

type mpoolDetails struct {
	mu    sync.Mutex
	alloc map[string]detailInfo
	free  map[string]detailInfo
}

func newMpoolDetails() *mpoolDetails {
	mpd := mpoolDetails{}
	mpd.alloc = make(map[string]detailInfo)
	mpd.free = make(map[string]detailInfo)
	return &mpd
}

func (d *mpoolDetails) recordAlloc(nb int64) {
	f := stack.Caller(2)
	k := fmt.Sprintf("%v", f)
	d.mu.Lock()
	defer d.mu.Unlock()

	info := d.alloc[k]
	info.cnt += 1
	info.bytes += nb
	d.alloc[k] = info
}

func (d *mpoolDetails) recordFree(nb int64) {
	f := stack.Caller(2)
	k := fmt.Sprintf("%v", f)
	d.mu.Lock()
	defer d.mu.Unlock()

	info := d.free[k]
	info.cnt += 1
	info.bytes += nb
	d.free[k] = info
}

func (d *mpoolDetails) reportJson() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	ret := `{"alloc": {`
	allocs := make([]string, 0)
	for k, v := range d.alloc {
		kvs := fmt.Sprintf("\"%s\": [%d, %d]", k, v.cnt, v.bytes)
		allocs = append(allocs, kvs)
	}
	ret += strings.Join(allocs, ",")
	ret += `}, "free": {`
	frees := make([]string, 0)
	for k, v := range d.free {
		kvs := fmt.Sprintf("\"%s\": [%d, %d]", k, v.cnt, v.bytes)
		frees = append(frees, kvs)
	}
	ret += strings.Join(frees, ",")
	ret += "}}"
	return ret
}

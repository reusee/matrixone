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
	"sync/atomic"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/logutil"
)

// Stats
type MPoolStats struct {
	NumAlloc      atomic.Int64 // number of allocations
	NumFree       atomic.Int64 // number of frees
	NumAllocBytes atomic.Int64 // number of bytes allocated
	NumFreeBytes  atomic.Int64 // number of bytes freed
	NumCurrBytes  atomic.Int64 // current number of bytes
	HighWaterMark atomic.Int64 // high water mark
}

func (s *MPoolStats) Report(tab string) string {
	if s.HighWaterMark.Load() == 0 {
		// empty, reduce noise.
		return ""
	}

	ret := ""
	ret += fmt.Sprintf("%s allocations : %d\n", tab, s.NumAlloc.Load())
	ret += fmt.Sprintf("%s frees : %d\n", tab, s.NumFree.Load())
	ret += fmt.Sprintf("%s alloc bytes : %d\n", tab, s.NumAllocBytes.Load())
	ret += fmt.Sprintf("%s free bytes : %d\n", tab, s.NumFreeBytes.Load())
	ret += fmt.Sprintf("%s current bytes : %d\n", tab, s.NumCurrBytes.Load())
	ret += fmt.Sprintf("%s high water mark : %d\n", tab, s.HighWaterMark.Load())
	return ret
}

func (s *MPoolStats) ReportJson() string {
	if s.HighWaterMark.Load() == 0 {
		return ""
	}
	ret := "{"
	ret += fmt.Sprintf("\"alloc\": %d,", s.NumAlloc.Load())
	ret += fmt.Sprintf("\"free\": %d,", s.NumFree.Load())
	ret += fmt.Sprintf("\"allocBytes\": %d,", s.NumAllocBytes.Load())
	ret += fmt.Sprintf("\"freeBytes\": %d,", s.NumFreeBytes.Load())
	ret += fmt.Sprintf("\"currBytes\": %d,", s.NumCurrBytes.Load())
	ret += fmt.Sprintf("\"highWaterMark\": %d", s.HighWaterMark.Load())
	ret += "}"
	return ret
}

// Update alloc stats, return curr bytes
func (s *MPoolStats) RecordAlloc(tag string, sz int64) int64 {
	s.NumAlloc.Add(1)
	s.NumAllocBytes.Add(sz)
	curr := s.NumCurrBytes.Add(sz)
	hwm := s.HighWaterMark.Load()
	if curr > hwm {
		swapped := s.HighWaterMark.CompareAndSwap(hwm, curr)
		if swapped && curr/GB != hwm/GB {
			logutil.Infof("MPool %s new high watermark\n%s", tag, s.Report("    "))
		}
	}
	return curr
}

// Update free stats, return curr bytes.
func (s *MPoolStats) RecordFree(tag string, sz int64) int64 {
	if sz < 0 {
		logutil.Errorf("Mpool %s free bug, stats: %s", tag, s.Report("    "))
		panic(moerr.NewInternalErrorNoCtx("mpool freed -1"))
	}
	s.NumFree.Add(1)
	s.NumFreeBytes.Add(sz)
	curr := s.NumCurrBytes.Add(-sz)
	if curr < 0 {
		logutil.Errorf("Mpool %s free bug, stats: %s", tag, s.Report("    "))
		panic(moerr.NewInternalErrorNoCtx("mpool freed more bytes than alloc"))
	}
	return curr
}

func (s *MPoolStats) RecordManyFrees(tag string, nfree, sz int64) int64 {
	if sz < 0 {
		logutil.Errorf("Mpool %s free bug, stats: %s", tag, s.Report("    "))
		panic(moerr.NewInternalErrorNoCtx("mpool freed -1"))
	}
	s.NumFree.Add(nfree)
	s.NumFreeBytes.Add(sz)
	curr := s.NumCurrBytes.Add(-sz)
	if curr < 0 {
		logutil.Errorf("Mpool %s free many bug, stats: %s", tag, s.Report("    "))
		panic(moerr.NewInternalErrorNoCtx("mpool freemany freed more bytes than alloc"))
	}
	return curr
}

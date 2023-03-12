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

package mometric

import (
	"context"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"sync"
	"sync/atomic"
	"time"
)

type StatsLogExporter struct {
	isRunning int32
	cancel    context.CancelFunc
	stopWg    sync.WaitGroup

	registry       *metric.StatsRegistry
	gatherInterval time.Duration
}

func newStatsLogExporter(registry *metric.StatsRegistry, gatherInterval time.Duration) *StatsLogExporter {
	return &StatsLogExporter{
		registry:       registry,
		gatherInterval: gatherInterval,
	}
}

func (e *StatsLogExporter) Start(inputCtx context.Context) bool {
	if atomic.SwapInt32(&e.isRunning, 1) == 1 {
		return false
	}
	ctx, cancel := context.WithCancel(inputCtx)
	e.cancel = cancel
	e.stopWg.Add(1)
	go func() {
		defer e.stopWg.Done()
		ticker := time.NewTicker(e.gatherInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.gatherAndExport()
			case <-ctx.Done():
				return
			}
		}
	}()
	return true
}

func (e *StatsLogExporter) Stop(_ bool) (<-chan struct{}, bool) {
	if atomic.SwapInt32(&e.isRunning, 0) == 0 {
		return nil, false
	}
	e.cancel()
	stopCh := make(chan struct{})
	go func() { e.stopWg.Wait(); close(stopCh) }()
	return stopCh, true
}

func (e *StatsLogExporter) gatherAndExport() {
	statsCollection := e.registry.Gather()
	for statsFName, stats := range statsCollection {
		logutil.Info("cache stats of "+statsFName, stats...)
	}
}

// Copyright 2021 Matrix Origin
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

package main

import (
	"runtime/metrics"
	"time"

	"github.com/matrixorigin/matrixone/pkg/logutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (m *Manager) Metrics() (
	p Parser,
	usage Usage,
) {

	var logIntervalSeconds uint64
	p = p.MatchStr("-log-metrics-interval")(
		p.Uint64(&logIntervalSeconds)(
			p.End(func() {
				if logIntervalSeconds == 0 {
					return
				}
				m.on(evInit, func() {
					go startMetricsLogging(logIntervalSeconds)
				})
			})))

	usage.Header = `-log-metrics-interval seconds`
	usage.Desc = `log metrics every specified seconds. 0 means disable logging`

	return
}

func startMetricsLogging(intervalSeconds uint64) {
	samples := []metrics.Sample{
		// gc infos
		{
			Name: "/gc/heap/allocs:bytes",
		},
		{
			Name: "/gc/heap/frees:bytes",
		},
		{
			Name: "/gc/heap/goal:bytes",
		},
		// memory infos
		{
			Name: "/memory/classes/heap/free:bytes",
		},
		{
			Name: "/memory/classes/heap/objects:bytes",
		},
		{
			Name: "/memory/classes/heap/released:bytes",
		},
		{
			Name: "/memory/classes/heap/unused:bytes",
		},
		{
			Name: "/memory/classes/total:bytes",
		},
		// goroutine infos
		{
			Name: "/sched/goroutines:goroutines",
		},
	}

	for range time.NewTicker(time.Second * time.Duration(intervalSeconds)).C {
		metrics.Read(samples)

		var fields []zapcore.Field
		for _, sample := range samples {
			switch sample.Value.Kind() {
			case metrics.KindUint64:
				fields = append(fields, zap.Uint64(sample.Name, sample.Value.Uint64()))
			case metrics.KindFloat64:
				fields = append(fields, zap.Float64(sample.Name, sample.Value.Float64()))
			}
		}

		logutil.Debug("runtime metrics", fields...)

	}
}

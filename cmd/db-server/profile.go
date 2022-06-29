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
	"os"
	"runtime/metrics"
	"runtime/pprof"
	"time"

	"github.com/matrixorigin/matrixone/pkg/logutil"
)

func (_ Def) Profiles(
	on On,
) (
	parsers ArgumentParsers,
	usages Usages,
) {

	var p Parser

	// cpu
	var cpuProfilePath string
	parsers = append(parsers, p.MatchStr("-cpu-profile")(
		p.String(&cpuProfilePath)(
			p.End(func() {
				if cpuProfilePath == "" {
					return
				}
				on(evInit, func() {
					stop := startCPUProfile(cpuProfilePath)
					on(evExit, func() {
						stop()
					})
				})
			}))))
	usages = append(usages, [2]string{`-cpu-profile`, `write cpu profile to the specified file`})

	// allocs
	var allocsProfilePath string
	parsers = append(parsers, p.MatchStr("-allocs-profile")(
		p.String(&allocsProfilePath)(
			p.End(func() {
				if allocsProfilePath != "" {
					on(evExit, func() {
						writeAllocsProfile(allocsProfilePath)
					})
				}
			}))))
	usages = append(usages, [2]string{`-allocs-profile`, `write allocs profile to the specified file`})

	// heap
	var heapProfilePath string
	heapProfileThreshold := uint64(8 * 1024 * 1024 * 1024)
	parsers = append(parsers, p.MatchStr("-heap-profile")(
		p.String(&heapProfilePath)(
			p.End(func() {
				if heapProfilePath == "" {
					return
				}
				on(evInit, func() {
					go startHeapProfiler(heapProfilePath, heapProfileThreshold)
				})
			}))))
	usages = append(usages, [2]string{`-heap-profile`, `write heap profile to the specified file`})
	parsers = append(parsers, p.MatchStr("-heap-profile-threshold")(
		p.Uint64(&heapProfileThreshold)(nil)))
	usages = append(usages, [2]string{`-heap-profile-threshold`, `take a heap profile if mapped memory changes exceed the specified threshold bytes`})

	return
}

func startCPUProfile(filePath string) func() {
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}
	logutil.Infof("CPU profiling enabled, writing to %s", filePath)
	return func() {
		pprof.StopCPUProfile()
		f.Close()
	}
}

func writeAllocsProfile(filePath string) {
	profile := pprof.Lookup("allocs")
	if profile == nil {
		return
	}
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := profile.WriteTo(f, 0); err != nil {
		panic(err)
	}
	logutil.Infof("Allocs profile written to %s", filePath)
}

func startHeapProfiler(filePath string, threshold uint64) {
	samples := []metrics.Sample{
		{
			Name: "/memory/classes/total:bytes",
		},
	}
	var lastUsing uint64
	for range time.NewTicker(time.Second).C {

		writeHeapProfile := func() {
			metrics.Read(samples)
			using := samples[0].Value.Uint64()
			if using-lastUsing > threshold {
				profile := pprof.Lookup("heap")
				if profile == nil {
					return
				}
				profilePath := filePath
				profilePath += "." + time.Now().Format("15:04:05.000000")
				f, err := os.Create(profilePath)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				if err := profile.WriteTo(f, 0); err != nil {
					panic(err)
				}
				logutil.Infof("Heap profile written to %s", profilePath)
			}
			lastUsing = using
		}

		writeHeapProfile()
	}

}

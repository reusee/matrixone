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

import "sync"

type OnCleanup func(fn func())

type Cleanup func()

func (_ Def) Cleanup() (
	onCleanup OnCleanup,
	cleanup Cleanup,
) {

	var l sync.Mutex
	var fns []func()

	onCleanup = func(fn func()) {
		l.Lock()
		defer l.Unlock()
		fns = append(fns, fn)
	}

	var once sync.Once
	cleanup = func() {
		once.Do(func() {
			l.Lock()
			defer l.Unlock()
			for _, fn := range fns {
				fn()
			}
		})
	}

	return
}

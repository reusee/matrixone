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

const (
	evInit = "init"
	evExit = "exit"
)

func (m *Manager) on(ev string, fn func()) {
	m.hooks.Lock()
	defer m.hooks.Unlock()
	m.hooks.Map[ev] = append(m.hooks.Map[ev], fn)
}

func (m *Manager) emit(ev string) {
	m.hooks.Lock()
	evs := m.hooks.Map[ev]
	m.hooks.Unlock()
	for _, fn := range evs {
		fn()
	}
}

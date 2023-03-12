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

package metric

import "go.uber.org/zap"

type StatsCollector interface {
	Collect() []zap.Field
}

// StatsRegistry holds the Collectable objects.
type StatsRegistry struct {
	registry map[string]*StatsCollector
}

// DefaultStatsRegistry will be used for registering default developer stats
var DefaultStatsRegistry = StatsRegistry{}

// Register registers stats family to registry
// statsFName is the family name of the stats
// stats is the pointer to the stats object
func (r *StatsRegistry) Register(statsFName string, stats *StatsCollector) {
	if _, exists := r.registry[statsFName]; exists {
		panic("Duplicate Stats Family Name")
	}
	r.registry[statsFName] = stats
}

// Gather returns the snapshot of all the statsFamily in the registry
func (r *StatsRegistry) Gather() (statsFamilies map[string][]zap.Field) {
	for statsFName, statsCollector := range r.registry {
		statsFamilies[statsFName] = (*statsCollector).Collect()
	}
	return
}

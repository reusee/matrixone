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

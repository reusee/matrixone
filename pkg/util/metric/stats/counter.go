package stats

import "sync/atomic"

type Counter struct {
	fName         string
	name          string
	currCounter   atomic.Int64
	globalCounter atomic.Int64
}

func NewStatsCounter(fName, name string) *Counter {
	return &Counter{
		fName: fName,
		name:  name,
	}
}

func (c *Counter) Incr() {
	c.currCounter.Add(1)
}

func (c *Counter) Load() int64 {
	return c.currCounter.Load()
}

func (c *Counter) LoadG() int64 {
	return c.globalCounter.Load()
}

func (c *Counter) MergeAndReset() {
	//TODO: Are you sure, we don't need lock here?
	c.globalCounter.Add(c.currCounter.Load())
	c.currCounter.Store(0)
}

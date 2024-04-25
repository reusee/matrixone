// Copyright 2022 Matrix Origin
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

package logtailreplay

import (
	"bytes"
	"context"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/db/checkpoint"
	"sync"
	"sync/atomic"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
)

// a partition corresponds to a dn
type Partition struct {
	//lock is used to protect pointer of PartitionState from concurrent mutation
	lock  chan struct{}
	state *PartitionState

	// assuming checkpoints will be consumed once
	checkpointConsumed atomic.Bool

	//current partitionState can serve snapshot read only if start <= ts <= end
	mu struct {
		sync.Mutex
		start types.TS
		end   types.TS
	}

	// update
	update struct {
		mutex    sync.Mutex
		cond     *sync.Cond
		updating bool
	}
}

func (p *Partition) CanServe(ts types.TS) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ts.GreaterEq(&p.mu.start) && ts.LessEq(&p.mu.end)
}

func NewPartition() *Partition {
	lock := make(chan struct{}, 1)
	lock <- struct{}{}
	ret := &Partition{
		lock: lock,
	}
	ret.mu.start = types.MaxTs()
	ret.state = NewPartitionState(false)
	ret.update.cond = sync.NewCond(&ret.update.mutex)
	return ret
}

type RowID types.Rowid

func (r RowID) Less(than RowID) bool {
	return bytes.Compare(r[:], than[:]) < 0
}

func (p *Partition) Snapshot() *PartitionState {
	p.update.mutex.Lock()
	defer p.update.mutex.Unlock()
	for p.update.updating {
		p.update.cond.Wait()
	}
	return p.state.Copy()
}

func (*Partition) CheckPoint(ctx context.Context, ts timestamp.Timestamp) error {
	panic("unimplemented")
}

func (p *Partition) MutateState() (*PartitionState, func()) {
	p.update.mutex.Lock()
	p.update.updating = true
	p.update.mutex.Unlock()
	return p.state, func() {
		p.update.mutex.Lock()
		p.update.updating = false
		p.update.mutex.Unlock()
		p.update.cond.Broadcast()
	}
}

func (p *Partition) Lock(ctx context.Context) error {
	select {
	case <-p.lock:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Partition) Unlock() {
	p.lock <- struct{}{}
}

func (p *Partition) checkValid() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mu.start.LessEq(&p.mu.end)
}

func (p *Partition) UpdateStart(ts types.TS) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mu.start != types.MaxTs() {
		p.mu.start = ts
	}
}

// [start, end]
func (p *Partition) UpdateDuration(start types.TS, end types.TS) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mu.start = start
	p.mu.end = end
}

func (p *Partition) ConsumeSnapCkps(
	_ context.Context,
	ckps []*checkpoint.CheckpointEntry,
	fn func(
		ckp *checkpoint.CheckpointEntry,
		state *PartitionState,
	) error,
) (
	err error,
) {

	p.update.mutex.Lock()
	p.update.updating = true
	p.update.mutex.Unlock()
	defer func() {
		p.update.mutex.Lock()
		p.update.updating = false
		p.update.mutex.Unlock()
	}()

	//Notice that checkpoints must contain only one or zero global checkpoint
	//followed by zero or multi continuous incremental checkpoints.
	start := types.MaxTs()
	end := types.TS{}
	for _, ckp := range ckps {
		if err = fn(ckp, p.state); err != nil {
			return
		}
		if ckp.GetType() == checkpoint.ET_Global {
			start = ckp.GetEnd()
		}
		if ckp.GetType() == checkpoint.ET_Incremental {
			ckpstart := ckp.GetStart()
			if ckpstart.Less(&start) {
				start = ckpstart
			}
			ckpend := ckp.GetEnd()
			if ckpend.Greater(&end) {
				end = ckpend
			}
		}
	}
	if end.IsEmpty() {
		//only one global checkpoint.
		end = start
	}
	p.UpdateDuration(start, end)
	if !p.checkValid() {
		panic("invalid checkpoint")
	}

	return nil
}

func (p *Partition) ConsumeCheckpoints(
	ctx context.Context,
	fn func(
		checkpoint string,
		state *PartitionState,
	) error,
) (
	err error,
) {

	p.update.mutex.Lock()
	p.update.updating = true
	p.update.mutex.Unlock()
	defer func() {
		p.update.mutex.Lock()
		p.update.updating = false
		p.update.mutex.Unlock()
	}()

	if p.checkpointConsumed.Load() {
		return nil
	}
	if len(p.state.checkpoints) == 0 {
		p.UpdateDuration(types.TS{}, types.MaxTs())
		return nil
	}

	lockErr := p.Lock(ctx)
	if lockErr != nil {
		return lockErr
	}
	defer p.Unlock()

	if len(p.state.checkpoints) == 0 {
		logutil.Infof("xxxx impossible path")
		p.UpdateDuration(types.TS{}, types.MaxTs())
		return nil
	}

	if err := p.state.consumeCheckpoints(fn); err != nil {
		return err
	}

	p.UpdateDuration(p.state.start, types.MaxTs())

	p.checkpointConsumed.Store(true)

	return
}

func (p *Partition) Truncate(ctx context.Context, ids [2]uint64, ts types.TS) error {
	err := p.Lock(ctx)
	if err != nil {
		return err
	}
	defer p.Unlock()

	p.update.mutex.Lock()
	p.update.updating = true
	p.update.mutex.Unlock()
	defer func() {
		p.update.mutex.Lock()
		p.update.updating = false
		p.update.mutex.Unlock()
	}()

	p.state.truncate(ids, ts)

	//TODO::update partition's start and end

	return nil
}

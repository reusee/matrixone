// Copyright 2023 Matrix Origin
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

package lockservice

import (
	"context"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/common/reuse"
	pb "github.com/matrixorigin/matrixone/pkg/pb/lock"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPut(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue()
		defer q.close(notifyValue{})
		w := acquireWaiter(pb.WaitTxn{TxnID: []byte("w")}, "", nil)
		defer w.close("", nil)
		q.put(w)
		assert.Equal(t, 1, q.size())
	})
}

func TestReset(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue()
		defer q.close(notifyValue{})

		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		defer w1.close("", nil)
		w2 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w2")}, "", nil)
		defer w2.close("", nil)
		w3 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w3")}, "", nil)
		defer w3.close("", nil)

		q.put(w1, w2, w3)
		assert.Equal(t, 3, q.size())

		q.iter(func(w *waiter) bool {
			w.close("", nil)
			return true
		})

		q.reset()
		assert.Equal(t, 0, q.size())
	})
}

func TestIterTxns(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue()
		defer q.close(notifyValue{})

		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		defer w1.close("", nil)
		w2 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w2")}, "", nil)
		defer w2.close("", nil)
		w3 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w3")}, "", nil)
		defer w3.close("", nil)

		q.put(w1, w2, w3)

		var values [][]byte
		v := 0
		q.iter(func(w *waiter) bool {
			values = append(values, w.txn.TxnID)
			v++
			return v < 2
		})
		assert.Equal(t, [][]byte{[]byte("w1"), []byte("w2")}, values)
	})
}

func TestIterTxnsCannotReadUncommitted(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue()
		defer q.close(notifyValue{})

		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		defer w1.close("", nil)

		q.put(w1)

		w2 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w2")}, "", nil)
		defer w2.close("", nil)
		q.beginChange()
		q.put(w2)

		var values [][]byte
		q.iter(func(w *waiter) bool {
			values = append(values, w.txn.TxnID)
			return true
		})
		assert.Equal(t, [][]byte{[]byte("w1")}, values)
		q.rollbackChange()
	})
}

func TestChange(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue().(*sliceBasedWaiterQueue)
		defer q.close(notifyValue{})

		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		defer w1.close("", nil)

		q.put(w1)

		q.beginChange()
		assert.Equal(t, 1, q.beginChangeIdx)

		w2 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w2")}, "", nil)
		defer w2.close("", nil)
		w3 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w3")}, "", nil)
		defer w3.close("", nil)

		q.put(w2, w3)
		assert.Equal(t, 3, len(q.waiters))

		q.rollbackChange()
		assert.Equal(t, 1, len(q.waiters))
		assert.Equal(t, -1, q.beginChangeIdx)

		q.beginChange()
		assert.Equal(t, 1, q.beginChangeIdx)

		w4 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w4")}, "", nil)
		defer w4.close("", nil)
		w5 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w5")}, "", nil)
		defer w5.close("", nil)

		q.put(w4, w5)
		assert.Equal(t, 3, len(q.waiters))

		q.commitChange()
		assert.Equal(t, 3, len(q.waiters))
		assert.Equal(t, -1, q.beginChangeIdx)
	})
}

func TestChangeRef(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue().(*sliceBasedWaiterQueue)
		defer q.close(notifyValue{})

		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		defer w1.close("", nil)

		q.beginChange()
		q.put(w1)
		assert.Equal(t, int32(2), w1.refCount.Load())
		q.rollbackChange()
		assert.Equal(t, int32(1), w1.refCount.Load())

		q.beginChange()
		q.put(w1)
		assert.Equal(t, int32(2), w1.refCount.Load())
		q.commitChange()
		assert.Equal(t, int32(2), w1.refCount.Load())
	})
}

func TestSkipCompletedWaiters(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue()

		// w1 will skipped
		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		w1.setStatus(completed)
		defer w1.close("", nil)

		// w2 get the notify
		w2 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w2")}, "", nil)
		w2.setStatus(blocking)
		defer func() {
			w2.wait(context.Background(), getLogger(""))
			w2.close("", nil)
		}()

		// w3 get notify when queue closed
		w3 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w3")}, "", nil)
		w3.setStatus(blocking)
		defer func() {
			w3.wait(context.Background(), getLogger(""))
			w3.close("", nil)
		}()

		q.put(w1, w2, w3)

		q.notify(notifyValue{})
		ws := make([]*waiter, 0, q.size())
		q.iter(func(w *waiter) bool {
			ws = append(ws, w)
			return true
		})
		assert.Equal(t, 2, len(ws))
		assert.Equal(t, []byte("w2"), ws[0].txn.TxnID)
		assert.Equal(t, []byte("w3"), ws[1].txn.TxnID)

		q.close(notifyValue{})
	})
}

func TestCanGetCommitTSInWaitQueue(t *testing.T) {
	reuse.RunReuseTests(func() {
		q := newWaiterQueue()
		defer q.close(notifyValue{})

		w2 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w2")}, "", nil)
		w2.setStatus(blocking)
		defer w2.close("", nil)

		w3 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w3")}, "", nil)
		w3.setStatus(blocking)
		defer w3.close("", nil)

		w4 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w4")}, "", nil)
		w4.setStatus(blocking)
		defer func() {
			w4.close("", nil)
		}()

		w5 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w5")}, "", nil)
		w5.setStatus(blocking)
		defer w5.close("", nil)

		q.put(w2, w3, w4, w5)

		// commit at 1
		q.notify(notifyValue{ts: timestamp.Timestamp{PhysicalTime: 1}})

		// w2 get notify and abort
		assert.Equal(t, int64(1), w2.wait(context.Background(), getLogger("")).ts.PhysicalTime)
		q.removeByTxnID(w2.txn.TxnID)
		q.notify(notifyValue{})

		// w3 get notify and commit at 3
		assert.Equal(t, int64(1), w3.wait(context.Background(), getLogger("")).ts.PhysicalTime)
		q.removeByTxnID(w3.txn.TxnID)
		q.notify(notifyValue{ts: timestamp.Timestamp{PhysicalTime: 3}})

		// w4 get notify and commit at 2
		assert.Equal(t, int64(3), w4.wait(context.Background(), getLogger("")).ts.PhysicalTime)
		q.removeByTxnID(w4.txn.TxnID)
		q.notify(notifyValue{ts: timestamp.Timestamp{PhysicalTime: 2}})

		// w5 get notify
		assert.Equal(t, int64(3), w5.wait(context.Background(), getLogger("")).ts.PhysicalTime)
		q.removeByTxnID(w5.txn.TxnID)
	})
}

func TestMoveToCannotCloseWaiter(t *testing.T) {
	reuse.RunReuseTests(func() {

		from := newWaiterQueue()
		defer from.close(notifyValue{})

		to := newWaiterQueue()

		w1 := acquireWaiter(pb.WaitTxn{TxnID: []byte("w1")}, "", nil)
		defer w1.close("", nil)

		from.put(w1)
		require.Equal(t, int32(2), w1.refCount.Load())

		to.beginChange()
		from.moveTo(to)
		require.Equal(t, int32(3), w1.refCount.Load())
		to.rollbackChange()

		require.Equal(t, int32(2), w1.refCount.Load())
	})

}

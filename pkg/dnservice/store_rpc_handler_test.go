// Copyright 2021 - 2022 Matrix Origin
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

package dnservice

import (
	"context"
	"testing"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/service"
	"github.com/stretchr/testify/assert"
)

func TestHandleRead(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestReadRequest(1, service.NewTestTxn(1, 1, 1), 1)
		req.CNRequest.Target.ReplicaID = 2
		assert.NoError(t, s.handleRead(context.Background(), &req, &txn.TxnResponse{}))
	})
}

func TestHandleReadWithRetry(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		req := service.NewTestReadRequest(1, service.NewTestTxn(1, 1, 1), 1)
		req.CNRequest.Target.ReplicaID = 2
		req.Options = &txn.TxnRequestOptions{
			RetryCodes: []int32{int32(moerr.ErrTNShardNotFound)},
		}
		go func() {
			time.Sleep(time.Second)
			shard := newTestTNShard(1, 2, 3)
			assert.NoError(t, s.StartDNReplica(shard))
		}()
		assert.NoError(t, s.handleRead(context.Background(), &req, &txn.TxnResponse{}))
	})
}

func TestHandleReadWithRetryWithTimeout(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		req := service.NewTestReadRequest(1, service.NewTestTxn(1, 1, 1), 1)
		req.CNRequest.Target.ReplicaID = 2
		req.Options = &txn.TxnRequestOptions{
			RetryCodes: []int32{int32(moerr.ErrTNShardNotFound)},
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		resp := &txn.TxnResponse{}
		assert.NoError(t, s.handleRead(ctx, &req, resp))
		assert.Equal(t, uint32(moerr.ErrTNShardNotFound), resp.TxnError.Code)
	})
}

func TestHandleWrite(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestWriteRequest(1, service.NewTestTxn(1, 1, 1), 1)
		req.CNRequest.Target.ReplicaID = 2
		assert.NoError(t, s.handleWrite(context.Background(), &req, &txn.TxnResponse{}))
	})
}

func TestHandleCommit(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestCommitRequest(service.NewTestTxn(1, 1, 1))
		req.Txn.TNShards[0].ReplicaID = 2
		assert.NoError(t, s.handleCommit(context.Background(), &req, &txn.TxnResponse{}))
	})
}

func TestHandleRollback(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestRollbackRequest(service.NewTestTxn(1, 1, 1))
		req.Txn.TNShards[0].ReplicaID = 2
		assert.NoError(t, s.handleRollback(context.Background(), &req, &txn.TxnResponse{}))
	})
}

func TestHandlePrepare(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestPrepareRequest(service.NewTestTxn(1, 1, 1), 1)
		req.PrepareRequest.TNShard.ReplicaID = 2
		assert.NoError(t, s.handlePrepare(context.Background(), &req, &txn.TxnResponse{}))

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_, err := s.sender.Send(ctx, []txn.TxnRequest{req})
		assert.NoError(t, err)
	})
}

func TestHandleGetStatus(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestGetStatusRequest(service.NewTestTxn(1, 1, 1), 1)
		req.GetStatusRequest.TNShard.ReplicaID = 2
		assert.NoError(t, s.handleGetStatus(context.Background(), &req, &txn.TxnResponse{}))

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_, err := s.sender.Send(ctx, []txn.TxnRequest{req})
		assert.NoError(t, err)
	})
}

func TestHandleCommitTNShard(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestCommitShardRequest(service.NewTestTxn(1, 1, 1))
		req.CommitTNShardRequest.TNShard.ReplicaID = 2
		assert.NoError(t, s.handleCommitTNShard(context.Background(), &req, &txn.TxnResponse{}))

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_, err := s.sender.Send(ctx, []txn.TxnRequest{req})
		assert.NoError(t, err)
	})
}

func TestHandleRollbackTNShard(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestRollbackShardRequest(service.NewTestTxn(1, 1, 1))
		req.RollbackTNShardRequest.TNShard.ReplicaID = 2
		assert.NoError(t, s.handleRollbackTNShard(context.Background(), &req, &txn.TxnResponse{}))

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_, err := s.sender.Send(ctx, []txn.TxnRequest{req})
		assert.NoError(t, err)
	})
}

func TestHandleDNShardNotFound(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		req := service.NewTestRollbackShardRequest(service.NewTestTxn(1, 1, 1))
		resp := &txn.TxnResponse{}
		assert.NoError(t, s.handleRollbackTNShard(context.Background(), &req, resp))
		assert.Equal(t, uint32(moerr.ErrTNShardNotFound), resp.TxnError.Code)
	})
}

func TestHandleDebug(t *testing.T) {
	runTNStoreTest(t, func(s *store) {
		shard := newTestTNShard(1, 2, 3)
		assert.NoError(t, s.StartDNReplica(shard))

		req := service.NewTestReadRequest(1, service.NewTestTxn(1, 1, 1), 1)
		req.Method = txn.TxnMethod_DEBUG
		req.CNRequest.Target.ReplicaID = 2
		assert.NoError(t, s.handleDebug(context.Background(), &req, &txn.TxnResponse{}))
	})
}

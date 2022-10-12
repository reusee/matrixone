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

package engine

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/config"
	"github.com/matrixorigin/matrixone/pkg/frontend"
	logservicepb "github.com/matrixorigin/matrixone/pkg/pb/logservice"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/matrixorigin/matrixone/pkg/txn/clock"
	"github.com/matrixorigin/matrixone/pkg/txn/storage/memorystorage"
	"github.com/matrixorigin/matrixone/pkg/txn/storage/memorystorage/memtable"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/memoryengine"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)
	defer cancel()

	frontendParameters := &config.FrontendParameters{
		MoVersion:    "1",
		RootName:     "root",
		RootPassword: "111",
		DumpUser:     "dump",
		DumpPassword: "111",
	}
	frontendParameters.SetDefaultValues()

	mp := mpool.MustNewZero()

	clock := clock.NewHLCClock(func() int64 {
		return time.Now().Unix()
	}, math.MaxInt)

	var dnStores []logservicepb.DNStore
	storages := make(map[string]*memorystorage.Storage)
	numShards := 1
	for i := 0; i < numShards; i++ {

		shard := logservicepb.DNShardInfo{
			ShardID:   uint64(i + 8),
			ReplicaID: uint64(i + 8),
		}
		shards := []logservicepb.DNShardInfo{
			shard,
		}
		dnAddr := fmt.Sprintf("1.1.1.%d", i+8)
		dnStore := logservicepb.DNStore{
			UUID:           uuid.NewString(),
			ServiceAddress: dnAddr,
			Shards:         shards,
		}

		dnStores = append(dnStores, dnStore)

		storage, err := memorystorage.NewMemoryStorage(
			mp,
			memtable.SnapshotIsolation,
			clock,
			memoryengine.RandomIDGenerator,
		)
		assert.Nil(t, err)

		storages[dnAddr] = storage
	}

	engine := memoryengine.New(
		ctx,
		memoryengine.NewDefaultShardPolicy(mp),
		func() (logservicepb.ClusterDetails, error) {
			return logservicepb.ClusterDetails{
				DNStores: dnStores,
			}, nil
		},
		memoryengine.RandomIDGenerator,
	)

	txnClient := memorystorage.NewStorageTxnClient(
		clock,
		storages,
	)

	pu := &config.ParameterUnit{
		SV:            frontendParameters,
		StorageEngine: engine,
		TxnClient:     txnClient,
		FileService:   testutil.NewFS(),
	}
	ctx = context.WithValue(ctx, config.ParameterUnitKey, pu)

	//ctx, rsStubs := mockRecordStatement(ctx)
	//defer rsStubs.Reset()

	err := frontend.InitSysTenant(ctx)
	assert.Nil(t, err)

	globalVars := new(frontend.GlobalSystemVariables)
	frontend.InitGlobalSystemVariables(globalVars)

	session := frontend.NewSession(
		frontend.NewMysqlClientProtocol(
			0,
			nil, // goetty IOSession
			1024,
			frontendParameters,
		),
		nil,
		pu,
		globalVars,
	)
	session.SetRequestContext(ctx)

	_, err = session.AuthenticateUser("root")
	assert.Nil(t, err)

	compilerCtxOp, err := txnClient.New()
	assert.Nil(t, err)
	compilerCtx := engine.NewCompilerContext(ctx, "test", compilerCtxOp)

	driver := New(engine, txnClient, compilerCtx)
	connector, err := driver.OpenConnector("")
	assert.Nil(t, err)

	db := sql.OpenDB(connector)
	defer func() {
		assert.Nil(t, db.Close())
	}()

	assert.Nil(t, db.Ping())

	dbx := sqlx.NewDb(db, "")

	var dbs []string
	err = dbx.Select(&dbs, `show databases`)
	assert.Nil(t, err)

}

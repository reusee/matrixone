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
	"bytes"
	"context"
	"encoding/gob"
	"sync"

	"github.com/matrixorigin/matrixone/pkg/logservice"
	logservicepb "github.com/matrixorigin/matrixone/pkg/pb/logservice"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

// Engine is an engine.Engine impl
type Engine struct {

	// hakeeper
	hakeeperClient logservice.CNHAKeeperClient
	clusterDetails struct {
		sync.Mutex
		logservicepb.ClusterDetails
	}

	fatal func()
}

func New(
	ctx context.Context,
	hakeeperClient logservice.CNHAKeeperClient,
) *Engine {

	ctx, cancel := context.WithCancel(ctx)
	fatal := func() {
		cancel()
	}

	engine := &Engine{
		hakeeperClient: hakeeperClient,
		fatal:          fatal,
	}
	go engine.startHAKeeperLoop(ctx)

	return engine
}

var _ engine.Engine = new(Engine)

func (e *Engine) Create(ctx context.Context, dbName string, txnOperator client.TxnOperator) error {

	// for ddl operations, broadcast to all DNs
	var requests []txn.TxnRequest
	for _, node := range e.getDataNodes() {
		requests = append(requests, txn.TxnRequest{
			Method: txn.TxnMethod_Write,
			CNRequest: &txn.CNOpRequest{
				OpCode: opCreateDatabase,
				Payload: mustEncodePayload(createDatabasePayload{
					Name: dbName,
				}),
				Target: metadata.DNShard{
					Address: node.ServiceAddress,
				},
			},
		})
	}

	result, err := txnOperator.WriteAndCommit(ctx, requests)
	if err != nil {
		return err
	}
	if err := errorFromTxnResponses(result.Responses); err != nil {
		return err
	}

	return nil
}

func (e *Engine) Database(ctx context.Context, dbName string, txnOperator client.TxnOperator) (engine.Database, error) {

	result, err := txnOperator.Read(ctx, []txn.TxnRequest{
		{
			Method: txn.TxnMethod_Read,
			CNRequest: &txn.CNOpRequest{
				OpCode: opOpenDatabase,
				Payload: mustEncodePayload(openDatabasePayload{
					Name: dbName,
				}),
				Target: metadata.DNShard{
					// use first DN
					Address: e.getDataNodes()[0].ServiceAddress,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if err := errorFromTxnResponses(result.Responses); err != nil {
		return nil, err
	}

	var payload openDatabasePayload
	if err := gob.NewDecoder(bytes.NewReader(result.Responses[0].CNOpResponse.Payload)).Decode(&payload); err != nil {
		return nil, err
	}

	db := &Database{
		engine:      e,
		txnOperator: txnOperator,
		id:          payload.ID,
	}

	return db, nil
}

func (e *Engine) Databases(ctx context.Context, txnOperator client.TxnOperator) ([]string, error) {

	result, err := txnOperator.Read(ctx, []txn.TxnRequest{
		{
			Method: txn.TxnMethod_Read,
			CNRequest: &txn.CNOpRequest{
				OpCode: opGetDatabases,
				Target: metadata.DNShard{
					// use first DN
					Address: e.getDataNodes()[0].ServiceAddress,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if err := errorFromTxnResponses(result.Responses); err != nil {
		return nil, err
	}

	var dbNames []string
	for _, resp := range result.Responses {
		var payload getDatabasesPayload
		if err := gob.NewDecoder(bytes.NewReader(resp.CNOpResponse.Payload)).Decode(&payload); err != nil {
			return nil, err
		}
		dbNames = append(dbNames, payload.Names...)
	}

	return dbNames, nil
}

func (e *Engine) Delete(ctx context.Context, dbName string, txnOperator client.TxnOperator) error {

	// for ddl operations, broadcast to all DNs
	var requests []txn.TxnRequest
	for _, node := range e.getDataNodes() {
		requests = append(requests, txn.TxnRequest{
			Method: txn.TxnMethod_Write,
			CNRequest: &txn.CNOpRequest{
				OpCode: opDeleteDatabase,
				Payload: mustEncodePayload(deleteDatabasePayload{
					Name: dbName,
				}),
				Target: metadata.DNShard{
					Address: node.ServiceAddress,
				},
			},
		})
	}

	result, err := txnOperator.WriteAndCommit(ctx, requests)
	if err != nil {
		return err
	}
	if err := errorFromTxnResponses(result.Responses); err != nil {
		return err
	}

	return nil
}

func (e *Engine) Nodes(ctx context.Context, txnOperator client.TxnOperator) engine.Nodes {

	var nodes engine.Nodes
	for _, node := range e.getDataNodes() {
		nodes = append(nodes, engine.Node{
			Mcpu: 1,
			Id:   node.UUID,
			Addr: node.ServiceAddress,
		})
	}

	return nodes
}

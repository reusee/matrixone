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

	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type Database struct {
	engine      *Engine
	txnOperator client.TxnOperator

	id int64
}

var _ engine.Database = new(Database)

func (d *Database) Create(ctx context.Context, relName string, defs []engine.TableDef) error {

	// for ddl operations, broadcast to all DNs
	var requests []txn.TxnRequest
	for _, node := range d.engine.getDataNodes() {
		requests = append(requests, txn.TxnRequest{
			Method: txn.TxnMethod_Write,
			CNRequest: &txn.CNOpRequest{
				OpCode: opCreateRelation,
				Payload: mustEncodePayload(createRelationReq{
					DatabaseID: d.id,
					Name:       relName,
					Defs:       defs,
				}),
				Target: metadata.DNShard{
					Address: node.ServiceAddress,
				},
			},
		})
	}

	result, err := d.txnOperator.WriteAndCommit(ctx, requests)
	if err != nil {
		return err
	}
	if err := errorFromTxnResponses(result.Responses); err != nil {
		return err
	}

	return nil
}

func (d *Database) Delete(ctx context.Context, relName string) error {

	// for ddl operations, broadcast to all DNs
	var requests []txn.TxnRequest
	for _, node := range d.engine.getDataNodes() {
		requests = append(requests, txn.TxnRequest{
			Method: txn.TxnMethod_Write,
			CNRequest: &txn.CNOpRequest{
				OpCode: opDeleteRelation,
				Payload: mustEncodePayload(deleteRelationReq{
					DatabaseID: d.id,
					Name:       relName,
				}),
				Target: metadata.DNShard{
					Address: node.ServiceAddress,
				},
			},
		})
	}

	result, err := d.txnOperator.WriteAndCommit(ctx, requests)
	if err != nil {
		return err
	}
	if err := errorFromTxnResponses(result.Responses); err != nil {
		return err
	}

	return nil
}

func (d *Database) Relation(ctx context.Context, relName string) (engine.Relation, error) {

	result, err := d.txnOperator.Read(ctx, []txn.TxnRequest{
		{
			Method: txn.TxnMethod_Read,
			CNRequest: &txn.CNOpRequest{
				OpCode: opOpenRelation,
				Payload: mustEncodePayload(openRelationReq{
					DatabaseID: d.id,
					Name:       relName,
				}),
				Target: metadata.DNShard{
					// use first DN
					Address: d.engine.getDataNodes()[0].ServiceAddress,
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

	var resp openRelationResp
	if err := gob.NewDecoder(bytes.NewReader(result.Responses[0].CNOpResponse.Payload)).Decode(&resp); err != nil {
		return nil, err
	}

	switch resp.Type {

	case RelationTable:
		table := &Table{
			id: resp.ID,
		}
		return table, nil

	default:
		panic("unknown type")
	}

}

func (d *Database) Relations(ctx context.Context) ([]string, error) {

	result, err := d.txnOperator.Read(ctx, []txn.TxnRequest{
		{
			Method: txn.TxnMethod_Read,
			CNRequest: &txn.CNOpRequest{
				OpCode: opGetRelations,
				Payload: mustEncodePayload(getRelationsReq{
					DatabaseID: d.id,
				}),
				Target: metadata.DNShard{
					// use first DN
					Address: d.engine.getDataNodes()[0].ServiceAddress,
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

	var relNames []string
	for _, resp := range result.Responses {
		var r getRelationsResp
		if err := gob.NewDecoder(bytes.NewReader(resp.CNOpResponse.Payload)).Decode(&r); err != nil {
			return nil, err
		}
		relNames = append(relNames, r.Names...)
	}

	return relNames, nil
}

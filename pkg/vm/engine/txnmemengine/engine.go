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
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

// Engine is an engine.Engine impl
type Engine struct {
}

func New() *Engine {
	return &Engine{}
}

var _ engine.Engine = new(Engine)

func (*Engine) Create(dbName string, txn client.TxnOperator) error {
	//TODO
	return nil
}

func (*Engine) Database(dbName string, txn client.TxnOperator) (engine.Database, error) {
	//TODO
	return nil, nil
}

func (*Engine) Databases(txn client.TxnOperator) []string {
	//TODO
	return nil
}

func (*Engine) Delete(dbName string, txn client.TxnOperator) error {
	//TODO
	return nil
}

func (*Engine) Nodes(txn client.TxnOperator) engine.Nodes {
	//TODO
	return nil
}

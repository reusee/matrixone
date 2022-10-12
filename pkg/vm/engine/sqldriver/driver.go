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
	"database/sql/driver"

	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type Driver struct {
	engine          engine.Engine
	txnClient       TxnClient
	compilerContext plan.CompilerContext
}

func New(
	engine engine.Engine,
	txnClient TxnClient,
	compilerContext plan.CompilerContext,
) *Driver {
	return &Driver{
		engine:          engine,
		txnClient:       txnClient,
		compilerContext: compilerContext,
	}
}

var _ driver.Driver = new(Driver)

func (d *Driver) Open(name string) (driver.Conn, error) {
	return newConn(d.engine, d.txnClient, d.compilerContext), nil
}

var _ driver.DriverContext = new(Driver)

func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	return d.newConnector(d.engine, d.txnClient, d.compilerContext), nil
}

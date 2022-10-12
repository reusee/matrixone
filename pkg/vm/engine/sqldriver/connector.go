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
	"database/sql/driver"

	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type Connector struct {
	driver          *Driver
	engine          engine.Engine
	txnClient       TxnClient
	compilerContext plan.CompilerContext
}

func (d *Driver) newConnector(
	engine engine.Engine,
	txnClient TxnClient,
	compilerContext plan.CompilerContext,
) *Connector {
	return &Connector{
		driver:          d,
		engine:          engine,
		txnClient:       txnClient,
		compilerContext: compilerContext,
	}
}

var _ driver.Connector = new(Connector)

func (c *Connector) Connect(context.Context) (driver.Conn, error) {
	return newConn(c.engine, c.txnClient, c.compilerContext), nil
}

func (c *Connector) Driver() driver.Driver {
	return c.driver
}

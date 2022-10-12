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
	"fmt"

	"github.com/matrixorigin/matrixone/pkg/sql/parsers/dialect/mysql"
	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type conn struct {
	engine          engine.Engine
	txnClient       TxnClient
	compilerContext plan.CompilerContext
}

func newConn(
	engine engine.Engine,
	txnClient TxnClient,
	compilerContext plan.CompilerContext,
) *conn {
	return &conn{
		engine:          engine,
		txnClient:       txnClient,
		compilerContext: compilerContext,
	}
}

var _ driver.Conn = new(conn)

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	stmts, err := mysql.Parse(query)
	if err != nil {
		return nil, err
	}
	if len(stmts) == 0 {
		return nil, fmt.Errorf("no statement")
	}
	if len(stmts) > 1 {
		return nil, fmt.Errorf("multiple statements")
	}
	statement := stmts[0]

	plan, err := plan.BuildPlan(c.compilerContext, statement)
	if err != nil {
		return nil, err
	}

	return &stmt{
		engine:    c.engine,
		txnClient: c.txnClient,
		query:     query,
		statement: statement,
		plan:      plan,
	}, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	panic("TODO")
}

var _ driver.ConnBeginTx = new(conn)

func (c *conn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	panic("TODO")
}

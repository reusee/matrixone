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

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/compile"
	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

type stmt struct {
	engine    engine.Engine
	txnClient TxnClient
	query     string
	statement tree.Statement
	plan      *plan.Plan
}

var _ driver.Stmt = new(stmt)

func (*stmt) Close() error {
	return nil
}

func (*stmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("TODO")
}

func (*stmt) NumInput() int {
	return -1
}

func (s *stmt) Query(args []driver.Value) (_ driver.Rows, err error) {
	ctx := context.Background()
	txnOperator, err := s.txnClient.New()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			txnOperator.Rollback(ctx)
		} else {
			err = txnOperator.Commit(ctx)
		}
	}()
	proc := process.New(
		ctx,
		memPool,
		s.txnClient,
		txnOperator,
		nil, //TODO set file service
	)
	c := compile.New("", s.query, "", context.Background(), s.engine, proc, s.statement)
	rows := new(rows)
	if err := c.Compile(s.plan, nil, func(_ any, bat *batch.Batch) error {
		rows.bat = bat
		return nil
	}); err != nil {
		return nil, err
	}
	c.GetAffectedRows()
	if err := c.Run(0); err != nil {
		return nil, err
	}
	return rows, nil
}

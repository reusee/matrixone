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
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type Table struct {
}

var _ engine.Relation = new(Table)

func (*Table) Rows() int64 {
	//TODO
	return 0
}

func (*Table) Size(string) int64 {
	//TODO
	return 0
}

func (*Table) AddTableDef(def engine.TableDef) error {
	//TODO
	return nil
}

func (*Table) DelTableDef(def engine.TableDef) error {
	//TODO
	return nil
}

func (*Table) Delete(vec *vector.Vector) error {
	//TODO
	return nil
}

func (*Table) GetHideKey() *engine.Attribute {
	//TODO
	return nil
}

func (*Table) GetPriKeyOrHideKey() ([]engine.Attribute, bool) {
	//TODO
	return nil, false
}

func (*Table) GetPrimaryKeys() []*engine.Attribute {
	//TODO
	return nil
}

func (*Table) NewReader(parallel int, expr *plan.Expr, data []byte) []engine.Reader {
	//TODO
	return nil
}

func (*Table) Nodes() engine.Nodes {
	//TODO
	return nil
}

func (*Table) TableDefs() []engine.TableDef {
	//TODO
	return nil
}

func (*Table) Truncate() (uint64, error) {
	//TODO
	return 0, nil
}

func (*Table) Update(data *batch.Batch) error {
	//TODO
	return nil
}

func (*Table) Write(data *batch.Batch) error {
	//TODO
	return nil
}

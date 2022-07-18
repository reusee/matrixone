// Copyright 2021 Matrix Origin
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

	"github.com/matrixorigin/matrixone/pkg/compress"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
)

type Nodes []Node

type Node struct {
	Mcpu int
	Id   string `json:"id"`
	Addr string `json:"address"`
	Data []byte `json:"payload"`
}

type Attribute struct {
	IsHide  bool
	Name    string      // name of attribute
	Alg     compress.T  // compression algorithm
	Type    types.Type  // type of attribute
	Default DefaultExpr // default value of this attribute.
	Primary bool        // if true, it is primary key
}

type DefaultExpr struct {
	Exist  bool
	Expr   *plan.Expr
	IsNull bool
}

type Statistics interface {
	Rows() int64
	Size(string) int64
}

type AttributeDef struct {
	Attr Attribute
}

type CommentDef struct {
	Comment string
}

type TableDef interface {
	tableDef()
}

func (*CommentDef) tableDef()   {}
func (*AttributeDef) tableDef() {}

type Relation interface {
	Statistics

	Ranges(context.Context) [][]byte

	TableDefs(context.Context) ([]TableDef, error)

	GetPrimaryKeys(context.Context) ([]*Attribute, error)

	GetHideKey() *Attribute
	// true: primary key, false: hide key
	GetPriKeyOrHideKey() ([]Attribute, bool)

	Write(context.Context, *batch.Batch) error

	Update(context.Context, *batch.Batch) error

	Delete(context.Context, *vector.Vector, string) error

	Truncate(context.Context) (uint64, error)

	AddTableDef(context.Context, TableDef) error
	DelTableDef(context.Context, TableDef) error

	// first argument is the number of reader, second argument is the filter extend,  third parameter is the payload required by the engine
	NewReader(context.Context, int, *plan.Expr, [][]byte) ([]Reader, error)
}

type Reader interface {
	Read([]string) (*batch.Batch, error)
	Close() error
}

type Database interface {
	Relations(context.Context) ([]string, error)
	Relation(context.Context, string) (Relation, error)

	Delete(context.Context, string) error
	Create(context.Context, string, []TableDef) error // Create Table - (name, table define)
}

type Engine interface {
	Delete(context.Context, string, client.TxnOperator) error
	Create(context.Context, string, client.TxnOperator) error // Create Database - (name, engine type)

	Databases(context.Context, client.TxnOperator) ([]string, error)
	Database(context.Context, string, client.TxnOperator) (Database, error)

	Nodes() Nodes
}

// MakeDefaultExpr returns a new DefaultExpr
func MakeDefaultExpr(exist bool, expr *plan.Expr, isNull bool) DefaultExpr {
	return DefaultExpr{
		Expr:   expr,
		Exist:  exist,
		IsNull: isNull,
	}
}

// EmptyDefaultExpr means there is no definition for default expr
var EmptyDefaultExpr = DefaultExpr{Exist: false}

func (node Attribute) HasDefaultExpr() bool {
	return node.Default.Exist
}

func (node Attribute) GetDefaultExpr() (*plan.Expr, bool) {
	return node.Default.Expr, node.Default.IsNull
}

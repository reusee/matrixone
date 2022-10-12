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

package memorystorage

import (
	"context"
	"database/sql"
	"database/sql/driver"

	mysqldriver "github.com/dolthub/go-mysql-server/driver"
	mysqlsql "github.com/dolthub/go-mysql-server/sql"
	"github.com/jmoiron/sqlx"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

// SQL returns an interface for testing SQL execution
func (s *StorageTxnClient) SQL(engine engine.Engine) (*sqlx.DB, error) {
	provider := &sqlProvider{
		engine:    engine,
		txnClient: s,
	}
	driver := mysqldriver.New(provider, &mysqldriver.Options{})
	connector := &sqlConnector{
		engine: engine,
		driver: driver,
	}
	db := sql.OpenDB(connector)
	return sqlx.NewDb(db, ""), nil
}

type sqlProvider struct {
	engine    engine.Engine
	txnClient TxnClient
}

var _ mysqldriver.Provider = new(sqlProvider)

func (s *sqlProvider) Resolve(name string, options *mysqldriver.Options) (string, mysqlsql.DatabaseProvider, error) {
	provider, err := newDBProvider(s.txnClient, s.engine)
	if err != nil {
		return "", nil, err
	}
	return name, provider, nil
}

type dbProvider struct {
	ctx    context.Context
	engine engine.Engine
	op     TxnOperator
}

func newDBProvider(txnClient TxnClient, engine engine.Engine) (*dbProvider, error) {
	op, err := txnClient.New()
	if err != nil {
		return nil, err
	}
	return &dbProvider{
		ctx:    context.Background(),
		engine: engine,
		op:     op,
	}, nil
}

var _ mysqlsql.DatabaseProvider = new(dbProvider)

func (s *dbProvider) AllDatabases(ctx *mysqlsql.Context) (ret []mysqlsql.Database) {
	names, err := s.engine.Databases(s.ctx, s.op)
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		ret = append(ret, &mysqlDB{
			name: name,
		})
	}
	return
}

func (*dbProvider) Database(ctx *mysqlsql.Context, name string) (mysqlsql.Database, error) {
	//TODO
	panic("unimplemented")
}

func (s *dbProvider) HasDatabase(ctx *mysqlsql.Context, name string) bool {
	_, err := s.engine.Database(s.ctx, name, s.op)
	if err != nil {
		panic(err)
	}
	return err == nil
}

var _ mysqlsql.MutableDatabaseProvider = new(dbProvider)

func (*dbProvider) CreateDatabase(ctx *mysqlsql.Context, name string) error {
	panic("unimplemented")
}

func (*dbProvider) DropDatabase(ctx *mysqlsql.Context, name string) error {
	panic("unimplemented")
}

type sqlConnector struct {
	driver *mysqldriver.Driver
	engine engine.Engine
}

var _ driver.Connector = new(sqlConnector)

func (s *sqlConnector) Connect(context.Context) (driver.Conn, error) {
	return s.driver.Open("catalog")
}

func (s *sqlConnector) Driver() driver.Driver {
	return s.driver
}

type mysqlDB struct {
	name string
}

var _ mysqlsql.Database = new(mysqlDB)

// Name implements sql.Database
func (m *mysqlDB) Name() string {
	return m.name
}

// GetTableInsensitive implements sql.Database
func (*mysqlDB) GetTableInsensitive(ctx *mysqlsql.Context, tblName string) (mysqlsql.Table, bool, error) {
	//TODO
	panic("unimplemented")
}

// GetTableNames implements sql.Database
func (*mysqlDB) GetTableNames(ctx *mysqlsql.Context) ([]string, error) {
	//TODO
	panic("unimplemented")
}

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

package sqlitestorage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func New() (*Storage, error) {
	db, err := sql.Open("sqlite3", "file:db?mode=memory")
	if err != nil {
		return nil, err
	}

	//TODO re-design
	// 使用单一的表，使用字段的并集
	// partial index, view 等等都可以用上
	stmts := []string{

		0: `
    create table databases (
      id integer primary key autoincrement,
      tx_id integer,
      physical_time integer not null,
      logical_time integer not null,

      name text,

      foreign key(tx_id) references transactions(id)
    );
    `,

		1: `
    create table relations (
      id integer primary key autoincrement,
      tx_id integer,
      physical_time integer not null,
      logical_time integer not null,

      name text not null,
      database_id integer not null,

      foreign key(tx_id) references transactions(id),
      foreign key(database_id) references databases(id)
    );
    `,

		2: `
    create table attributes (
      id integer primary key autoincrement,
      tx_id integer,
      physical_time integer not null,
      logical_time integer not null,

      name text not null,
      type text not null,
      table_id integer not null,

      foreign key(tx_id) references transactions(id)
      foreign key(table_id) references tables(id)
    );
    `,

		3: `
    create table rows (
      id integer primary key autoincrement,
      tx_id integer,
      physical_time integer not null,
      logical_time integer not null,

      data json not null,
      table_id integer not null,

      foreign key(tx_id) references transactions(id),
      foreign key(table_id) references tables(id)
    );
    `,

		4: `
    create table transactions (
      id text primary key
    );
    `,
	}

	for _, stmt := range stmts {
		_, err = db.Exec(stmt)
		if err != nil {
			return nil, err
		}
	}

	s := &Storage{
		db: db,
	}

	return s, nil
}

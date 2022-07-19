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

	_ "modernc.org/sqlite"
)

func New() (*Storage, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	stmts := []string{
		0: `
    create table databases (
      id integer primary key autoincrement
    );
    `,

		1: `
    create table relations (
      id integer primary key autoincrement,
      database_id integer not null,
      name text not null,
      foreign key(database_id) references databases(id)
    );
    `,

		2: `
    create table attributes (
      id integer primary key autoincrement,
      table_id integer not null,
      name text not null,
      type text not null,
      foreign key(table_id) references tables(id)
    );
    `,

		3: `
    create table rows (
      id integer primary key autoincrement,
      table_id integer not null,
      data json not null,
      foreign key(table_id) references tables(id)
    );
    `,
	}

	for _, stmt := range stmts {
		_, err = db.Exec(stmt)
		if err != nil {
			return nil, err
		}
	}

	return &Storage{
		db: db,
	}, nil

}

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

	commonAttrs := `
    _row_id integer primary key autoincrement,
    min_tx_id text not null references transactions(id),
    min_physical_time integer not null,
    min_logical_time integer not null,
    max_tx_id text,
    max_physical_time integer,
    max_logical_time integer,
  `

	stmts := []string{

		0: `
    create table databases (
      ` + commonAttrs + `
      id text not null,
      name text not null
    );
    `,

		1: `
    create table relations (
      ` + commonAttrs + `
      id text not null,
      name text not null,
      database_id text not null
    );
    `,

		2: `
    create table attributes (
      ` + commonAttrs + `
      id text not null,
      name text not null,
      type text not null,
      table_id text not null
    );
    `,

		3: `
    create table rows (
      ` + commonAttrs + `
      id text not null,
      data json not null,
      table_id text not null
    );
    `,

		4: `
    create table transactions (
      id text primary key,
      snapshot_physical_time integer not null,
      snapshot_logical_time integer not null,
      commit_physical_time integer,
      commit_logical_time integer
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

// Copyright 2021 - 2022 Matrix Origin
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

package memdbstorage

import (
	"github.com/hashicorp/go-memdb"
)

func New() (*Storage, error) {

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"databases": databasesSchema,
			"relations": relationsSchema,
			"attrs":     attrsSchema,
			"rows":      rowsSchema,
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		db: db,
	}

	return storage, nil
}

var databasesSchema = &memdb.TableSchema{
	Name: "databases",
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.IntFieldIndex{Field: "ID"},
		},
	},
}

var relationsSchema = &memdb.TableSchema{
	Name: "relations",
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.IntFieldIndex{Field: "ID"},
		},
	},
}

var attrsSchema = &memdb.TableSchema{
	Name: "attrs",
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.IntFieldIndex{Field: "ID"},
		},
	},
}

var rowsSchema = &memdb.TableSchema{
	Name: "rows",
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.IntFieldIndex{Field: "ID"},
		},
	},
}

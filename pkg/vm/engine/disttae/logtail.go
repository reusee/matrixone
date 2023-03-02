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

package disttae

import (
	"context"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/pb/api"
)

func consumeEntry(
	ctx context.Context,
	primaryIdx int,
	engine *Engine,
	partition *Partition,
	state *PartitionState,
	e *api.Entry,
) error {

	state.HandleLogtailEntry(ctx, e, primaryIdx)

	if e.EntryType == api.Entry_Insert {
		if isMetaTable(e.TableName) {
			return nil
		}
		switch e.TableId {
		case catalog.MO_TABLES_ID:
			bat, _ := batch.ProtoBatchToBatch(e.Bat)
			engine.catalog.InsertTable(bat)
		case catalog.MO_DATABASE_ID:
			bat, _ := batch.ProtoBatchToBatch(e.Bat)
			engine.catalog.InsertDatabase(bat)
		case catalog.MO_COLUMNS_ID:
			bat, _ := batch.ProtoBatchToBatch(e.Bat)
			engine.catalog.InsertColumns(bat)
		}
		if primaryIdx >= 0 {
			return partition.Insert(ctx, MO_PRIMARY_OFF+primaryIdx, e.Bat, false)
		}
		return partition.Insert(ctx, primaryIdx, e.Bat, false)
	}
	if isMetaTable(e.TableName) {
		return nil
	}
	switch e.TableId {
	case catalog.MO_TABLES_ID:
		bat, _ := batch.ProtoBatchToBatch(e.Bat)
		engine.catalog.DeleteTable(bat)
	case catalog.MO_DATABASE_ID:
		bat, _ := batch.ProtoBatchToBatch(e.Bat)
		engine.catalog.DeleteDatabase(bat)
	}
	return partition.Delete(ctx, e.Bat)
}

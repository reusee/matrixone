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

import "github.com/matrixorigin/matrixone/pkg/pb/txn"

func (a *Args) visible(meta txn.TxnMeta) string {
	return `not (

    -- future write
    (
      ` + a.bind(meta.SnapshotTS.PhysicalTime) + ` < min_physical_time
      or (
        ` + a.bind(meta.SnapshotTS.PhysicalTime) + ` = min_physical_time
        and ` + a.bind(meta.SnapshotTS.LogicalTime) + ` < min_logical_time
      )
    )

    -- uncommit non-local write
    or
    (
      min_tx_id <> ` + a.bind(string(meta.ID)) + `
      and (
        select commit_physical_time is null
        from transactions t
        where t.id = min_tx_id
      )
    )

    -- deleted
    or (
      max_tx_id is not null
      and max_tx_id <> ` + a.bind(string(meta.ID)) + `
      and (
        ` + a.bind(meta.SnapshotTS.PhysicalTime) + ` > max_physical_time
        or (
          ` + a.bind(meta.SnapshotTS.PhysicalTime) + ` = max_physical_time
          and ` + a.bind(meta.SnapshotTS.LogicalTime) + ` > max_logical_time
        )
      )
      and (
        select commit_physical_time is not null
        from transactions t
        where t.id = max_tx_id
      )
    )

  )`
}

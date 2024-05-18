// Copyright 2024 Matrix Origin
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

package v1_2_0

import (
	"context"

	"go.uber.org/zap"

	"github.com/matrixorigin/matrixone/pkg/bootstrap/versions"
	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/util/executor"
)

var (
	Handler = &versionHandle{
		metadata: versions.Version{
			Version:           "1.2.0",
			MinUpgradeVersion: "1.1.0",
			UpgradeCluster:    versions.Yes,
			UpgradeTenant:     versions.Yes,
			VersionOffset:     uint32(len(clusterUpgEntries) + len(tenantUpgEntries)),
		},
	}
)

type versionHandle struct {
	metadata versions.Version
}

func (v *versionHandle) Metadata() versions.Version {
	return v.metadata
}

func (v *versionHandle) Prepare(
	ctx context.Context,
	txn executor.TxnExecutor,
	final bool) error {

	for _, upgEntry := range UpgPrepareEntres {
		err := upgEntry.Upgrade(txn, catalog.System_Account)
		if err != nil {
			getLogger().Error("prepare upgrade entry execute error", zap.Error(err), zap.String("version", v.Metadata().Version), zap.String("upgrade entry", upgEntry.String()))
			return err
		}
	}

	return nil
}

func (v *versionHandle) HandleTenantUpgrade(
	ctx context.Context,
	tenantID int32,
	txn executor.TxnExecutor) error {

	for _, upgEntry := range tenantUpgEntries {
		err := upgEntry.Upgrade(txn, uint32(tenantID))
		if err != nil {
			getLogger().Error("tenant upgrade entry execute error", zap.Error(err), zap.Int32("tenantId", tenantID), zap.String("version", v.Metadata().Version), zap.String("upgrade entry", upgEntry.String()))
			return err
		}
	}

	return nil
}

func (v *versionHandle) HandleClusterUpgrade(
	ctx context.Context,
	txn executor.TxnExecutor) error {
	txn.Use(catalog.MO_CATALOG)
	for _, upgEntry := range clusterUpgEntries {
		err := upgEntry.Upgrade(txn, catalog.System_Account)
		if err != nil {
			getLogger().Error("cluster upgrade entry execute error", zap.Error(err), zap.String("version", v.Metadata().Version), zap.String("upgrade entry", upgEntry.String()))
			return err
		}
	}
	return nil
}

func (v *versionHandle) HandleCreateFrameworkDeps(txn executor.TxnExecutor) error {
	// create new upgrade framework tables for the first time,
	// which means using v1.2.0 for the first time
	// NOTE: The `alter table` statements used for upgrading system table rely on `mo_foreign_keys` and `mo_indexes`,
	// so preprocessing is performed first
	for _, upgEntry := range createFrameworkDepsEntres {
		err := upgEntry.Upgrade(txn, catalog.System_Account)
		if err != nil {
			getLogger().Error("Handle create framework dependencies upgrade entry execute error", zap.Error(err), zap.String("upgrade entry", upgEntry.String()))
			return err
		}
	}
	return nil
}

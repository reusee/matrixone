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

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cockroachdb/pebble"
	"github.com/matrixorigin/matrixcube/pb/metapb"
	"github.com/matrixorigin/matrixcube/storage"
	"github.com/matrixorigin/matrixcube/storage/kv"
	cPebble "github.com/matrixorigin/matrixcube/storage/kv/pebble"
	"github.com/matrixorigin/matrixcube/vfs"
	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/config"
	"github.com/matrixorigin/matrixone/pkg/frontend"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/sql/compile"
	"github.com/matrixorigin/matrixone/pkg/vm/driver"
	aoeDriver "github.com/matrixorigin/matrixone/pkg/vm/driver/aoe"
	dConfig "github.com/matrixorigin/matrixone/pkg/vm/driver/config"
	kvDriver "github.com/matrixorigin/matrixone/pkg/vm/driver/kv"
	"github.com/matrixorigin/matrixone/pkg/vm/driver/pb"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	aoeEngine "github.com/matrixorigin/matrixone/pkg/vm/engine/aoe/engine"
	aoeStorage "github.com/matrixorigin/matrixone/pkg/vm/engine/aoe/storage"
)

//TODO remove these global vars
var (
	c   *catalog.Catalog
	mo  *frontend.MOServer
	pci *frontend.PDCallbackImpl
)

type aoeHandler struct {
	cube       driver.CubeDriver
	port       int64
	kvStorage  storage.DataStorage
	aoeStorage storage.DataStorage
	eng        engine.Engine
}

func initAoe(configFilePath string) *aoeHandler {

	targetDir := config.GlobalSystemVariables.GetStorePath()
	if err := recreateDir(targetDir); err != nil {
		logutil.Infof("Recreate dir error:%v\n", err)
		os.Exit(RecreateDirExit)
	}

	cfg := parseConfig(configFilePath, targetDir)

	//aoe : kvstorage config
	_, kvStorage := getKVDataStorage(targetDir, cfg)

	//aoe : catalog
	catalogListener := catalog.NewCatalogListener()
	aoeStorage := getAOEDataStorage(configFilePath, targetDir, catalogListener, cfg)

	//aoe cube driver
	a, err := driver.NewCubeDriverWithOptions(kvStorage, aoeStorage, &cfg)
	if err != nil {
		logutil.Infof("Create cube driver failed, %v", err)
		os.Exit(CreateCubeExit)
	}
	err = a.Start()
	if err != nil {
		logutil.Infof("Start cube driver failed, %v", err)
		os.Exit(StartCubeExit)
	}

	//aoe: address for computation
	addr := cfg.CubeConfig.AdvertiseClientAddr
	if len(addr) != 0 {
		logutil.Infof("compile init address from cube AdvertiseClientAddr %s", addr)
	} else {
		logutil.Infof("compile init address from cube ClientAddr %s", cfg.CubeConfig.ClientAddr)
		addr = cfg.CubeConfig.ClientAddr
	}

	//put the node info to the computation
	compile.InitAddress(addr)

	//aoe: catalog
	c = catalog.NewCatalog(a)
	config.ClusterCatalog = c
	catalogListener.UpdateCatalog(c)
	cngineConfig := aoeEngine.EngineConfig{}
	_, err = toml.DecodeFile(configFilePath, &cngineConfig)
	if err != nil {
		logutil.Infof("Decode cube config error:%v\n", err)
		os.Exit(DecodeCubeConfigExit)
	}

	eng := aoeEngine.New(c, &cngineConfig)

	err = waitClusterStartup(a, 300*time.Second, int(cfg.CubeConfig.Prophet.Replication.MaxReplicas), int(cfg.ClusterConfig.PreAllocatedGroupNum))

	if err != nil {
		logutil.Infof("wait cube cluster startup failed, %v", err)
		os.Exit(WaitCubeStartExit)
	}

	//test storage aoe_storage
	config.StorageEngine = eng

	li := strings.LastIndex(cfg.CubeConfig.ClientAddr, ":")
	if li == -1 {
		logutil.Infof("There is no port in client addr")
		os.Exit(LoadConfigExit)
	}
	cubePort, err := strconv.ParseInt(string(cfg.CubeConfig.ClientAddr[li+1:]), 10, 32)
	if err != nil {
		logutil.Infof("Invalid port")
		os.Exit(LoadConfigExit)
	}
	return &aoeHandler{
		cube:       a,
		port:       cubePort,
		kvStorage:  kvStorage,
		aoeStorage: aoeStorage,
		eng:        eng,
	}
}

/*
*
call the catalog service to remove the epoch
*/
func removeEpoch(epoch uint64) {
	//logutil.Infof("removeEpoch %d",epoch)
	var err error
	if c != nil {
		_, err = c.RemoveDeletedTable(epoch)
		if err != nil {
			fmt.Printf("catalog remove ddl failed. error :%v \n", err)
		}
	}
}

func closeAoe(aoe *aoeHandler) {
	aoe.kvStorage.Close()
	aoe.aoeStorage.Close()
	aoe.cube.Close()
}

func waitClusterStartup(driver driver.CubeDriver, timeout time.Duration, maxReplicas int, minimalAvailableShard int) error {
	timeoutC := time.After(timeout)
	for {
		select {
		case <-timeoutC:
			return errors.New("wait for available shard timeout")
		default:
			router := driver.RaftStore().GetRouter()
			if router != nil {
				nodeCnt := maxReplicas
				shardCnt := 0
				router.ForeachShards(uint64(pb.AOEGroup), func(shard metapb.Shard) bool {
					fmt.Printf("shard %d, peer count is %d\n", shard.ID, len(shard.Replicas))
					shardCnt++
					if len(shard.Replicas) < nodeCnt {
						nodeCnt = len(shard.Replicas)
					}
					return true
				})
				if nodeCnt >= maxReplicas && shardCnt >= minimalAvailableShard {
					kvNodeCnt := maxReplicas
					kvCnt := 0
					router.ForeachShards(uint64(pb.KVGroup), func(shard metapb.Shard) bool {
						kvCnt++
						if len(shard.Replicas) < kvNodeCnt {
							kvNodeCnt = len(shard.Replicas)
						}
						return true
					})
					if kvCnt >= 1 && kvNodeCnt >= maxReplicas {
						fmt.Println("ClusterStatus is ok now")
						return nil
					}

				}
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func parseConfig(configFilePath, targetDir string) dConfig.Config {
	cfg := dConfig.Config{}
	_, err := toml.DecodeFile(configFilePath, &cfg.CubeConfig)
	if err != nil {
		logutil.Infof("Decode cube config error:%v\n", err)
		os.Exit(DecodeCubeConfigExit)
	}
	_, err = toml.DecodeFile(configFilePath, &cfg.FeaturesConfig)
	if err != nil {
		logutil.Infof("Decode cube config error:%v\n", err)
		os.Exit(DecodeCubeConfigExit)
	}

	cfg.CubeConfig.DataPath = targetDir + "/cube"
	_, err = toml.DecodeFile(configFilePath, &cfg.ClusterConfig)
	if err != nil {
		logutil.Infof("Decode cluster config error:%v\n", err)
		os.Exit(DecodeClusterConfigExit)
	}

	if !config.GlobalSystemVariables.GetDisablePCI() {
		cfg.CubeConfig.Customize.CustomStoreHeartbeatDataProcessor = pci
	}
	cfg.CubeConfig.Logger = logutil.GetGlobalLogger()
	return cfg
}

func getKVDataStorage(targetDir string, cfg dConfig.Config) (*cPebble.Storage, storage.DataStorage) {
	kvs, err := cPebble.NewStorage(targetDir+"/pebble/data", nil, &pebble.Options{
		FS:                          vfs.NewPebbleFS(vfs.Default),
		MemTableSize:                1024 * 1024 * 128,
		MemTableStopWritesThreshold: 4,
	})
	if err != nil {
		logutil.Infof("create kv data storage error, %v\n", err)
		os.Exit(CreateAoeExit)
	}

	kvBase := kv.NewBaseStorage(kvs, vfs.Default)
	return kvs, kv.NewKVDataStorage(kvBase, kvDriver.NewkvExecutor(kvs),
		kv.WithLogger(cfg.CubeConfig.Logger),
		kv.WithFeature(cfg.FeaturesConfig.KV.Feature()))
}

func getAOEDataStorage(configFilePath, targetDir string,
	catalogListener *catalog.CatalogListener,
	cfg dConfig.Config) storage.DataStorage {
	var aoeDataStorage *aoeDriver.Storage
	opt := aoeStorage.Options{}
	_, err := toml.DecodeFile(configFilePath, &opt)
	if err != nil {
		logutil.Infof("Decode aoe config error:%v\n", err)
		os.Exit(DecodeAoeConfigExit)
	}

	opt.EventListener = catalogListener
	aoeDataStorage, err = aoeDriver.NewStorageWithOptions(targetDir+"/aoe",
		cfg.FeaturesConfig.AOE.Feature(), &opt)
	if err != nil {
		logutil.Infof("Create aoe driver error, %v\n", err)
		os.Exit(CreateAoeExit)
	}
	return aoeDataStorage
}

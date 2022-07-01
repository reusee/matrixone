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
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	ie "github.com/matrixorigin/matrixone/pkg/util/internalExecutor"
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/db"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/moengine"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/config"
	"github.com/matrixorigin/matrixone/pkg/frontend"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/rpcserver"
	"github.com/matrixorigin/matrixone/pkg/sql/handler"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/guest"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/host"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

const (
	InitialValuesExit       = 1
	LoadConfigExit          = 2
	RecreateDirExit         = 3
	DecodeCubeConfigExit    = 6
	DecodeClusterConfigExit = 7
	CreateCubeExit          = 8
	StartCubeExit           = 9
	CreateRPCExit           = 10
	WaitCubeStartExit       = 11
	StartMOExit             = 12
	CreateTpeExit           = 13
	RunRPCExit              = 14
	ShutdownExit            = 15
	CreateTaeExit           = 16
	InitCatalogExit         = 17
)

var (
	c   *catalog.Catalog
	mo  *frontend.MOServer
	pci *frontend.PDCallbackImpl
)

type StartServer func()

func (_ Def) StartServer(
	posArgs PositionalArguments,
) (
	start StartServer,
) {

	start = func() {
		args := *posArgs

		configFilePath := args[0]
		logutil.SetupMOLogger(configFilePath)

		//before anything using the configuration
		if err := config.GlobalSystemVariables.LoadInitialValues(); err != nil {
			logutil.Infof("Initial values error:%v\n", err)
			os.Exit(InitialValuesExit)
		}

		if err := config.LoadvarsConfigFromFile(
			configFilePath,
			&config.GlobalSystemVariables,
		); err != nil {
			logutil.Infof("Load config error:%v\n", err)
			os.Exit(LoadConfigExit)
		}

		//just initialize the tae after configuration has been loaded
		if len(args) == 2 && args[1] == "initdb" {
			fmt.Println("Initialize the TAE engine ...")
			taeWrapper := initTae()
			err := frontend.InitDB(taeWrapper.eng)
			if err != nil {
				logutil.Infof("Initialize catalog failed. error:%v", err)
				os.Exit(InitCatalogExit)
			}
			fmt.Println("Initialize the TAE engine Done")
			closeTae(taeWrapper)
			os.Exit(0)
		}

		logutil.Infof("Shutdown The Server With Ctrl+C | Ctrl+\\.")

		config.HostMmu = host.New(
			config.GlobalSystemVariables.GetHostMmuLimitation(),
		)

		Host := config.GlobalSystemVariables.GetHost()
		NodeId := config.GlobalSystemVariables.GetNodeID()

		ppu := frontend.NewPDCallbackParameterUnit(
			int(config.GlobalSystemVariables.GetPeriodOfEpochTimer()),
			int(config.GlobalSystemVariables.GetPeriodOfPersistence()),
			int(config.GlobalSystemVariables.GetPeriodOfDDLDeleteTimer()),
			int(config.GlobalSystemVariables.GetTimeoutOfHeartbeat()),
			config.GlobalSystemVariables.GetEnableEpochLogging(),
			math.MaxInt64,
		)

		pci = frontend.NewPDCallbackImpl(ppu)
		pci.Id = int(NodeId)

		engineName := config.GlobalSystemVariables.GetStorageEngine()
		var port int64
		port = config.GlobalSystemVariables.GetPortOfRpcServerInComputationEngine()

		var tae *taeHandler
		if engineName == "tae" {
			fmt.Println("Initialize the TAE engine ...")
			tae = initTae()
			if config.GlobalSystemVariables.GetNeedInitdb() {
				err := frontend.InitDB(tae.eng)
				if err != nil {
					logutil.Infof("Initialize catalog failed. error:%v", err)
					os.Exit(InitCatalogExit)
				}
			}
			fmt.Println("Initialize the TAE engine Done")
		} else {
			logutil.Errorf("undefined engine %s", engineName)
			os.Exit(LoadConfigExit)
		}

		srv, err := rpcserver.New(
			net.JoinHostPort(Host, strconv.FormatInt(port+100, 10)),
			1<<30,
			logutil.GetGlobalLogger(),
		)
		if err != nil {
			logutil.Infof("Create rpcserver failed, %v", err)
			os.Exit(CreateRPCExit)
		}
		hm := host.New(1 << 40)
		gm := guest.New(1<<40, hm)
		proc := process.New(mheap.New(gm))
		hp := handler.New(config.StorageEngine, proc)
		srv.Register(hp.Process)

		go func() {
			if err := srv.Run(); err != nil {
				logutil.Infof("Start rpcserver failed, %v", err)
				os.Exit(RunRPCExit)
			}
		}()

		createMOServer(pci)

		err = mo.Start()
		if err != nil {
			logutil.Infof("Start MOServer failed, %v", err)
			os.Exit(StartMOExit)
		}

		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)
		<-sigchan

		if err := mo.Stop(); err != nil {
			logutil.Infof("Server shutdown failed, %v", err)
			os.Exit(ShutdownExit)
		}

		fmt.Println("\rBye!")

		if engineName == "tae" {
			closeTae(tae)
		}

	}
	return
}

func createMOServer(callback *frontend.PDCallbackImpl) {
	address := net.JoinHostPort(
		config.GlobalSystemVariables.GetHost(),
		strconv.FormatInt(config.GlobalSystemVariables.GetPort(), 10),
	)
	pu := config.NewParameterUnit(
		&config.GlobalSystemVariables,
		config.HostMmu,
		config.Mempool,
		config.StorageEngine,
		config.ClusterNodes,
		config.ClusterCatalog,
	)
	mo = frontend.NewMOServer(address, pu, callback)
	ieFactory := func() ie.InternalExecutor {
		return frontend.NewIternalExecutor(pu, callback)
	}
	metric.InitMetric(ieFactory, pu, callback.Id, metric.ALL_IN_ONE_MODE)
	frontend.InitServerVersion(MoVersion)
}

func recreateDir(dir string) (err error) {
	mask := syscall.Umask(0)
	defer syscall.Umask(mask)
	err = os.MkdirAll(dir, os.FileMode(0755))
	return err
}

type taeHandler struct {
	eng engine.Engine
	tae *db.DB
}

func initTae() *taeHandler {
	targetDir := config.GlobalSystemVariables.GetStorePath()
	if err := recreateDir(targetDir); err != nil {
		logutil.Infof("Recreate dir error:%v\n", err)
		os.Exit(RecreateDirExit)
	}

	tae, err := db.Open(targetDir+"/tae", nil)
	if err != nil {
		logutil.Infof("Open tae failed. error:%v", err)
		os.Exit(CreateTaeExit)
	}

	eng := moengine.NewEngine(tae)

	//test storage aoe_storage
	config.StorageEngine = eng

	return &taeHandler{
		eng: eng,
		tae: tae,
	}
}

func closeTae(tae *taeHandler) {
	_ = tae.tae.Close()
}

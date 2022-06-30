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

	"github.com/matrixorigin/matrixone/pkg/config"
	"github.com/matrixorigin/matrixone/pkg/frontend"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/rpcserver"
	"github.com/matrixorigin/matrixone/pkg/sql/handler"
	ie "github.com/matrixorigin/matrixone/pkg/util/internalExecutor"
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/guest"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/host"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

type StartServer func()

func (_ Def) Server(
	_ LoggerOK,
	_ ConfigOK,
	configFilePath ConfigFilePath,
) (
	start StartServer,
) {

	start = func() {

		config.HostMmu = host.New(config.GlobalSystemVariables.GetHostMmuLimitation())

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
		pci.SetRemoveEpoch(removeEpoch)

		engineName := config.GlobalSystemVariables.GetStorageEngine()
		var port int64
		port = config.GlobalSystemVariables.GetPortOfRpcServerInComputationEngine()

		//aoe : epochgc ?
		var aoe *aoeHandler
		var tae *taeHandler
		if engineName == "aoe" {
			aoe = initAoe(string(configFilePath))
			port = aoe.port
		} else if engineName == "tae" {
			fmt.Println("Initialize the TAE engine ...")
			tae = initTae()
			err := frontend.InitDB(tae.eng)
			if err != nil {
				logutil.Infof("Initialize catalog failed. error:%v", err)
				os.Exit(InitCatalogExit)
			}
			fmt.Println("Initialize the TAE engine Done")
		} else {
			logutil.Errorf("undefined engine %s", engineName)
			os.Exit(LoadConfigExit)
		}

		srv, err := rpcserver.New(
			net.JoinHostPort(Host, strconv.FormatInt(port, 10)),
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

		err = runMOServer()
		if err != nil {
			logutil.Infof("Start MOServer failed, %v", err)
			os.Exit(StartMOExit)
		}

		waitSignal()
		//srv.Stop()
		if err := serverShutdown(true); err != nil {
			logutil.Infof("Server shutdown failed, %v", err)
			os.Exit(ShutdownExit)
		}

		fmt.Println("\rBye!")

		if engineName == "aoe" {
			closeAoe(aoe)
		} else if engineName == "tae" {
			closeTae(tae)
		}

	}

	return
}

func createMOServer(callback *frontend.PDCallbackImpl) {
	address := fmt.Sprintf("%s:%d", config.GlobalSystemVariables.GetHost(), config.GlobalSystemVariables.GetPort())
	pu := config.NewParameterUnit(&config.GlobalSystemVariables, config.HostMmu, config.Mempool, config.StorageEngine, config.ClusterNodes, config.ClusterCatalog)
	mo = frontend.NewMOServer(address, pu, callback)
	if config.GlobalSystemVariables.GetEnableMetric() {
		ieFactory := func() ie.InternalExecutor {
			return frontend.NewIternalExecutor(pu, callback)
		}
		metric.InitMetric(ieFactory, pu, callback.Id, metric.ALL_IN_ONE_MODE)
	}
	frontend.InitServerVersion(MoVersion)
}

func runMOServer() error {
	return mo.Start()
}

func serverShutdown(isgraceful bool) error {
	return mo.Stop()
}

func waitSignal() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)
	<-sigchan
}

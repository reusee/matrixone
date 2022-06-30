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
	"os"

	"github.com/matrixorigin/matrixone/pkg/config"
	"github.com/matrixorigin/matrixone/pkg/frontend"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/db"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/moengine"
)

func (_ Def) TAE(
	on On,
) (
	parsers ArgumentParsers,
	usages Usages,
) {

	var p Parser

	// initdb command
	parsers = append(parsers, p.MatchStr("initdb")(
		p.End(func() {
			on(evInit, func(
				_ LoggerOK,
				_ ConfigOK,
			) {
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
			})
		})))
	usages = append(usages, [2]string{"initdb", "initialize TAE database"})

	return
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

// Copyright 2023 Matrix Origin
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

package external

import (
	"encoding/csv"

	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

type LineReader struct {
	csvReader *csv.Reader
	batchSize int
	buffer    [][]string
}

func newLineReader(param *ExternalParam, proc *process.Process) (*LineReader, error) {
	var err error
	param.reader, err = readFile(param, proc)
	if err != nil || param.reader == nil {
		return nil, err
	}
	param.reader, err = getUnCompressReader(param.Extern, param.Fileparam.Filepath, param.reader)
	if err != nil {
		return nil, err
	}

	var cma byte
	if param.Extern.Tail.Fields == nil {
		cma = ','
		param.Close = 0
	} else {
		cma = param.Extern.Tail.Fields.Terminated[0]
		param.Close = param.Extern.Tail.Fields.EnclosedBy
	}
	if param.Extern.Format == tree.JSONLINE {
		cma = '\t'
	}
	lineReader := &LineReader{
		csvReader: newReaderWithOptions(param.reader, rune(cma), '#', true, false),
		buffer:    param.LinesBuffer,
	}
	return lineReader, nil
}

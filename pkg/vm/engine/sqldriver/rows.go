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

package engine

import (
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
)

type rows struct {
	bat *batch.Batch
}

var _ driver.Rows = new(rows)

func (r *rows) Close() error {
	return nil
}

func (r *rows) Columns() []string {
	fmt.Printf("%v\n", r.bat.Attrs)
	return r.bat.Attrs
}

func (*rows) Next(dest []driver.Value) error {
	//TODO
	return io.EOF
}

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
	"bytes"
	"encoding/gob"
)

func mustEncodePayload(o any) []byte {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(o); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type createDatabasePayload struct {
	Name string
}

type openDatabasePayload struct {
	Name string
}

type getDatabasesPayload struct {
	Names []string
}

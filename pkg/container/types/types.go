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

package types

import (
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
)

const (
	// any family
	T_any uint8 = uint8(plan.Type_ANY)

	// bool family
	T_bool uint8 = uint8(plan.Type_BOOL)

	// numeric/integer family
	T_int8   uint8 = uint8(plan.Type_INT8)
	T_int16  uint8 = uint8(plan.Type_INT16)
	T_int32  uint8 = uint8(plan.Type_INT32)
	T_int64  uint8 = uint8(plan.Type_INT64)
	T_uint8  uint8 = uint8(plan.Type_UINT8)
	T_uint16 uint8 = uint8(plan.Type_UINT16)
	T_uint32 uint8 = uint8(plan.Type_UINT32)
	T_uint64 uint8 = uint8(plan.Type_UINT64)

	// numeric/float family - unsigned attribute is deprecated
	T_float32 uint8 = uint8(plan.Type_FLOAT32)
	T_float64 uint8 = uint8(plan.Type_FLOAT64)

	// date family
	T_date     uint8 = uint8(plan.Type_DATE)
	T_datetime uint8 = uint8(plan.Type_DATETIME)

	// string family
	T_char    uint8 = uint8(plan.Type_CHAR)
	T_varchar uint8 = uint8(plan.Type_VARCHAR)

	// json family
	T_json uint8 = uint8(plan.Type_JSON)

	// numeric/decimal family - unsigned attribute is deprecated
	T_decimal64  = uint8(plan.Type_DECIMAL64)
	T_decimal128 = uint8(plan.Type_DECIMAL128)

	// system family
	T_sel   uint8 = uint8(plan.Type_SEL)   //selection
	T_tuple uint8 = uint8(plan.Type_TUPLE) // immutable, size = 24
)

type Element[T any] interface {
	Size() int
}

type Type struct {
	Oid  uint8 `json:"oid,string"`
	Size int32 `json:"size,string"` // e.g. int8.Size = 1, int16.Size = 2, char.Size = 24(SliceHeader size)

	// Width means max Display width for float and double, char and varchar // todo: need to add new attribute DisplayWidth ?
	Width int32 `json:"width,string"`

	Scale int32 `json:"Scale,string"`

	Precision int32 `json:"Precision,string"`
}

type Bool bool
type Int8 int8
type Int16 int16
type Int32 int32
type Int64 int64
type UInt8 uint8
type UInt16 uint16
type UInt32 uint32
type UInt64 uint64
type Float32 float32
type Float64 float64

type Bytes []byte

type Date int32

type Datetime int64

type Decimal64 int64

type Decimal128 struct {
	Lo int64
	Hi int64
}

func New(oid uint8) Type {
	return Type{Oid: oid, Size: int32(TypeSize(oid))}
}

func TypeSize(oid uint8) int {
	switch oid {
	case T_bool, T_int8, T_uint8:
		return 1
	case T_int16, T_uint16:
		return 2
	case T_int32, T_uint32, T_float32, T_date:
		return 4
	case T_int64, T_uint64, T_datetime, T_decimal64:
		return 8
	case T_decimal128:
		return 16
	case T_char, T_varchar:
		return 24
	}
	return -1
}

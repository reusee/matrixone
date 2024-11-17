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

package perfcounter

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/matrixorigin/matrixone/pkg/util/metric/stats"
)

var statsCounterType = reflect.TypeFor[stats.Counter]()

func (c *CounterSet) IterFields() iter.Seq2[[]string, *stats.Counter] {
	return func(yield func([]string, *stats.Counter) bool) {
		iterFields(
			reflect.ValueOf(c),
			[]string{},
			yield,
		)
	}
}

func iterFields(v reflect.Value, path []string, yield func([]string, *stats.Counter) bool) bool {

	if v.Type() == statsCounterType {
		return yield(path, v.Addr().Interface().(*stats.Counter))
	}

	t := v.Type()

	switch t.Kind() {

	case reflect.Pointer:
		return iterFields(v.Elem(), path, yield)

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !iterFields(v.Field(i), append(path, field.Name), yield) {
				return false
			}
		}

	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic(fmt.Sprintf("unknown type: %v", v.Type()))
		}
		iter := v.MapRange()
		for iter.Next() {
			if !iterFields(iter.Value(), append(path, iter.Key().String()), yield) {
				return false
			}
		}

	default:
		panic(fmt.Sprintf("unknown type: %v, %v", path, v.Type()))
	}

	return true
}

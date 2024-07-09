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

package malloc

import (
	"fmt"
	"testing"
)

func TestRef(t *testing.T) {
	holder1 := NewRefHolder[int]()
	holder2 := NewRefHolder[int]()

	t.Run("borrow", func(t *testing.T) {
		ref1 := holder1.Own(1)
		ref2 := ref1.Borrow(holder2)
		ref2.End()
		ref1.End()
	})

	t.Run("double borrow", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "already borrowed" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		_ = ref1.Borrow(holder2)
		_ = ref1.Borrow(holder2)
	})

	t.Run("borrow from borrower", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "cannot borrow" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref2 := ref1.Borrow(holder2)
		ref2.Borrow(holder1)
	})

	t.Run("borrow to owner", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "borrow to owner" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref1.Borrow(holder1)
	})

	t.Run("null ref", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "null Ref" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref1.End()
		ref1.End()
	})

	t.Run("end with borrowing", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "still being borrowed" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		_ = ref1.Borrow(holder2)
		ref1.End()
	})

	t.Run("owner not found", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "owner not found" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref2 := ref1.Borrow(holder2)
		ref3 := ref2 // should not copy Ref
		ref2.End()
		ref3.End()
	})

	t.Run("invalid role in End", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "invalid role" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref2 := ref1.Borrow(holder2)
		ref2.role = 99
		ref2.End()
	})

	t.Run("move", func(t *testing.T) {
		ref1 := holder1.Own(1)
		ref1.Move(holder2)
		ref2 := ref1.Borrow(holder1)
		holder3 := NewRefHolder[int]()
		ref1.Move(holder3)
		ref2.Move(holder2)
	})

	t.Run("bad holder in move", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "not holder" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		holder2.move(&ref1, holder2)
	})

	t.Run("same holder in move", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "same holder" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref1.Move(holder1)
	})

	t.Run("move ownership to borrower", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "cannot move ownership to borrower" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		_ = ref1.Borrow(holder2)
		ref1.Move(holder2)
	})

	t.Run("owner not found in move", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "owner not found" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref2 := ref1.Borrow(holder2)
		ref3 := ref2 // should not copy Ref
		ref2.End()
		holder3 := NewRefHolder[int]()
		ref3.Move(holder3)
	})

	t.Run("invalid role in Move", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("should panic")
			}
			if msg := fmt.Sprintf("%v", p); msg != "invalid role" {
				t.Fatalf("got %v", msg)
			}
		}()
		ref1 := holder1.Own(1)
		ref1.role = 99
		ref1.Move(holder2)
	})

}

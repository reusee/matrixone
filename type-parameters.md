# 范型调研

本文是一个mo使用golang范型的提案，该提案的目的是在不明显损失性能的情况下利用范型达到类型安全和代码简洁的目的，提案的项目地址如下:

```html
https://github.com/nnsgmsone/matrixone/tree/generics
```



## 基本结构

首先给出基本的数据结构，这些基础的结构构成了mo的基础:

types定义了mo所提供的sql类型，所有的类型都被表示为Element。

```golang
type Element[T any] interface {
    Size() int // return the size of space  the Element need 
}  
```

vector定义了mo的基础数据结构vector，一个vector表示具体的一列，AnyVector是所有类型Vector需要实现的类型。

```golang
type Vector[T types.Element[T]] struct {                                                                   
    Col     []T                                                                                           
    Data    []byte   // raw data
    Offsets []uint64 // optional
    Lengths []uint64 // optional                                                                           
    Typ     types.Type                                                                                    
}

type AnyVector interface {
    Reset()
    Length() int
    SetLength(n int)
    Type() types.Type
    Free(*mheap.Mheap)
    Realloc(size int, m *mheap.Mheap) error
}
```

batch定义了mo的基础结构batch，一个batch是mo的最小处理单位，每个batch代表多个vector。

```golang
// Batch represents a part of a relationship
//  (Attrs) - list of attributes                                                                            
//  (vecs)  - columns
type Batch struct {                                                                                         
    // Attrs column name list
    Attrs []string
    // Vecs col data
    Vecs []vector.AnyVector
}
```



## 加法

首先给出性能测试数据:

```
BenchmarkPlus-20           	 2807587	       431.3 ns/op
BenchmarkGenericPlus-20    	 2581725	       460.4 ns/op

BenchmarkPlus-20           	 2879882	       416.4 ns/op
BenchmarkGenericPlus-20    	 2635405	       458.8 ns/op

BenchmarkPlus-20           	 2850846	       419.6 ns/op
BenchmarkGenericPlus-20    	 2619049	       460.5 ns/op
```

测试代码如下:

```golang
const Loop = 1000

var Xs, Ys, Zs []int64
var Vx, Vy, Vz *vector.Vector[types.Int64]

func init() {
    hm := host.New(1 << 20)
    gm := guest.New(1<<20, hm)
    m := mheap.New(gm)
    Xs = make([]int64, Loop)
    Ys = make([]int64, Loop)
    Zs = make([]int64, Loop)
    Vx = vector.New[types.Int64](types.New(types.T_int64))
    Vy = vector.New[types.Int64](types.New(types.T_int64))
    Vz = vector.New[types.Int64](types.New(types.T_int64))
    for i := 0; i < Loop; i++ {
        x := rand.Intn(Loop * 10)
        y := rand.Intn(Loop * 10)
        Xs[i] = int64(x)
        Ys[i] = int64(y)
        Vx.Append(types.Int64(x), m)
        Vy.Append(types.Int64(x), m)
        Vz.Append(types.Int64(0), m)
    }

}

func BenchmarkPlus(b *testing.B) {
    for i := 0; i < b.N; i++ {
        int64s.Plus(Xs, Ys, Zs)
    }
}

func BenchmarkGenericPlus(b *testing.B) {
    for i := 0; i < b.N; i++ {
        VectorPlus(Vx, Vy, Vz)
    }
}
```

范型plus具体实现代码如下:

```golang
func VectorPlus(vx, vy, vz vector.AnyVector) {      
    switch vx.Type().Oid {                                    
    case types.T_int8:                     
        Plus((any)(vx).(*vector.Vector[types.Int8]).Col, (any)(vy).(*vector.Vector[types.Int8]).Col, (any)(vz).(*vector.Vector[types.Int8]).Col)  
    case types.T_int16:                                                                                                                                                              
        Plus((any)(vx).(*vector.Vector[types.Int16]).Col, (any)(vy).(*vector.Vector[types.Int16]).Col, (any)(vz).(*vector.Vector[types.Int16]).Col)                                                                
    case types.T_int32:                                                                                                                                                                   
        Plus((any)(vx).(*vector.Vector[types.Int32]).Col, (any)(vy).(*vector.Vector[types.Int32]).Col, (any)(vz).(*vector.Vector[types.Int32]).Col)                                                                
    case types.T_int64:                                                                                                                                                                   
        Plus((any)(vx).(*vector.Vector[types.Int64]).Col, (any)(vy).(*vector.Vector[types.Int64]).Col, (any)(vz).(*vector.Vector[types.Int64]).Col)                                                                
    }                                                                                                                                                                               
} 

func Plus[T constraints.Integer | constraints.Float](xs, ys, zs []T) []T {
    for i, x := range xs {
        zs[i] = x + ys[i]
    }
    return zs
}
```

非范型的具体实现代码如下: 

```golang
func Plus(xs, ys, zs []int64) []int64 {
    for i, x := range xs {
        zs[i] = x + ys[i]
    }
    return zs
}  
```

此提案为了保证性能，还是不可避免的引入了switch case。



## 排序

首先给出性能测试数据:

```
BenchmarkSort-20           	  231228	      4986 ns/op
BenchmarkGenericSort-20    	  232078	      5019 ns/op

BenchmarkSort-20           	  230278	      4986 ns/op
BenchmarkGenericSort-20    	  237295	      4915 ns/op

BenchmarkSort-20           	  233094	      5044 ns/op
BenchmarkGenericSort-20    	  228297	      5143 ns/op
```

sort不管是范型还是非范型采用的代码都是go1.17的代码，需要注意的是go1.19的sort算法有巨大改进，建议mo用到排序算法的地方都借鉴go1.19。

具体的测试代码如下:

```golang
const Loop = 1000

var Xs []int64
var Vec *vector.Vector[types.Int64]

func init() {
    Xs = make([]int64, Loop)
    Vec = vector.New[types.Int64](types.New(types.T_int64))
    for i := 0; i < Loop; i++ {
        x := rand.Intn(Loop * 10)
        Xs[i] = int64(x)
        Vec.Col = append(Vec.Col, types.Int64(x))
    }
}

func BenchmarkSort(b *testing.B) {
    for i := 0; i < b.N; i++ {
        int64s.Sort(Xs)
    }
}

func BenchmarkGenericSort(b *testing.B) {
    for i := 0; i < b.N; i++ {
        VectorSort(Vec)
    }
}
```

具体的实现因为比较多，感兴趣的可以直接去github查看(pkg/sort)。



## hashtable构建

首先给出性能测试数据:

```
BenchmarkGroup-20           	   59856	     19565 ns/op
BenchmarkGenericGroup-20    	   66456	     17708 ns/op

BenchmarkGroup-20           	   59346	     19719 ns/op
BenchmarkGenericGroup-20    	   69619	     17221 ns/op

BenchmarkGroup-20           	   59346	     19719 ns/op
BenchmarkGenericGroup-20    	   69619	     17221 ns/op
```

具体的测试代码如下:

```golang
const (
    Loop = 10000
)

var Vecs []vector.AnyVector
var Int64Vecs []int64s.Vector

func init() {
    hm := host.New(1 << 20)
    gm := guest.New(1<<20, hm)
    m := mheap.New(gm)
    Vecs = make([]vector.AnyVector, 2)
    Int64Vecs = make([]int64s.Vector, 2)
    {
        vs := make([]int64, Loop)
        vec := vector.New[types.Int64](types.New(types.T_int64))
        for i := 0; i < Loop; i++ {
            vs[i] = int64(i)
            vec.Append(types.Int64(i), m)
        }
        Vecs[0] = vec
        Int64Vecs[0] = int64s.Vector{
            Col: vs,
            Typ: types.New(types.T_int64),
        }
    }
    {
        vs := make([]int64, Loop)
        vec := vector.New[types.Int64](types.New(types.T_int64))
        for i := 0; i < Loop; i++ {
            vs[i] = int64(i)
            vec.Append(types.Int64(i), m)
        }
        Vecs[1] = vec
        Int64Vecs[1] = int64s.Vector{
            Col: vs,
            Typ: types.New(types.T_int64),
        }

    }
}

func BenchmarkGroup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        int64s.Group(Int64Vecs, Loop)
    }
}

func BenchmarkGenericGroup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Group(Vecs, Loop)
    }
}
```

相比不用范型的填充代码会减少不少重复的代码，当然因为存在多种类型的hashtable，所以每类hashtable写一遍代码还是避免不了的。


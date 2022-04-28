package sort

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"golang.org/x/exp/constraints"
)

func VectorSort(vec vector.AnyVector) {
	switch vec.Type().Oid {
	case types.T_int8:
		Sort((any)(vec).(*vector.Vector[types.Int8]).Col)
	case types.T_int16:
		Sort((any)(vec).(*vector.Vector[types.Int16]).Col)
	case types.T_int32:
		Sort((any)(vec).(*vector.Vector[types.Int32]).Col)
	case types.T_int64:
		Sort((any)(vec).(*vector.Vector[types.Int64]).Col)
	}
}

// Sort sorts data.
// It makes one call to data.Len to determine n, and Operator(n*log(n)) calls to
// data.Less and data.Swap. The sort is not guaranteed to be stable.
func Sort[T constraints.Ordered](vs []T) {
	n := len(vs)
	quickSort(vs, 0, n, maxDepth(n))
}

// maxDepth returns a threshold at which quicksort should switch
// to heapsort. It returns 2*ceil(lg(n+1)).
func maxDepth(n int) int {
	var depth int
	for i := n; i > 0; i >>= 1 {
		depth++
	}
	return depth * 2
}

func quickSort[T constraints.Ordered](vs []T, a, b, maxDepth int) {
	for b-a > 12 { // Use ShellSort for slices <= 12 elements
		if maxDepth == 0 {
			heapSort(vs, a, b)
			return
		}
		maxDepth--
		mlo, mhi := doPivot(vs, a, b)
		// Avoiding recursion on the larger subproblem guarantees
		// a stack depth of at most lg(b-a).
		if mlo-a < b-mhi {
			quickSort(vs, a, mlo, maxDepth)
			a = mhi // i.e., quickSort(data, mhi, b)
		} else {
			quickSort(vs, mhi, b, maxDepth)
			b = mlo // i.e., quickSort(data, a, mlo)
		}
	}
	if b-a > 1 {
		// Do ShellSort pass with gap 6
		// It could be written in this simplified form cause b-a <= 12
		for i := a + 6; i < b; i++ {
			if vs[i] < vs[i-6] {
				vs[i], vs[i-6] = vs[i-6], vs[i]
			}
		}
		insertionSort(vs, a, b)
	}
}

// Insertion sort
func insertionSort[T constraints.Ordered](vs []T, a, b int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && vs[j] < vs[j-1]; j-- {
			vs[j], vs[j-1] = vs[j-1], vs[j]
		}
	}
}

// siftDown implements the heap property on data[lo, hi).
// first is an offset into the array where the root of the heap lies.
func siftDown[T constraints.Ordered](vs []T, lo, hi, first int) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}
		if child+1 < hi && vs[first+child] < vs[first+child+1] {
			child++
		}
		if vs[first+root] >= vs[first+child] {
			return
		}
		vs[first+root], vs[first+child] = vs[first+child], vs[first+root]
		root = child
	}
}

func heapSort[T constraints.Ordered](vs []T, a, b int) {
	first := a
	lo := 0
	hi := b - a

	// Build heap with greatest element at top.
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown(vs, i, hi, first)
	}

	// Pop elements, largest first, into end of data.
	for i := hi - 1; i >= 0; i-- {
		vs[first], vs[first+i] = vs[first+i], vs[first]
		siftDown(vs, lo, i, first)
	}
}

// Quicksort, loosely following Bentley and McIlroy,
// ``Engineering a Sort Function,'' SP&E November 1993.

// medianOfThree moves the median of the three values data[m0], data[m1], data[m2] into data[m1].
func medianOfThree[T constraints.Ordered](vs []T, m1, m0, m2 int) {
	// sort 3 elements
	if vs[m1] < vs[m0] {
		vs[m1], vs[m0] = vs[m0], vs[m1]
	}
	// data[m0] <= data[m1]
	if vs[m2] < vs[m1] {
		vs[m2], vs[m1] = vs[m1], vs[m2]
		// data[m0] <= data[m2] && data[m1] < data[m2]
		if vs[m1] < vs[m0] {
			vs[m1], vs[m0] = vs[m0], vs[m1]
		}
	}
	// now data[m0] <= data[m1] <= data[m2]
}

func doPivot[T constraints.Ordered](vs []T, lo, hi int) (midlo, midhi int) {
	m := int(uint(lo+hi) >> 1) // Written like this to avoid integer overflow.
	if hi-lo > 40 {
		// Tukey's ``Ninther,'' median of three medians of three.
		s := (hi - lo) / 8
		medianOfThree(vs, lo, lo+s, lo+2*s)
		medianOfThree(vs, m, m-s, m+s)
		medianOfThree(vs, hi-1, hi-1-s, hi-1-2*s)
	}
	medianOfThree(vs, lo, m, hi-1)

	// Invariants are:
	//	data[lo] = pivot (set up by ChoosePivot)
	//	data[lo < i < a] < pivot
	//	data[a <= i < b] <= pivot
	//	data[b <= i < c] unexamined
	//	data[c <= i < hi-1] > pivot
	//	data[hi-1] >= pivot
	pivot := lo
	a, c := lo+1, hi-1

	for ; a < c && vs[a] < vs[pivot]; a++ {
	}
	b := a
	for {
		for ; b < c && vs[pivot] >= vs[b]; b++ { // data[b] <= pivot
		}
		for ; b < c && vs[pivot] < vs[c-1]; c-- { // data[c-1] > pivot
		}
		if b >= c {
			break
		}
		// data[b] > pivot; data[c-1] <= pivot
		vs[b], vs[c-1] = vs[c-1], vs[b]
		b++
		c--
	}
	// If hi-c<3 then there are duplicates (by property of median of nine).
	// Let's be a bit more conservative, and set border to 5.
	protect := hi-c < 5
	if !protect && hi-c < (hi-lo)/4 {
		// Lets test some points for equality to pivot
		dups := 0
		if vs[pivot] >= vs[hi-1] { // data[hi-1] = pivot
			vs[c], vs[hi-1] = vs[hi-1], vs[c]
			c++
			dups++
		}
		if vs[b-1] >= vs[pivot] { // data[b-1] = pivot
			b--
			dups++
		}
		// m-lo = (hi-lo)/2 > 6
		// b-lo > (hi-lo)*3/4-1 > 8
		// ==> m < b ==> data[m] <= pivot
		if vs[m] >= vs[pivot] { // data[m] = pivot
			vs[m], vs[b-1] = vs[b-1], vs[m]
			b--
			dups++
		}
		// if at least 2 points are equal to pivot, assume skewed distribution
		protect = dups > 1
	}
	if protect {
		// Protect against a lot of duplicates
		// Add invariant:
		//	data[a <= i < b] unexamined
		//	data[b <= i < c] = pivot
		for {
			for ; a < b && vs[b-1] >= vs[pivot]; b-- { // data[b] == pivot
			}
			for ; a < b && vs[a] < vs[pivot]; a++ { // data[a] < pivot
			}
			if a >= b {
				break
			}
			// data[a] == pivot; data[b-1] < pivot
			vs[a], vs[b-1] = vs[b-1], vs[a]
			a++
			b--
		}
	}
	// Swap pivot into middle
	vs[pivot], vs[b-1] = vs[b-1], vs[pivot]
	return b - 1, c
}

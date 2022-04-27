package int64s

func Plus(xs, ys, zs []int64) []int64 {
	for i, x := range xs {
		zs[i] = x + ys[i]
	}
	return zs
}

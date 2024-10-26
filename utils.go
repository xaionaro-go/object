package object

import "cmp"

func max[T cmp.Ordered](a0 T, an ...T) T {
	max := a0
	for _, v := range an {
		if v > max {
			max = v
		}
	}
	return max
}

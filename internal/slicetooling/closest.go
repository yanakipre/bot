package slicetooling

import "github.com/samber/lo"

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Closest returns the number in the numbers slice that is closest to the given number.
// If numbers slice is empty, it returns the number.
// XXX: This function is not optimized for large sorted slices.
func Closest(numbers []int, num int) int {
	if len(numbers) == 0 {
		return num
	}
	current := numbers[0]
	lo.ForEach(numbers, func(n int, _ int) {
		if abs(n-num) < abs(current-num) {
			current = n
		}
	})
	return current
}

package slicetooling

// IntersectTakeRight returns a slice of elements that are present in both lists.
// The idFunc is used to compare elements.
// Returned elements are taken from the second list.
func IntersectTakeRight[T any, C comparable](list1 []T, list2 []T, idFunc func(T) C) []T {
	result := []T{}
	seen := map[C]bool{}

	for _, elem := range list1 {
		seen[idFunc(elem)] = true
	}

	for _, elem := range list2 {
		if seen[idFunc(elem)] {
			result = append(result, elem)
		}
	}

	return result
}
